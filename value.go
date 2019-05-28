package structs

import (
	"reflect"
	"time"

	"github.com/spf13/cast"
	"github.com/zeebo/errs"
)

var ( // cached types for comparison
	durationType = reflect.TypeOf(time.Duration(0))
	timeType     = reflect.TypeOf(time.Time{})
)

// setValue does the best job it can attempting to store the input into the output.
func setValue(output reflect.Value, input interface{}) (bool, error) {
	if !output.CanSet() {
		return false, nil
	}

	var val interface{}
	var err error

	if typ := output.Type(); typ == durationType {
		val, err = cast.ToDurationE(input)
	} else if typ == timeType {
		val, err = cast.ToTimeE(input)
	} else if kind := output.Kind(); kind == reflect.Bool {
		val, err = cast.ToBoolE(input)
	} else if kind == reflect.Int {
		val, err = cast.ToIntE(input)
	} else if kind == reflect.Int8 {
		val, err = cast.ToInt8E(input)
	} else if kind == reflect.Int16 {
		val, err = cast.ToInt16E(input)
	} else if kind == reflect.Int32 {
		val, err = cast.ToInt32E(input)
	} else if kind == reflect.Int64 {
		val, err = cast.ToInt64E(input)
	} else if kind == reflect.Uint {
		val, err = cast.ToUintE(input)
	} else if kind == reflect.Uint8 {
		val, err = cast.ToUint8E(input)
	} else if kind == reflect.Uint16 {
		val, err = cast.ToUint16E(input)
	} else if kind == reflect.Uint32 {
		val, err = cast.ToUint32E(input)
	} else if kind == reflect.Uint64 {
		val, err = cast.ToUint64E(input)
	} else if kind == reflect.Float32 {
		val, err = cast.ToFloat32E(input)
	} else if kind == reflect.Float64 {
		val, err = cast.ToFloat64E(input)
	} else if kind == reflect.String {
		val, err = cast.ToStringE(input)
	} else if kind == reflect.Interface && reflect.TypeOf(input).Implements(output.Type()) {
		val = input
	} else {
		return false, errs.New("can't set input of type %T into output of type %v",
			input, output.Type())
	}
	if err != nil {
		return false, err
	}

	output.Set(reflect.ValueOf(val))
	return true, nil
}
