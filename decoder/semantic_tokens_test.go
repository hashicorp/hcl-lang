package decoder

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_SemanticTokensInFile_emptyBody(t *testing.T) {
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}
	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.SemanticTokensInFile("test.tf")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_SemanticTokensInFile_json(t *testing.T) {
	f, pDiags := json.Parse([]byte(`{
	"customblock": {
		"label1": {}
	}
}`), "test.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf.json": f,
		},
	})

	_, err := d.SemanticTokensInFile("test.tf.json")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}

func TestDecoder_SemanticTokensInFile_zeroByteContent(t *testing.T) {
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

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
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.SemanticTokensInFile("foobar.tf")
	notFoundErr := &FileNotFoundError{}
	if !errors.As(err, &notFoundErr) {
		t.Fatal("expected FileNotFoundError for non-existent file")
	}
}

func TestDecoder_SemanticTokensInFile_basic(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"module": {
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"count": {
							Expr: schema.LiteralTypeOnly(cty.Number),
						},
						"source": {
							Expr:         schema.LiteralTypeOnly(cty.String),
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
	}

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

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

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
		{ // "./sub"
			Type:      lang.TokenString,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   2,
					Column: 12,
					Byte:   26,
				},
				End: hcl.Pos{
					Line:   2,
					Column: 19,
					Byte:   33,
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
		{ // 1
			Type:      lang.TokenNumber,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   3,
					Column: 12,
					Byte:   45,
				},
				End: hcl.Pos{
					Line:   3,
					Column: 13,
					Byte:   46,
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
	bodySchema := &schema.BodySchema{
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
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"deprecated": {
								Expr: schema.LiteralTypeOnly(cty.Bool),
							},
						},
					},
				},
			},
		},
	}

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

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	tokens, err := d.SemanticTokensInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedTokens := []lang.SemanticToken{
		{ // resource
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
		{ // "vault_auth_backend"
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
		{ // "alpha"
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
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
		{ // resource
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
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
		{ // "aws_instance"
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
		{ // "beta"
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
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
		{ // instance_type
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
		{ // "t2.micro"
			Type:      lang.TokenString,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 19,
					Byte:   125,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 29,
					Byte:   135,
				},
			},
		},
		{ // deprecated
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
		{ // true
			Type:      lang.TokenBool,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   6,
					Column: 16,
					Byte:   151,
				},
				End: hcl.Pos{
					Line:   6,
					Column: 20,
					Byte:   155,
				},
			},
		},
	}

	diff := cmp.Diff(expectedTokens, tokens)
	if diff != "" {
		t.Fatalf("unexpected tokens: %s", diff)
	}
}

func TestDecoder_SemanticTokensInFile_typeDeclaration(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"variable": {
				Labels: []*schema.LabelSchema{
					{Name: "name", IsDepKey: true},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"type": {
							Expr: schema.ExprConstraints{schema.TypeDeclarationExpr{}},
						},
					},
				},
			},
		},
	}

	testCfg := []byte(`variable "meh" {
	type = string 
}
variable "bah" {
	type = map(string) 
}
variable "bah" {
	type = object({
		model = string
		latest = bool
	}) 
}
`)

	f, pDiags := hclsyntax.ParseConfig(testCfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

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
					Column: 9,
					Byte:   8,
				},
			},
		},
		{
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{lang.TokenModifierDependent},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 10,
					Byte:   9,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 15,
					Byte:   14,
				},
			},
		},
		{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   2,
					Column: 2,
					Byte:   18,
				},
				End: hcl.Pos{
					Line:   2,
					Column: 6,
					Byte:   22,
				},
			},
		},
		{
			Type:      lang.TokenTypePrimitive,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   2,
					Column: 9,
					Byte:   25,
				},
				End: hcl.Pos{
					Line:   2,
					Column: 15,
					Byte:   31,
				},
			},
		},
		{
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   4,
					Column: 1,
					Byte:   35,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 9,
					Byte:   43,
				},
			},
		},
		{
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{lang.TokenModifierDependent},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   4,
					Column: 10,
					Byte:   44,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 15,
					Byte:   49,
				},
			},
		},
		{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 2,
					Byte:   53,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 6,
					Byte:   57,
				},
			},
		},
		{
			Type:      lang.TokenTypeCapsule,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 9,
					Byte:   60,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 12,
					Byte:   63,
				},
			},
		},
		{
			Type:      lang.TokenTypePrimitive,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   5,
					Column: 13,
					Byte:   64,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 19,
					Byte:   70,
				},
			},
		},

		////
		{
			Type:      lang.TokenBlockType,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   7,
					Column: 1,
					Byte:   75,
				},
				End: hcl.Pos{
					Line:   7,
					Column: 9,
					Byte:   83,
				},
			},
		},
		{
			Type:      lang.TokenBlockLabel,
			Modifiers: []lang.SemanticTokenModifier{lang.TokenModifierDependent},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   7,
					Column: 10,
					Byte:   84,
				},
				End: hcl.Pos{
					Line:   7,
					Column: 15,
					Byte:   89,
				},
			},
		},
		{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   8,
					Column: 2,
					Byte:   93,
				},
				End: hcl.Pos{
					Line:   8,
					Column: 6,
					Byte:   97,
				},
			},
		},
		{
			Type:      lang.TokenTypeCapsule,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   8,
					Column: 9,
					Byte:   100,
				},
				End: hcl.Pos{
					Line:   8,
					Column: 15,
					Byte:   106,
				},
			},
		},
		{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   9,
					Column: 3,
					Byte:   111,
				},
				End: hcl.Pos{
					Line:   9,
					Column: 8,
					Byte:   116,
				},
			},
		},
		{
			Type:      lang.TokenTypePrimitive,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   9,
					Column: 11,
					Byte:   119,
				},
				End: hcl.Pos{
					Line:   9,
					Column: 17,
					Byte:   125,
				},
			},
		},
		{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   10,
					Column: 3,
					Byte:   128,
				},
				End: hcl.Pos{
					Line:   10,
					Column: 9,
					Byte:   134,
				},
			},
		},
		{
			Type:      lang.TokenTypePrimitive,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   10,
					Column: 12,
					Byte:   137,
				},
				End: hcl.Pos{
					Line:   10,
					Column: 16,
					Byte:   141,
				},
			},
		},
	}

	diff := cmp.Diff(expectedTokens, tokens)
	if diff != "" {
		t.Fatalf("unexpected tokens: %s", diff)
	}
}
