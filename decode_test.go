package structs

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestDecode(t *testing.T) {
	type (
		d = map[string]interface{}
		l = []interface{}
	)

	s := func(ks ...string) map[string]struct{} {
		if len(ks) == 0 {
			return nil
		}
		out := make(map[string]struct{})
		for _, k := range ks {
			out[k] = struct{}{}
		}
		return out
	}

	t.Run("Basic", func(t *testing.T) {
		input := d{
			"x": d{
				"y": l{2},
			},
			"x.y.0": 3, // should always take precedence

			"z": l{
				d{"q": 0},
				d{"q": 1},
				d{"q": 2},
			},

			"q": d{"f": 6}, // missing
			"f": "broken",  // broken
		}

		type into struct {
			X struct {
				Y []int
			}
			Z []map[string]interface{}
			F int
		}
		var output into

		res := Decode(input, &output)
		assert.Error(t, res.Error)
		assert.DeepEqual(t, res.Used, s("x.y.0", "x.y.0", "z.0.q", "z.1.q", "z.2.q"))
		assert.DeepEqual(t, res.Missing, s("q.f"))
		assert.DeepEqual(t, res.Broken, s("f"))
		assert.DeepEqual(t, output, into{
			X: struct{ Y []int }{Y: []int{3}},
			Z: []map[string]interface{}{{"q": 0}, {"q": 1}, {"q": 2}},
		})
	})

	t.Run("Compound In Map", func(t *testing.T) {
		var output struct{ X map[string]struct{ Y, Z int } }
		input := d{"x.a.y": 1, "x.a.z": 2}

		res := Decode(input, &output)
		assert.NoError(t, res.Error)
		assert.DeepEqual(t, res.Used, s("x.a.y", "x.a.z"))
		assert.DeepEqual(t, res.Missing, s())
		assert.DeepEqual(t, res.Broken, s())
		assert.Equal(t, output.X["a"].Y, 1)
		assert.Equal(t, output.X["a"].Z, 2)
	})

	t.Run("Avoids Writes", func(t *testing.T) {
		var output map[string]*struct {
			X int
			Y *struct{ Z int }
		}
		input := d{"a.x": 1, "b.q": 2, "a.y.f": 3}

		res := Decode(input, &output)
		assert.NoError(t, res.Error)
		assert.DeepEqual(t, res.Used, s("a.x"))
		assert.DeepEqual(t, res.Missing, s("a.y.f", "b.q"))
		assert.DeepEqual(t, res.Broken, s())
		assert.Equal(t, output["a"].X, 1)
		assert.Equal(t, output["a"].Y, (*struct{ Z int })(nil))
		assert.Equal(t, len(output), 1)
	})

	t.Run("Embedding", func(t *testing.T) {
		type E3 struct{ X int }
		type E2 struct{ *E3 }
		type E1 struct{ E2 }
		var output struct{ E1 }
		input := d{"x": 1}

		res := Decode(input, &output)
		assert.NoError(t, res.Error)
		assert.DeepEqual(t, res.Used, s("x"))
		assert.DeepEqual(t, res.Missing, s())
		assert.DeepEqual(t, res.Broken, s())
		assert.Equal(t, output.X, 1)
	})

	t.Run("Embed Unexported", func(t *testing.T) {
		type e3 struct{ X int }
		type e2 struct{ *e3 }
		type e1 struct{ e2 }
		var output struct{ e1 }
		input := d{"x": 1}

		res := Decode(input, &output)
		assert.NoError(t, res.Error)
		assert.DeepEqual(t, res.Used, s())
		assert.DeepEqual(t, res.Missing, s("x"))
		assert.DeepEqual(t, res.Broken, s())
		assert.Equal(t, output.e1.e2.e3, (*e3)(nil))
	})
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
