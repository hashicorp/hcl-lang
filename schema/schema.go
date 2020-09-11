package schema

type schemaImplSigil struct{}

// Schema represents any schema (e.g. attribute, label, or a block)
type Schema interface {
	isSchemaImpl() schemaImplSigil
}
