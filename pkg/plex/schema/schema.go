package schema

import (
	"fmt"

	"github.com/gobuffalo/packr"
	"github.com/xeipuuv/gojsonschema"
)

type Validator struct {
	sLoader gojsonschema.JSONLoader
}

func NewValidator() (*Validator, error) {
	p := packr.NewBox("./defs")
	schema, err := p.FindString("webhook-payload-schema.json")
	if err != nil {
		return nil, err
	}
	v := Validator{
		sLoader: gojsonschema.NewStringLoader(schema),
	}
	return &v, nil
}

func (v *Validator) Validate(bytes []byte) error {
	doc := gojsonschema.NewStringLoader(string(bytes))
	result, err := gojsonschema.Validate(v.sLoader, doc)
	if err != nil {
		return err
	}

	if !result.Valid() {
		msg := ""
		for _, e := range result.Errors() {
			if len(msg) != 0 {
				msg += ", "
			}
			msg += fmt.Sprintf("%s", e)
		}
		return fmt.Errorf("Invalid payload, see errors: %s", msg)
	}

	return nil
}
