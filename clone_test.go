package converter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Clone(t *testing.T) {
	type t1 struct {
		Name string
	}

	original := t1{Name: "original"}

	got := t1{}
	err := Clone(original, &got)
	require.NoError(t, err)

	require.Equal(t, original, got)
}
