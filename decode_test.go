package structs

import (
	"testing"
)

func TestDecode(t *testing.T) {
	type (
		d = map[string]interface{}
		l = []interface{}
	)

	input := d{
		"x": d{
			"y": l{2},
		},

		"z": l{
			d{"q": 0},
			d{"q": 1},
			d{"q": 2},
		},

		"q": d{"f": 6},
	}

	var output struct {
		X struct {
			Y []int
		}
		Z []map[string]interface{}
	}

	res := Decode(input, &output)
	t.Logf("%+v", res)
	t.Logf("%+v", output)
	res = Decode(input, &output)
}

func BenchmarkDecode(b *testing.B) {
	type (
		d = map[string]interface{}
		l = []interface{}
	)

	input := d{
		"x.y.1": 5,
		"x":     d{"y.1": 6},

		"z": l{
			d{"q": 0},
			d{"q": 1},
			d{"q": 2},
		},

		"q": d{"f": 6},
	}

	var output struct {
		X struct {
			Y []int
		}
		Z []map[string]interface{}
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Decode(input, &output)
	}
}
