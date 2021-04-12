package decoder

import (
	"reflect"

	"github.com/mitchellh/copystructure"
	"github.com/zclconf/go-cty/cty"
)

// Some types have private fields, so we declare custom copiers for them
var copiers = map[reflect.Type]copystructure.CopierFunc{
	reflect.TypeOf(cty.NilType): ctyTypeCopier,
	reflect.TypeOf(cty.Value{}): ctyValueCopier,
}

func ctyTypeCopier(v interface{}) (interface{}, error) {
	return v.(cty.Type), nil
}

func ctyValueCopier(v interface{}) (interface{}, error) {
	return v.(cty.Value), nil
}
