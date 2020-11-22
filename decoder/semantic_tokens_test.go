package decoder

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_SemanticTokensInFile_emptyBody(t *testing.T) {
	d := NewDecoder()
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.SemanticTokensInFile("test.tf")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_SemanticTokensInFile_zeroByteContent(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := d.SemanticTokensInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}
	expectedTokens := []lang.SemanticToken{}
	if diff := cmp.Diff(expectedTokens, tokens); diff != "" {
		t.Fatalf("unexpected tokens: %s", diff)
	}
}

func TestDecoder_SemanticTokensInFile_fileNotFound(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.SemanticTokensInFile("foobar.tf")
	notFoundErr := &FileNotFoundError{}
	if !errors.As(err, &notFoundErr) {
		t.Fatal("expected FileNotFoundError for non-existent file")
	}
}

func TestDecoder_SemanticTokensInFile_basic(t *testing.T) {
	d := NewDecoder()
	d.SetSchema(&schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"module": {
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"count": {
							ValueType: cty.Number,
						},
						"source": {
							ValueType:    cty.String,
							IsDeprecated: true,
						},
					},
				},
			},
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type", IsDepKey: true},
					{Name: "name"},
				},
			},
		},
	})

	testCfg := []byte(`module "ref" {
  source = "./sub"
  count  = 1
}
resource "vault_auth_backend" "blah" {
  default_lease_ttl_seconds = 1
}
`)

	f, pDiags := hclsyntax.ParseConfig(testCfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := d.SemanticTokensInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedTokens := []lang.SemanticToken{
		{ // module
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   0,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 7,
					Byte:   6,
				},
			},
		},
		{ // source
			Type: lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDeprecated,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   2,
					Column: 3,
					Byte:   17,
				},
				End: hcl.Pos{
					Line:   2,
					Column: 9,
					Byte:   23,
				},
			},
		},
		{ // count
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   3,
					Column: 3,
					Byte:   36,
				},
				End: hcl.Pos{
					Line:   3,
					Column: 8,
					Byte:   41,
				},
			},
		},
		{ // resource
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 1,
					Byte:   49,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 9,
					Byte:   57,
				},
			},
		},
		{ // vault_auth_backend
			Type: lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDependent,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 10,
					Byte:   58,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 30,
					Byte:   78,
				},
			},
		},
		{ // blah
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 31,
					Byte:   79,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 37,
					Byte:   85,
				},
			},
		},
	}

	diff := cmp.Diff(expectedTokens, tokens)
	if diff != "" {
		t.Fatalf("unexpected tokens: %s", diff)
	}
}

func TestDecoder_SemanticTokensInFile_dependentSchema(t *testing.T) {
	d := NewDecoder()
	d.SetSchema(&schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type", IsDepKey: true},
					{Name: "name"},
				},
				DependentBody: map[schema.SchemaKey]*schema.BodySchema{
					schema.NewSchemaKey(schema.DependencyKeys{
						Labels: []schema.LabelDependent{
							{
								Index: 0,
								Value: "aws_instance",
							},
						},
					}): {
						Attributes: map[string]*schema.AttributeSchema{
							"instance_type": {
								ValueType: cty.String,
							},
							"deprecated": {
								ValueType: cty.Bool,
							},
						},
					},
				},
			},
		},
	})

	testCfg := []byte(`resource "vault_auth_backend" "alpha" {
  default_lease_ttl_seconds = 1
}
resource "aws_instance" "beta" {
  instance_type = "t2.micro"
  deprecated = true
}
`)

	f, pDiags := hclsyntax.ParseConfig(testCfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := d.SemanticTokensInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedTokens := []lang.SemanticToken{
		{
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   0,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 9,
					Byte:   8,
				},
			},
		},
		{
			Type: lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDependent,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 10,
					Byte:   9,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 30,
					Byte:   29,
				},
			},
		},
		{
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 31,
					Byte:   30,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 38,
					Byte:   37,
				},
			},
		},
		{
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   4,
					Column: 1,
					Byte:   74,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 9,
					Byte:   82,
				},
			},
		},
		{
			Type: lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDependent,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   4,
					Column: 10,
					Byte:   83,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 24,
					Byte:   97,
				},
			},
		},
		{
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   4,
					Column: 25,
					Byte:   98,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 31,
					Byte:   104,
				},
			},
		},
		{
			Type: lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDependent,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 3,
					Byte:   109,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 16,
					Byte:   122,
				},
			},
		},
		{
			Type: lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{
				lang.TokenModifierDependent,
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   6,
					Column: 3,
					Byte:   138,
				},
				End: hcl.Pos{
					Line:   6,
					Column: 13,
					Byte:   148,
				},
			},
		},
	}

	diff := cmp.Diff(expectedTokens, tokens)
	if diff != "" {
		t.Fatalf("unexpected tokens: %s", diff)
	}
}
