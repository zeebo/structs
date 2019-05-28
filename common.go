package structs

import (
	"fmt"
	"strings"
)

type dotWalker struct {
	data  string
	first bool
}

func newDotWalker(data string) dotWalker {
	return dotWalker{data: data, first: true}
}

func (d dotWalker) Empty() bool { return d.data == "" }

func (d *dotWalker) Next() (out string) {
	if d.data == "" {
		return ""
	} else if d.data[0] == '.' {
		if d.first {
			d.first = false
			return ""
		}
		d.data = d.data[1:]
	}
	d.first = false
	dot := strings.IndexByte(d.data, '.')
	if dot < 0 {
		dot = len(d.data)
	}
	out, d.data = d.data[:dot], d.data[dot:]
	return out
}

//
//
//

type callbacks []func()

func (c *callbacks) Reset() {
	for i := range *c {
		(*c)[i] = nil
	}
	*c = (*c)[:0]
}

func (c *callbacks) Append(cb func()) {
	*c = append(*c, cb)
}

func (c *callbacks) Run() {
	for len(*c) > 0 {
		cb := (*c)[len(*c)-1]
		(*c)[len(*c)-1] = nil
		*c = (*c)[:len(*c)-1]
		cb()
	}
}

//
//
//

func dotJoin(base, part string) string {
	if base == "" {
		return part
	}
	return base + "." + part
}

//
//
//

func gatherKeys(from interface{}, base string, into map[string]struct{}) map[string]struct{} {
	if into == nil {
		into = make(map[string]struct{})
	}
	switch from := from.(type) {
	case map[string]interface{}:
		for key, value := range from {
			into = gatherKeys(value, dotJoin(base, key), into)
		}

	case []interface{}:
		for key, value := range from {
			into = gatherKeys(value, dotJoin(base, fmt.Sprint(key)), into)
		}

	default:
		if base != "" {
			into[base] = struct{}{}
		}
	}

	return into
}
