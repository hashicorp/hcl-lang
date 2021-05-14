package decoder

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/mitchellh/copystructure"
)

type Decoder struct {
	files   map[string]*hcl.File
	filesMu *sync.RWMutex

	refReader     ReferenceReader
	rootSchema    *schema.BodySchema
	rootSchemaMu  *sync.RWMutex
	maxCandidates uint

	// UTM parameters for docs URLs
	// utm_source parameter, typically language server identification
	utmSource string
	// utm_medium parameter, typically language client identification
	utmMedium string
	// utm_content parameter, e.g. documentHover or documentLink
	useUtmContent bool
}

type ReferenceReader func() lang.References

// NewDecoder creates a new Decoder
//
// Decoder is safe for use without any schema, but configuration files are loaded
// via LoadFile and (optionally) schema is set via SetSchema.
func NewDecoder() *Decoder {
	return &Decoder{
		rootSchemaMu:  &sync.RWMutex{},
		files:         make(map[string]*hcl.File, 0),
		filesMu:       &sync.RWMutex{},
		maxCandidates: 100,
	}
}

// SetSchema sets the schema decoder uses for decoding the configuration
//
// This is useful for progressive enhancement experience, where a
// Decoder without schema can provide limited functionality (e.g. symbols), and
// the schema can be gradually enriched (e.g. Terraform core -> providers).
func (d *Decoder) SetSchema(schema *schema.BodySchema) {
	d.rootSchemaMu.Lock()
	defer d.rootSchemaMu.Unlock()
	d.rootSchema = schema
}

func (d *Decoder) SetReferenceReader(f ReferenceReader) {
	d.refReader = f
}

func (d *Decoder) SetUtmSource(src string) {
	d.utmSource = src
}

func (d *Decoder) SetUtmMedium(medium string) {
	d.utmMedium = medium
}

func (d *Decoder) UseUtmContent(use bool) {
	d.useUtmContent = use
}

// LoadFile loads a new (non-empty) parsed file
//
// e.g. result of hclsyntax.ParseConfig
func (d *Decoder) LoadFile(filename string, f *hcl.File) error {
	d.filesMu.Lock()
	defer d.filesMu.Unlock()

	if f == nil {
		return fmt.Errorf("%s: invalid content provided", filename)
	}

	if f.Body == nil {
		return fmt.Errorf("%s: file has no body", filename)
	}

	d.files[filename] = f

	return nil
}

// Filenames returns a slice of filenames already loaded via LoadFile
func (p *Decoder) Filenames() []string {
	p.filesMu.RLock()
	defer p.filesMu.RUnlock()

	var files []string
	for filename := range p.files {
		files = append(files, filename)
	}

	sort.Strings(files)

	return files
}

func (d *Decoder) bytesForFile(file string) ([]byte, error) {
	d.filesMu.RLock()
	defer d.filesMu.RUnlock()

	f, ok := d.files[file]
	if !ok {
		return nil, &FileNotFoundError{Filename: file}
	}

	return f.Bytes, nil
}

func (d *Decoder) bytesFromRange(rng hcl.Range) ([]byte, error) {
	b, err := d.bytesForFile(rng.Filename)
	if err != nil {
		return nil, err
	}

	return rng.SliceBytes(b), nil
}

func (d *Decoder) fileByName(name string) (*hcl.File, error) {
	d.filesMu.RLock()
	defer d.filesMu.RUnlock()

	f, ok := d.files[name]
	if !ok {
		return nil, &FileNotFoundError{Filename: name}
	}
	return f, nil
}

func (d *Decoder) bodyForFileAndPos(name string, f *hcl.File, pos hcl.Pos) (*hclsyntax.Body, error) {
	body, isHcl := f.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, &UnknownFileFormatError{Filename: name}
	}

	if !body.Range().ContainsPos(pos) &&
		!posEqual(body.Range().Start, pos) &&
		!posEqual(body.Range().End, pos) {

		return nil, &PosOutOfRangeError{
			Filename: name,
			Pos:      pos,
			Range:    body.Range(),
		}
	}

	return body, nil
}

func posEqual(pos, other hcl.Pos) bool {
	return pos.Line == other.Line &&
		pos.Column == other.Column &&
		pos.Byte == other.Byte
}

func mergeBlockBodySchemas(block *hclsyntax.Block, blockSchema *schema.BlockSchema) (*schema.BodySchema, error) {
	if len(blockSchema.DependentBody) == 0 {
		return blockSchema.Body, nil
	}

	schemaCopy, err := copystructure.Config{
		Copiers: copiers,
	}.Copy(blockSchema.Body)
	if err != nil {
		return nil, err
	}

	mergedSchema := schemaCopy.(*schema.BodySchema)
	if mergedSchema.Attributes == nil {
		mergedSchema.Attributes = make(map[string]*schema.AttributeSchema, 0)
	}
	if mergedSchema.Blocks == nil {
		mergedSchema.Blocks = make(map[string]*schema.BlockSchema, 0)
	}

	dk := dependencyKeysFromBlock(block, blockSchema)

	depSchema, ok := blockSchema.DependentBodySchema(dk)
	if ok {
		for name, attr := range depSchema.Attributes {
			if _, exists := mergedSchema.Attributes[name]; !exists {
				mergedSchema.Attributes[name] = attr
			} else {
				// Skip duplicate attribute
				continue
			}
		}
		for bType, block := range depSchema.Blocks {
			if _, exists := mergedSchema.Blocks[bType]; !exists {
				mergedSchema.Blocks[bType] = block
			} else {
				// Skip duplicate block type
				continue
			}
		}
	}

	return mergedSchema, nil
}

func dependencyKeysFromBlock(block *hclsyntax.Block, blockSchema *schema.BlockSchema) schema.DependencyKeys {
	dk := schema.DependencyKeys{
		Labels:     []schema.LabelDependent{},
		Attributes: []schema.AttributeDependent{},
	}
	for i, labelSchema := range blockSchema.Labels {
		if labelSchema.IsDepKey {
			dk.Labels = append(dk.Labels, schema.LabelDependent{
				Index: i,
				Value: block.Labels[i],
			})
		}
	}

	if blockSchema.Body == nil {
		return dk
	}

	for name, attrSchema := range blockSchema.Body.Attributes {
		if attrSchema.IsDepKey {
			attr, ok := block.Body.Attributes[name]
			if !ok {
				// dependent attribute not present
				continue
			}

			st, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr)
			if ok {
				addr, err := traversalToAddress(st.AsTraversal())
				if err != nil {
					// skip unparsable traversal
					continue
				}
				dk.Attributes = append(dk.Attributes, schema.AttributeDependent{
					Name: name,
					Expr: schema.ExpressionValue{
						Address: addr,
					},
				})
				continue
			}

			value, diags := attr.Expr.Value(nil)
			if len(diags) > 0 && value.IsNull() {
				// skip attribute if we can't get the value
				continue
			}

			dk.Attributes = append(dk.Attributes, schema.AttributeDependent{
				Name: name,
				Expr: schema.ExpressionValue{
					Static: value,
				},
			})
		}
	}
	return dk
}

func traversalToAddress(traversal hcl.Traversal) (lang.Address, error) {
	addr := lang.Address{}
	for _, tr := range traversal {
		switch t := tr.(type) {
		case hcl.TraverseRoot:
			addr = append(addr, lang.RootStep{
				Name: t.Name,
			})
		case hcl.TraverseAttr:
			addr = append(addr, lang.AttrStep{
				Name: t.Name,
			})
		case hcl.TraverseIndex:
			addr = append(addr, lang.IndexStep{
				Key: t.Key,
			})
		default:
			return addr, fmt.Errorf("invalid traversal: %#v", tr)
		}
	}
	return addr, nil
}

func stringPos(pos hcl.Pos) string {
	return fmt.Sprintf("%d,%d", pos.Line, pos.Column)
}
