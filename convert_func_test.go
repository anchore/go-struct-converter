package converter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FuncChainErrors(t *testing.T) {
	tests := []struct {
		name         string
		convertFuncs []any
		errorSubstr  string
	}{
		{
			name:         "no args",
			convertFuncs: []any{func() {}},
			errorSubstr:  "2 or 3",
		},
		{
			name:         "one args",
			convertFuncs: []any{func(_ T1) {}},
			errorSubstr:  "2 or 3",
		},
		{
			name:         "four args",
			convertFuncs: []any{func(_ T1, _ *T2, _ T2, _ T3) {}},
			errorSubstr:  "2 or 3",
		},
		{
			name:         "three args",
			convertFuncs: []any{func(_ T1, _ *T2, _ T1) {}},
			errorSubstr:  "must the first",
		},
		{
			name:         "same args",
			convertFuncs: []any{func(_ T1, _ T1) {}},
			errorSubstr:  "different types",
		},
		{
			name:         "2 arg func chain",
			convertFuncs: []any{func(_ FuncChain, _ T1) {}},
			errorSubstr:  "2 more",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var panicValue any
			func() {
				defer func() {
					panicValue = recover()
				}()
				_ = NewFuncChain(test.convertFuncs...)
			}()
			err, isErr := panicValue.(error)
			require.True(t, isErr)
			require.ErrorContains(t, err, test.errorSubstr)
		})
	}
}

func Test_FuncChain(t *testing.T) {
	chain := NewFuncChain(t1ToT2, t2ToT1, t2ToT3, t3ToT2, t3ToT4, t4ToT5, t3ToT5)

	from := t1{
		Name:    "name-value-from-1",
		Custom1: "custom-value-from-1",
	}

	to := t3{}

	err := chain.Convert(from, &to)
	require.NoError(t, err)
	require.Equal(t, from.Name, to.Name)
	require.Equal(t, from.Custom1, to.Custom3)

	backToT1 := t1{}
	err = chain.Convert(to, &backToT1)
	require.NoError(t, err)
	require.Equal(t, from.Name, backToT1.Name)
	require.Equal(t, from.Custom1, backToT1.Custom1)

	shortestToT5 := t5{}
	err = chain.Convert(to, &shortestToT5)
	require.NoError(t, err)
	require.Equal(t, "FromT3", shortestToT5.Name)
}

type t1 struct {
	Name    string
	Custom1 string
}
type t2 struct {
	Name    string
	Custom2 string
}
type t3 struct {
	Name    string
	Custom3 string
}
type t4 struct {
	Name    string
	Custom3 string
}
type t5 struct {
	Name    string
	Custom3 string
}

func t1ToT2(t1 t1, t2 *t2) error {
	t2.Custom2 = t1.Custom1
	return nil
}

func t2ToT3(t2 t2, t3 *t3) error {
	t3.Custom3 = t2.Custom2
	return nil
}

func t3ToT2(_ FuncChain, t3 t3, t2 *t2) error {
	t2.Custom2 = t3.Custom3
	return nil
}

func t2ToT1(_ FuncChain, t2 t2, t1 *t1) error {
	t1.Custom1 = t2.Custom2
	return nil
}

func t3ToT4(_ FuncChain, _ t3, _ *t4) error {
	return nil
}

func t4ToT5(_ FuncChain, _ t4, t5 *t5) error {
	t5.Name = "FromT4"
	return nil
}

func t3ToT5(_ FuncChain, _ t3, t5 *t5) error {
	t5.Name = "FromT3"
	return nil
}
