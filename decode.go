package structs

import (
	"reflect"
	"sort"
	"strconv"

	"github.com/spf13/cast"
	"github.com/zeebo/errs/v2"
)

// Result contains information about the result of a Decode.
type Result struct {
	Error   error
	Used    map[string]struct{}
	Missing map[string]struct{}
	Broken  map[string]struct{}
}

// DecodeOption controls the operation of a Decode.
type DecodeOption interface {
	private()

	apply(*decodeState)
}

// decodeOptionFunc is an implementation of DecodeOption for a func
type decodeOptionFunc func(*decodeState)

func (o decodeOptionFunc) private()              {}
func (o decodeOptionFunc) apply(ds *decodeState) { o(ds) }

// Decode takes values out of input and stores them into output, allocating as necessary.
func Decode(input map[string]interface{}, output interface{}, opts ...DecodeOption) Result {
	var ds decodeState
	for _, opt := range opts {
		opt.apply(&ds)
	}
	ds.decode(input, reflect.ValueOf(output), "")
	return ds.res
}

// decodeState keeps state across recursive calls to decode.
type decodeState struct {
	res Result
}

// decodeKeyValue decodes into output the value after walking through fields/indexing as described
// by key. It returns true if anything was set. The base is the path the output is at with respect
// to the top most decode.
func (d *decodeState) decodeKeyValue(key string, value interface{}, output reflect.Value, base string) bool {
	nextBase := dotJoin(base, key)

	var rw reflectWalker
	field, err := rw.Walk(output, key)
	if err != nil {
		d.res.Broken = gatherKeys(value, nextBase, d.res.Broken)
		d.res.Error = errs.Combine(d.res.Error, err)
		return false
	}
	if !field.IsValid() {
		d.res.Missing = gatherKeys(value, nextBase, d.res.Missing)
		return false
	}

	if d.decode(value, field, nextBase) {
		rw.Commit()
		return true
	}
	return false
}

// decode looks at the type of input and dispatches to helper routines to decode the input into
// the output. It returns true if anything was set.
func (d *decodeState) decode(input interface{}, output reflect.Value, base string) bool {
	switch input := input.(type) {
	case map[string]interface{}:
		// Go through the keys in sorted order to avoid randomness
		keys := make([]string, 0, len(input))
		for key := range input {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		any := false
		for _, key := range keys {
			any = d.decodeKeyValue(key, input[key], output, base) || any
		}
		return any

	case map[interface{}]interface{}:
		// Filter out the string keys and go through them in sorted order to avoid randomness
		keys := make([]string, 0, len(input))
		for key := range input {
			if key, err := cast.ToStringE(key); err == nil {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)

		any := false
		for _, key := range keys {
			any = d.decodeKeyValue(key, input[key], output, base) || any
		}
		return any

	case []interface{}:
		any := false
		for key, value := range input {
			any = d.decodeKeyValue(strconv.Itoa(key), value, output, base) || any
		}
		return any

	default:
		set, err := setValue(output, input)
		if !set || err != nil {
			d.res.Broken = gatherKeys(input, base, d.res.Broken)
			d.res.Error = errs.Combine(d.res.Error, err)
		} else if set {
			if d.res.Used == nil {
				d.res.Used = make(map[string]struct{})
			}
			d.res.Used[base] = struct{}{}
		}
		return set
	}
}
