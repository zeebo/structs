package structs

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/zeebo/errs"
)

// stringType is a cached type of strings.
var stringType = reflect.TypeOf("")

// reflectWalker holds buffers to make walking efficient.
type reflectWalker struct {
	cbs callbacks
}

// Commit performs stored mutations from Walk. Should be called after setting the field
// returned by Walk.
func (r *reflectWalker) Commit() {
	r.cbs.Run()
	r.cbs.Reset()
}

// removeHyphens removes hyphens from strings
var removeHyphens = strings.NewReplacer("-", "")

// Walk will walk the dotted path in the value allocating and descending through structs and
// maps along the way.
func (r *reflectWalker) Walk(val reflect.Value, path string) (out reflect.Value, err error) {
	dw := newDotWalker(path)
	r.cbs.Reset()
	for !dw.Empty() && val.IsValid() {
		val = r.indirectAlloc(val)
		val, err = r.selectField(val, dw.Next())
		if err != nil {
			return out, err
		}
	}
	return val, nil
}

// selectField returns a reflect.Value that selects for the provided field name. It looks up in
// maps of strings as well as embedded fields.
func (r *reflectWalker) selectField(val reflect.Value, name string) (reflect.Value, error) {
	switch val.Kind() {
	case reflect.Struct:
		// Figure out the set of field indicies to walk to the name for the type.
		idx, buf, ok := searchEmbedded(val.Type(), removeHyphens.Replace(name), nil)
		if !ok {
			return reflect.Value{}, nil
		}

		// Allocate and walk those field indicies.
		val = r.indirectAlloc(val.Field(idx))
		for _, i := range buf {
			val = r.indirectAlloc(val.Field(i))
		}

		return val, nil

	case reflect.Map:
		if val.Type().Key() != stringType {
			return reflect.Value{}, errs.New("attempt to walk into invalid map type: %v", val.Type())
		}

		// Sadly have to allocate a copy of the key to make it settable.
		key, value := reflect.ValueOf(name), reflect.New(val.Type().Elem()).Elem()
		if existing := val.MapIndex(key); existing.IsValid() {
			value.Set(existing)
		}
		r.cbs.Append(func() { val.SetMapIndex(key, value) })
		return value, nil

	case reflect.Slice, reflect.Array:
		// Check if the name is an integer.
		index, err := strconv.Atoi(name)
		if err != nil {
			return reflect.Value{}, errs.New("attempt to do numeric index with %q: %v", name, err)
		}

		// Check if we have enough length.
		if index < val.Len() {
			return val.Index(index), nil
		}

		// If we're an array, it's out of bounds
		if val.Kind() == reflect.Array {
			return reflect.Value{}, nil
		}

		// Otherwise, grow the capacity of the slice.
		nextVal, elemType := val, val.Type().Elem()
		for index >= nextVal.Len() {
			nextVal = reflect.Append(nextVal, reflect.Zero(elemType))
		}
		r.cbs.Append(func() { val.Set(nextVal) })
		return nextVal.Index(index), nil
	}

	return reflect.Value{}, errs.New("attempt to walk into invalid type: %v", val.Type())
}

// searchEmbedded looks through the type finding a field matching name, recursing through
// the embedded fields. It returns a list of indexes to the field with the name
func searchEmbedded(typ reflect.Type, name string, buf []int) (int, []int, bool) {
	// We have to walk the list of fields even through embedded fields
	// so we recursively look depth first.

	hadAnon, nf := false, typ.NumField()
	for i := 0; i < nf; i++ {
		// TODO(jeff): sadly, this allocates. this is why hadAnon exists.
		field := typ.Field(i)

		// Check if any of the names match first.
		if strings.EqualFold(field.Name, name) {
			return i, buf, true
		}

		if field.Anonymous && field.PkgPath == "" {
			hadAnon = true
		}
	}

	// If there were no anonymous fields, then there's no hope. :(
	if !hadAnon {
		return 0, nil, false
	}

	// Check for all the anonymous fields
	olen := len(buf)
	for i := 0; i < nf && hadAnon; i++ {
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}

		fieldTyp := field.Type
		if fieldTyp.Kind() == reflect.Ptr {
			fieldTyp = fieldTyp.Elem()
		}
		if fieldTyp.Kind() != reflect.Struct {
			continue
		}

		buf = append(buf[:olen], i)
		if idx, buf, ok := searchEmbedded(fieldTyp, name, buf); ok {
			return idx, buf, true
		}
	}

	// Nothing matched :(
	return 0, nil, false
}

// indirectAlloc will indirect pointers and allocate nil maps and pointers.
func (r *reflectWalker) indirectAlloc(val reflect.Value) reflect.Value {
	for {
		switch kind := val.Kind(); {
		// If we have a pointer, we should deref it
		case kind == reflect.Ptr:
			// Make sure that the pointer contains a value before deref
			if val.IsNil() {
				curVal, nextVal := val, reflect.New(val.Type().Elem())
				r.cbs.Append(func() { curVal.Set(nextVal) })
				val = nextVal
			} else {
				val = val.Elem()
			}

		// If we have a map, make sure it's made
		case kind == reflect.Map && val.IsNil():
			curVal, nextVal := val, reflect.MakeMap(val.Type())
			r.cbs.Append(func() { curVal.Set(nextVal) })
			val = nextVal

		// Otherwise, we're done indirecting
		default:
			return val
		}
	}
}
