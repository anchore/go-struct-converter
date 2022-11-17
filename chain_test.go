package converter

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_ConverterChain(t *testing.T) {
	chain := NewChain(V1{}, V2{}, V3{}, V4{})

	tests := []struct {
		name string
		from interface{}
		to   interface{}
		err  string
	}{
		{
			name: "from 1 to 2",
			from: V1{
				Name:   "some-name",
				Field1: "some-value",
			},
			to: V2{
				Name:   "some-name",
				Field2: "some-value",
			},
		},
		{
			name: "from 1 to 3",
			from: V1{
				Field1: "some-value",
			},
			to: V3{
				Field3: "some-value",
			},
		},
		{
			name: "from 2 to 4",
			from: V2{
				Name:   "some-name",
				Field2: "some-value",
			},
			to: V4{
				Name:          "some-name",
				UpdatedField1: "some-value",
			},
		},
		{
			name: "from 4 to 1",
			from: V4{
				Name:          "some-name",
				UpdatedField1: "some-value",
			},
			to: V1{
				Name:   "some-name",
				Field1: "some-value",
			},
		},
		{
			name: "from nil value",
			from: (*V1)(nil),
			to:   V3{},
		},
		{
			name: "invalid FROM type",
			from: Invalid{},
			to:   V1{},
			err:  "invalid FROM type provided, not in the conversion chain: Invalid",
		},
		{
			name: "invalid TO type",
			from: V1{},
			to:   Invalid{},
			err:  "invalid TO type provided, not in the conversion chain: Invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typ := reflect.TypeOf(test.to)
			newInstance := reflect.New(typ)
			result := newInstance.Interface()

			err := chain.Convert(test.from, result)
			if test.err != "" {
				msg := fmt.Sprintf("%v", err)
				if test.err != msg {
					t.Fatalf("expected error '%s' but got: '%v'", test.err, msg)
				}
				return
			} else if err != nil {
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

// a "version 1" struct
type V1 struct {
	Name   string
	Field1 string
}

func (t *V1) ConvertFrom(from interface{}) error {
	if c, ok := from.(V2); ok { // reverse migration
		t.Field1 = c.Field2
	}
	return nil
}

// a "version 2" struct
type V2 struct {
	Name   string
	Field2 string
}

func (t *V2) ConvertFrom(from interface{}) error {
	if c, ok := from.(V1); ok { // forward migration
		t.Field2 = c.Field1
	}
	if c, ok := from.(V3); ok { // reverse migration
		t.Field2 = c.Field3
	}
	return nil
}

// a "version 3" struct
type V3 struct {
	Name   string
	Field3 string
}

func (t *V3) ConvertFrom(from interface{}) error {
	if c, ok := from.(V2); ok { // forward migration
		t.Field3 = c.Field2
	}
	if c, ok := from.(V4); ok { // reverse migration
		t.Field3 = c.UpdatedField1
	}
	return nil
}

// a "version 4" struct
type V4 struct {
	Name          string
	UpdatedField1 string
}

func (t *V4) ConvertFrom(from interface{}) error {
	if c, ok := from.(V3); ok { // forward migration
		t.UpdatedField1 = c.Field3
	}
	return nil
}

// a struct not in the chain
type Invalid struct{}
