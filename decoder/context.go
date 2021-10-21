package decoder

// DecoderContext represents global context relevant for all possible paths
// served by the Decoder
type DecoderContext struct {
	// UTM parameters for docs URLs
	// utm_source parameter, typically language server identification
	UtmSource string
	// utm_medium parameter, typically language client identification
	UtmMedium string
	// utm_content parameter, e.g. documentHover or documentLink
	UseUtmContent bool
}

func (d *Decoder) SetContext(ctx DecoderContext) {
	d.ctx = ctx
}
