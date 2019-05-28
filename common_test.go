package structs

import (
	"sort"
	"testing"

	"github.com/zeebo/assert"
)

func TestDotWalker(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		empty := newDotWalker("")
		assert.That(t, empty.Empty())
		assert.Equal(t, empty.Next(), "")
	})

	t.Run("One", func(t *testing.T) {
		one := newDotWalker("one")
		assert.That(t, !one.Empty())
		assert.Equal(t, one.Next(), "one")
		assert.That(t, one.Empty())
		assert.Equal(t, one.Next(), "")
	})

	t.Run("Two", func(t *testing.T) {
		two := newDotWalker("one.two")
		assert.That(t, !two.Empty())
		assert.Equal(t, two.Next(), "one")
		assert.That(t, !two.Empty())
		assert.Equal(t, two.Next(), "two")
		assert.That(t, two.Empty())
		assert.Equal(t, two.Next(), "")
	})

	t.Run("Trailing", func(t *testing.T) {
		trailing := newDotWalker("one..")
		assert.That(t, !trailing.Empty())
		assert.Equal(t, trailing.Next(), "one")
		assert.That(t, !trailing.Empty())
		assert.Equal(t, trailing.Next(), "")
		assert.That(t, !trailing.Empty())
		assert.Equal(t, trailing.Next(), "")
		assert.That(t, trailing.Empty())
		assert.Equal(t, trailing.Next(), "")
	})

	t.Run("Leading", func(t *testing.T) {
		leading := newDotWalker("..one")
		assert.That(t, !leading.Empty())
		assert.Equal(t, leading.Next(), "")
		assert.That(t, !leading.Empty())
		assert.Equal(t, leading.Next(), "")
		assert.That(t, !leading.Empty())
		assert.Equal(t, leading.Next(), "one")
		assert.That(t, leading.Empty())
		assert.Equal(t, leading.Next(), "")
	})
}

func TestCallbacks(t *testing.T) {
	x, y, z := 0, 0, 0

	var cbs callbacks
	cbs.Append(func() { z = y + 1 })
	cbs.Append(func() { y = x + 1 })
	cbs.Append(func() { x = 1 })

	cbs.Run()
	assert.Equal(t, x, 1)
	assert.Equal(t, y, 2)
	assert.Equal(t, z, 3)
}

func TestGatherKeys(t *testing.T) {
	type (
		d = map[string]interface{}
		l = []interface{}
	)

	from := d{"foo": 2, "bar": d{"baz": 3, "bif": d{"x": l{1, 2, d{"z": 3}}}}}

	keys := gatherKeys(from, "", nil)
	sort.Strings(keys)

	assert.DeepEqual(t, keys, []string{
		"bar.baz",
		"bar.bif.x.0",
		"bar.bif.x.1",
		"bar.bif.x.2.z",
		"foo",
	})
}
