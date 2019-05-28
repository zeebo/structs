package structs

import (
	"reflect"
	"testing"

	"github.com/zeebo/assert"
)

func TestReflectWalker(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		var x struct {
			Y *struct {
				Z **struct{ M map[string]*struct{ F int } }
			}
		}

		var rw reflectWalker
		val, err := rw.Walk(reflect.ValueOf(&x), "y.z.m.key.f")
		assert.NoError(t, err)
		val.SetInt(1)
		rw.Commit()

		assert.Equal(t, (*x.Y.Z).M["key"].F, 1)
	})

	t.Run("Slice", func(t *testing.T) {
		var x []map[string]interface{}

		var rw reflectWalker
		val, err := rw.Walk(reflect.ValueOf(&x), "3.f")
		assert.NoError(t, err)
		assert.That(t, val.IsValid())
		val.Set(reflect.ValueOf(1))
		rw.Commit()

		assert.DeepEqual(t, x, []map[string]interface{}{3: {"f": 1}})
	})

	t.Run("Underscore", func(t *testing.T) {
		var x struct{ FooBar int }

		var rw reflectWalker
		val, err := rw.Walk(reflect.ValueOf(&x), "foo-bar")
		assert.NoError(t, err)
		assert.That(t, val.IsValid())
		val.Set(reflect.ValueOf(1))
		rw.Commit()

		assert.Equal(t, x.FooBar, 1)
	})
}

func BenchmarkReflectWalker(b *testing.B) {
	var rw reflectWalker
	x := reflect.ValueOf(&struct {
		Y *struct {
			Z **struct{ M map[string]*struct{ F int } }
		}
	}{})

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		val, _ := rw.Walk(x, "y.z.m.key.f")
		val.SetInt(1)
		rw.Commit()
	}
}
