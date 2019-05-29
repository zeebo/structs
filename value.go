package structs

import (
	"encoding/csv"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/zeebo/errs"
)

type setter interface {
	Set(s string) error
}

var ( // cached types for comparison
	durationType    = reflect.TypeOf(time.Duration(0))
	timeType        = reflect.TypeOf(time.Time{})
	setterType      = reflect.TypeOf((*setter)(nil)).Elem()
	stringSliceType = reflect.TypeOf([]string(nil))
)

// setValue does the best job it can attempting to store the input into the output.
func setValue(output reflect.Value, input interface{}) (set bool, err error) {
	if !output.CanSet() {
		return false, nil
	}

	var val interface{}
	switch typ, kind := output.Type(), output.Kind(); {
	// specific concrete types
	case typ == durationType:
		val, err = cast.ToDurationE(input)
	case typ == timeType:
		val, err = cast.ToTimeE(input)
	case typ == stringSliceType:
		var sval string
		sval, err = cast.ToStringE(input)
		if err != nil {
			return false, err
		}
		sval = strings.TrimPrefix(sval, "[")
		sval = strings.TrimSuffix(sval, "]")
		val, err = csv.NewReader(strings.NewReader(sval)).Read()

	// if it can be set by string, do that
	case typ.Implements(setterType):
		sval, err := cast.ToStringE(input)
		if err != nil {
			return false, err
		}
		return true, output.Interface().(setter).Set(sval)
	case output.CanAddr() && reflect.PtrTo(typ).Implements(setterType):
		sval, err := cast.ToStringE(input)
		if err != nil {
			return false, err
		}
		return true, output.Addr().Interface().(setter).Set(sval)

	// go by kind
	case kind == reflect.Bool:
		val, err = cast.ToBoolE(input)
	case kind == reflect.Int:
		val, err = cast.ToIntE(input)
	case kind == reflect.Int8:
		val, err = cast.ToInt8E(input)
	case kind == reflect.Int16:
		val, err = cast.ToInt16E(input)
	case kind == reflect.Int32:
		val, err = cast.ToInt32E(input)
	case kind == reflect.Int64:
		val, err = cast.ToInt64E(input)
	case kind == reflect.Uint:
		val, err = cast.ToUintE(input)
	case kind == reflect.Uint8:
		val, err = cast.ToUint8E(input)
	case kind == reflect.Uint16:
		val, err = cast.ToUint16E(input)
	case kind == reflect.Uint32:
		val, err = cast.ToUint32E(input)
	case kind == reflect.Uint64:
		val, err = cast.ToUint64E(input)
	case kind == reflect.Float32:
		val, err = cast.ToFloat32E(input)
	case kind == reflect.Float64:
		val, err = cast.ToFloat64E(input)
	case kind == reflect.String:
		val, err = cast.ToStringE(input)

	// check interface matching
	case kind == reflect.Interface && reflect.TypeOf(input).Implements(typ):
		val = input

	// ran out of options
	default:
		return false, errs.New("can't set input of type %T into output of type %v",
			input, output.Type())
	}

	if err != nil {
		return false, err
	}

	output.Set(reflect.ValueOf(val))
	return true, nil
}
