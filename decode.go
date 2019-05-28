package structs

import (
	"fmt"
	"reflect"

	"github.com/zeebo/errs"
)

type Result struct {
	Error   error
	Missing []string
	Broken  []string
}

type Option interface {
	private()
}

func Decode(input map[string]interface{}, output interface{}, opts ...Option) Result {
	var ds decodeState
	ds.decode(input, reflect.ValueOf(output), "")
	return ds.res
}

type decodeState struct {
	res Result
}

func (d *decodeState) decode(input interface{}, output reflect.Value, base string) {
	switch input := input.(type) {
	case map[string]interface{}:
		for key, value := range input {
			d.decodeKeyValue(key, value, output, base)
		}

	case []interface{}:
		for key, value := range input {
			d.decodeKeyValue(fmt.Sprint(key), value, output, base)
		}

	default:
		// TODO: scalar
		output.Set(reflect.ValueOf(input))
	}
}

func (d *decodeState) decodeKeyValue(key string, value interface{}, output reflect.Value, base string) {
	nextBase := dotJoin(base, key)

	var rw reflectWalker
	field, err := rw.Walk(output, key)
	if err != nil {
		d.res.Broken = append(d.res.Broken, gatherKeys(value, nextBase, nil)...)
		d.res.Error = errs.Combine(d.res.Error, err)
		return
	}
	if !field.IsValid() {
		d.res.Missing = append(d.res.Missing, gatherKeys(value, nextBase, nil)...)
		return
	}

	d.decode(value, field, nextBase)
	rw.Commit()
}
