package freeLockCache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_object2bytes(t *testing.T) {
	type A struct {
		F string
	}
	a := A{F: "hello"}
	bytes, _ := serialize(a)
	var b A
	_ = deserialize(bytes, &b)
	assert.Equal(t, a, b)
}
