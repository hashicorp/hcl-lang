# hcl-lang [![Go Reference](https://pkg.go.dev/badge/github.com/hashicorp/hcl-lang.svg)](https://pkg.go.dev/github.com/hashicorp/hcl-lang)

This library provides basic building blocks for an HCL2-based
language server in the form of a schema and a decoder.

## Current Status

This project is in use by the Terraform Language Server and it is designed
to be used by any HCL2 language server, but it's still in early stage
of development.

For that reason the API is **not considered stable yet** and should not be relied upon.

**Breaking changes may be introduced.**

## What is HCL?

See https://github.com/hashicorp/hcl

## What is a Language Server?

See https://microsoft.github.io/language-server-protocol/

## What is Schema?

Schema plays an important role in most HCL2 deployments as it describes
what to expect in the configuration (e.g. attribute or block names, their types etc.),
which in turn provides predictability, static early validation and more.

### Other Schemas

There are other known HCL schema implementations, e.g.:

 - [hcl.BlockHeaderSchema](https://pkg.go.dev/github.com/hashicorp/hcl/v2#BlockHeaderSchema) + [hcl.AttributeSchema](https://pkg.go.dev/github.com/hashicorp/hcl/v2#AttributeSchema)
 - [hcldec.Spec](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hcldec#Spec) (used by [Packer 1.5+](https://pkg.go.dev/github.com/hashicorp/packer@v1.6.2/hcl2template#Decodable))
 - [Terraform Plugin SDK's `helper/schema`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema) (used by all Terraform plugins)
 - [Terraform's `configschema`](https://pkg.go.dev/github.com/hashicorp/terraform@v0.13.5/configs/configschema) (basically core's "`0.12+` understanding of `helper/schema`")
 - [Nomad's `hclspec.Spec`](https://pkg.go.dev/github.com/hashicorp/nomad/plugins/shared/hclspec#Spec) used by all Nomad plugins
 - [`gohcl`](https://pkg.go.dev/github.com/hashicorp/hcl/v2/gohcl) / struct field tags (`v1` effectively used by [Vault](https://pkg.go.dev/github.com/hashicorp/vault@v1.5.3/internalshared/configutil#SharedConfig) and [Consul](https://pkg.go.dev/github.com/hashicorp/consul@v1.8.3/acl#Policy) and `v2` by [Waypoint](https://pkg.go.dev/github.com/hashicorp/waypoint@v0.1.4/internal/config#App) and Nomad 1.0+)

These each have slightly different valid reasons to exist and they will likely
continue to exist - i.e. no schema on that list (nor the one contained in this library)
is meant to replace another, or at least it wasn't designed with that intention.

However in the interest of compatibility and adoption it's expected that
some conversion mechanisms from/to the above schemas will emerge.

## Schema

The `schema` package provides a way of describing schema for an HCL2 language.

For example (simplified Terraform `provider` block):

```go
import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/zclconf/go-cty/cty"
)

var providerBlockSchema = &schema.BlockSchema{
	Labels: []*schema.LabelSchema{
		{
			Name:        "name",
			Description: lang.PlainText("Provider Name"),
			IsDepKey:    true,
		},
	},
	Description: lang.PlainText("A provider block is used to specify a provider configuration. The body of the block (between " +
		"{ and }) contains configuration arguments for the provider itself. Most arguments in this section are " +
		"specified by the provider itself."),
	Body: &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"alias": {
				ValueType:   cty.String,
				Description: lang.PlainText("Alias for using the same provider with different configurations for different resources"),
			},
		},
	},
}
```

### Dependent Body Schema

In most known complex HCL2 deployments (e.g. in Terraform, Nomad, Waypoint),
schemas of some block bodies are defined _partially_ by its type.

e.g. `resource` in Terraform in itself brings `count` attribute.

```hcl
resource "..." "..." {
  count = 2
  
}
```

Other attributes or blocks are then defined by the block's _labels_ or _attributes_.

e.g. 1st label + (optional) `provider` attribute in Terraform's `resource`
brings all other attributes (such as `ami` or `instance_type`).

```hcl
resource "aws_instance" "ref_name" {
  provider = aws.west

  # dependent attributes
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
}
```

Such schema can be represented using the `DependentBody` field
of `BlockSchema`, for example:

```go
var resourceBlockSchema = &schema.BlockSchema{
	Labels: []*schema.LabelSchema{
		{
			Name:        "type",
			Description: lang.PlainText("Resource type"),
		},
		{Name: "name"},
	},
	Body: &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"count":      {ValueType: cty.Number},
			// ...
		},
	},
	DependentBody: map[schema.SchemaKey]*schema.BodySchema{
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "aws_instance"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"ami":           {ValueType: cty.String},
				"instance_type": {ValueType: cty.String},
				// ...
			},
		},
	},
}
```

#### Nested `DependentBody`

It is discouraged from declaring `DependentBody` as part of another ("parent")
`DependentBody` due to complexity (reduced readability of code).

This complex scenario is however supported and is used e.g. in Terraform
for `terraform_remote_state` `data` block, where `config` attribute
is dependent on `backend` value, which itself is dependent on the value
of the 1st block label.

```hcl
data "terraform_remote_state" "name" {
  backend = "local"

  config = {
    workspace_dir = "value"
  }
}
```

Such nested bodies have to declare full schema key, including labels.

#### Populating Dependent Body Schemas

It is expected for `DependentBody` to be populated
based on the needs and capabilities of a particular tool.

For example in Terraform, dependent schemas come from providers (plugins)
and these can be obtained via `terraform providers schema -json` (Terraform CLI 0.12+).
_In the future_ these may also be made available in the [Terraform Registry](https://registry.terraform.io),
or from the provider binaries (via gRPC protocol).

`hcl-lang` does _not_ care _how_ any part of the schema is obtained or _where from_.

It expects `SetSchema` to be called either with full schema
(including fully populated `DependentBody`), or `SetSchema` to be called
repeatedly as more schema is known. The functionality will adapt to the amount
of schema provided (e.g. label completion isn't available without `DependentBody`).

This means that the same configuration may need to be parsed and some minimal
form of schema used for the first time, before the _full_ schema is assembled
and passed to `hcl-lang`'s decoder for the second decoding stage.

[`terraform-config-inspect`](https://github.com/hashicorp/terraform-config-inspect) and
[`terraform-schema`](https://github.com/hashicorp/terraform-schema)
represent examples of how this is done in Terraform.

## Decoder

The `decoder` package provides a decoder which can be utilized by a language server.

```go
d, err := NewDecoder()
if err != nil {
	// ...
}
d.SetSchema(schema)

// for each (known) file (e.g. any *.tf file in Terraform)
f, pDiags := hclsyntax.ParseConfig(configBytes, "example.tf", hcl.InitialPos)
if len(pDiags) > 0 {
	// ...
}
err = d.LoadFile("example.tf", f)
if err != nil {
	// ...
}
```

See available methods in [the documentation](https://pkg.go.dev/github.com/hashicorp/hcl-lang/decoder#Decoder).

## Experimental Status

By using the software in this repository (the "Software"), you acknowledge that: (1) the Software is still in development, may change, and has not been released as a commercial product by HashiCorp and is not currently supported in any way by HashiCorp; (2) the Software is provided on an "as-is" basis, and may include bugs, errors, or other issues; (3) the Software is NOT INTENDED FOR PRODUCTION USE, use of the Software may result in unexpected results, loss of data, or other unexpected results, and HashiCorp disclaims any and all liability resulting from use of the Software; and (4) HashiCorp reserves all rights to make all decisions about the features, functionality and commercial release (or non-release) of the Software, at any time and without any obligation or liability whatsoever.
