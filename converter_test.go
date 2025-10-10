package converter

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ConvertWithKnownTypes(t *testing.T) {
	from := T5{
		Version: "2.2",
		Embedded: T3{
			T1: T1{
				Same:     "same value",
				OldValue: "old value",
			},
		},
	}

	tests := []struct {
		name     string
		chain    []any
		expected any
	}{
		{
			name:  "no conversions",
			chain: []any{},
			expected: T6{
				Version: "2.2",
				Embedded: T4{
					T2: T2{
						Same:     "",
						NewValue: "",
					},
				},
			},
		},
		{
			name:  "top conversion",
			chain: []any{t5_t6},
			expected: T6{
				Version: "2.3",
				Embedded: T4{
					T2: T2{
						Same:     "",
						NewValue: "",
					},
				},
			},
		},
		{
			name:  "mid conversion",
			chain: []any{t3_t4},
			expected: T6{
				Version: "2.2",
				Embedded: T4{
					T2: T2{
						Same:     "same value",
						NewValue: "",
					},
				},
			},
		},
		{
			name:  "all conversions",
			chain: []any{t1_t2, t3_t4, t5_t6},
			expected: T6{
				Version: "2.3",
				Embedded: T4{
					T2: T2{
						Same:     "same value",
						NewValue: "old value",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var to T6

			chain := NewFuncChain(test.chain...).AllowImplicit()

			err := chain.Convert(from, &to)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expected, to); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_Convert(t *testing.T) {
	type alt string

	type s1 struct {
		Value string
	}

	type s2 struct {
		Value string
	}

	tests := []struct {
		name string
		from any
		to   any
	}{
		{
			name: "primitive support",
			from: 25,
			to:   "25",
		},
		{
			name: "from nil value",
			from: (*T1)(nil),
			to: struct {
				Other string
			}{},
		},
		{
			name: "missing properties are omitted",
			from: struct {
				Value string
			}{
				Value: "the value",
			},
			to: struct {
				Other string
			}{},
		},
		{
			name: "missing pointer properties are omitted",
			from: struct {
				Value *string
			}{},
			to: struct {
				Value *string
			}{},
		},
		{
			name: "missing slice properties are omitted",
			from: struct {
				Value []string
			}{},
			to: struct {
				Value []string
			}{},
		},
		{
			name: "nil pointer properties are nil",
			from: struct {
				Value *string
			}{
				Value: nil,
			},
			to: struct {
				Value string
			}{},
		},
		{
			name: "string equals",
			from: struct {
				Value string
			}{
				Value: "the value",
			},
			to: struct {
				Value string
			}{
				Value: "the value",
			},
		},
		{
			name: "string to alt type",
			from: struct {
				Value string
			}{
				Value: "the value",
			},
			to: struct {
				Value alt
			}{
				Value: "the value",
			},
		},
		{
			name: "alt type to string",
			from: struct {
				Value alt
			}{
				Value: "the value",
			},
			to: struct {
				Value string
			}{
				Value: "the value",
			},
		},
		{
			name: "int equals",
			from: struct {
				Int int
			}{
				Int: 2,
			},
			to: struct {
				Int int
			}{
				Int: 2,
			},
		},
		{
			name: "bool equals",
			from: struct {
				Int bool
			}{
				Int: true,
			},
			to: struct {
				Int bool
			}{
				Int: true,
			},
		},
		{
			name: "string ptr equals",
			from: struct {
				StringPtr *string
			}{
				StringPtr: s("the value"),
			},
			to: struct {
				StringPtr *string
			}{
				StringPtr: s("the value"),
			},
		},
		{
			name: "string to ptr equals",
			from: struct {
				StringPtr string
			}{
				StringPtr: "the value",
			},
			to: struct {
				StringPtr *string
			}{
				StringPtr: s("the value"),
			},
		},
		{
			name: "string from ptr equals",
			from: struct {
				StringPtr *string
			}{
				StringPtr: s("the value"),
			},
			to: struct {
				StringPtr string
			}{
				StringPtr: "the value",
			},
		},
		{
			name: "string slice",
			from: struct {
				Strings []string
			}{
				Strings: []string{"the name"},
			},
			to: struct {
				Strings []string
			}{
				Strings: []string{"the name"},
			},
		},
		{
			name: "string ptr slice",
			from: struct {
				StringsPtr []*string
			}{
				StringsPtr: []*string{s("thing 1"), s("thing 2")},
			},
			to: struct {
				StringsPtr []*string
			}{
				StringsPtr: []*string{s("thing 1"), s("thing 2")},
			},
		},
		{
			name: "string ptr to string slice",
			from: struct {
				StringsPtrToStr []*string
			}{
				StringsPtrToStr: []*string{s("thing 1"), s("thing 2")},
			},
			to: struct {
				StringsPtrToStr []string
			}{
				StringsPtrToStr: []string{"thing 1", "thing 2"},
			},
		},
		{
			name: "string slice to ptrs slice",
			from: struct {
				StringsToPtrStr []string
			}{
				StringsToPtrStr: []string{"thing 1", "thing 2"},
			},
			to: struct {
				StringsToPtrStr []*string
			}{
				StringsToPtrStr: []*string{s("thing 1"), s("thing 2")},
			},
		},
		{
			name: "string slice ptr",
			from: struct {
				PtrToStrings *[]string
			}{
				PtrToStrings: &[]string{"the thing 1", "the thing 2"},
			},
			to: struct {
				PtrToStrings *[]string
			}{
				PtrToStrings: &[]string{"the thing 1", "the thing 2"},
			},
		},
		{
			name: "string slice ptr to slice",
			from: struct {
				PtrToStrings *[]string
			}{
				PtrToStrings: &[]string{"the thing 1", "the thing 2"},
			},
			to: struct {
				PtrToStrings []string
			}{
				PtrToStrings: []string{"the thing 1", "the thing 2"},
			},
		},
		{
			name: "string slice ptr to slice",
			from: struct {
				PtrToStrings []string
			}{
				PtrToStrings: []string{"the thing 1", "the thing 2"},
			},
			to: struct {
				PtrToStrings *[]string
			}{
				PtrToStrings: &[]string{"the thing 1", "the thing 2"},
			},
		},
		{
			name: "struct to struct",
			from: struct {
				Struct s1
			}{
				Struct: s1{
					Value: "something",
				},
			},
			to: struct {
				Struct s2
			}{
				Struct: s2{
					Value: "something",
				},
			},
		},
		{
			name: "struct to ptr",
			from: struct {
				Struct s1
			}{
				Struct: s1{
					Value: "something",
				},
			},
			to: struct {
				Struct *s2
			}{
				Struct: &s2{
					Value: "something",
				},
			},
		},
		{
			name: "struct ptr to struct",
			from: struct {
				Struct *s1
			}{
				Struct: &s1{
					Value: "something",
				},
			},
			to: struct {
				Struct s2
			}{
				Struct: s2{
					Value: "something",
				},
			},
		},
		{
			name: "struct ptr to struct ptr",
			from: struct {
				Struct *s1
			}{
				Struct: &s1{
					Value: "something",
				},
			},
			to: struct {
				Struct *s2
			}{
				Struct: &s2{
					Value: "something",
				},
			},
		},
		{
			name: "map to map",
			from: struct {
				Map map[s1]string
			}{
				Map: map[s1]string{
					{Value: "some-key"}: "some-value",
				},
			},
			to: struct {
				Map map[s2]string
			}{
				Map: map[s2]string{
					{Value: "some-key"}: "some-value",
				},
			},
		},
		{
			name: "map to map of ptr",
			from: struct {
				Map map[s1]string
			}{
				Map: map[s1]string{
					{Value: "some-key"}: "some-value",
				},
			},
			to: struct {
				Map map[s2]*string
			}{
				Map: map[s2]*string{
					{Value: "some-key"}: s("some-value"),
				},
			},
		},
		{
			name: "map key ptr to map",
			from: struct {
				Map map[*s1]string
			}{
				Map: map[*s1]string{
					{Value: "some-key"}: "some-value",
				},
			},
			to: struct {
				Map map[s2]string
			}{
				Map: map[s2]string{
					{Value: "some-key"}: "some-value",
				},
			},
		},
		{
			name: "map ptr to map",
			from: struct {
				Map *map[*s1]string
			}{
				Map: &map[*s1]string{
					{Value: "some-key"}: "some-value",
				},
			},
			to: struct {
				Map map[s2]string
			}{
				Map: map[s2]string{
					{Value: "some-key"}: "some-value",
				},
			},
		},
		{
			name: "map string to int",
			from: struct {
				Value string
			}{
				Value: "12",
			},
			to: struct {
				Value int
			}{
				Value: 12,
			},
		},
		{
			name: "map int to string",
			from: struct {
				Value int
			}{
				Value: 12,
			},
			to: struct {
				Value string
			}{
				Value: "12",
			},
		},
		{
			name: "map string slice to string",
			from: struct {
				Value []string
			}{
				Value: []string{"thing 1"},
			},
			to: struct {
				Value string
			}{
				Value: "thing 1",
			},
		},
		{
			name: "map string to string slice",
			from: struct {
				Value string
			}{
				Value: "thing 1",
			},
			to: struct {
				Value []string
			}{
				Value: []string{"thing 1"},
			},
		},
		{
			name: "map int slice to string",
			from: struct {
				Value []int
			}{
				Value: []int{84},
			},
			to: struct {
				Value string
			}{
				Value: "84",
			},
		},
		{
			name: "map string to uint slice",
			from: struct {
				Value string
			}{
				Value: "63",
			},
			to: struct {
				Value []uint
			}{
				Value: []uint{63},
			},
		},
		{
			name: "map string to uint16 slice",
			from: struct {
				Value string
			}{
				Value: "63",
			},
			to: struct {
				Value []uint16
			}{
				Value: []uint16{63},
			},
		},
		{
			name: "test top-level conversion functions",
			from: T3{
				T1: T1{
					Same:     "same value",
					OldValue: "old value",
				},
			},
			to: T4{
				T2: T2{
					Same:     "same value",
					NewValue: "old value",
				},
			},
		},
		{
			name: "test nested conversion functions",
			from: T5{
				Version: "2.2",
				Embedded: T3{
					T1: T1{
						Same:     "same value",
						OldValue: "old value",
					},
				},
			},
			to: T6{
				Version: "2.3",
				Embedded: T4{
					T2: T2{
						Same:     "same value",
						NewValue: "old value",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typ := reflect.TypeOf(test.to)
			newInstance := reflect.New(typ)
			result := newInstance.Interface()

			chain := NewFuncChain(t1_t2, t2_t3, t3_t4, t4_t5, t5_t6).AllowImplicit()
			err := chain.Convert(test.from, result)
			if err != nil {
				t.Fatalf("error during conversion: %v", err)
			}

			// need to align elem vs. pointer of the result
			result = reflect.ValueOf(result).Elem().Interface()

			to := test.to
			if !reflect.DeepEqual(to, result) {
				t.Fatalf("Convert output does not match: %+v %+v", to, result)
			}
		})
	}
}

func s(s string) *string {
	return &s
}

// -------- structs to test conversion functions:

type T1 struct {
	Same     string
	OldValue string
}

type T2 struct {
	Same     string
	NewValue string
}

func t1_t2(t1 T1, t2 *T2) error {
	t2.NewValue = t1.OldValue
	return nil
}

type T3 struct {
	T1 T1
}

func t2_t3(c FuncChain, t2 T2, t3 *T3) error {
	return c.Convert(t2, t3.T1)
}

type T4 struct {
	T2 T2
}

func t3_t4(c FuncChain, t3 T3, t4 *T4) error {
	return c.Convert(t3.T1, &t4.T2)
}

type T5 struct {
	Version  string
	Embedded T3
}

func t4_t5(t4 T4, t5 *T5) error {
	return nil
}

type T6 struct {
	Version  string
	Embedded T4
}

func t5_t6(t5 T5, t6 *T6) error {
	t6.Version = "2.3"
	return nil
}
