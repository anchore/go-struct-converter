package converter

import (
	"reflect"
	"testing"
)

func Test_FromTemplate(t *testing.T) {
	type sub struct {
		Subval   string
		Subslice []string
		Submap   map[string]string
	}

	type typ struct {
		Name  string
		Slice []sub
		Map   map[string]sub
		Int   int
	}

	tests := []struct {
		name     string
		template typ
		args     map[string]string
		expected typ
	}{
		{
			name: "multi-level struct with slice and map",
			template: typ{
				Name: "{{.Key1}}",
				Slice: []sub{
					{
						Subval: "{{.Key2}}",
						Subslice: []string{
							"{{.Key1}}",
							"{{.Key2}}",
						},
						Submap: map[string]string{
							"subV1":      "{{.Key1}}",
							"subMissing": "{{.Missing}}",
							"subV3":      "{{.Key3}}",
						},
					},
				},
				Map: map[string]sub{
					"topV1": {
						Subval: "{{.Key2}}",
						Subslice: []string{
							"{{.Key1}}",
							"{{.Key2}}",
						},
						Submap: map[string]string{
							"subV1":      "{{.Key1}}",
							"subMissing": "{{.Missing}}",
							"subV3":      "{{.Key3}}",
						},
					},
				},
				Int: 9,
			},
			args: map[string]string{
				"Key1": "Val1",
				"Key2": "Val2",
				"Key3": "Val3",
			},
			expected: typ{
				Name: "Val1",
				Slice: []sub{
					{
						Subval: "Val2",
						Subslice: []string{
							"Val1",
							"Val2",
						},
						Submap: map[string]string{
							"subV1":      "Val1",
							"subMissing": "",
							"subV3":      "Val3",
						},
					},
				},
				Map: map[string]sub{
					"topV1": {
						Subval: "Val2",
						Subslice: []string{
							"Val1",
							"Val2",
						},
						Submap: map[string]string{
							"subV1":      "Val1",
							"subMissing": "",
							"subV3":      "Val3",
						},
					},
				},
				Int: 9,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FromTemplate(test.args, test.template)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !reflect.DeepEqual(test.expected, got) {
				t.Fatalf("Convert output does not match:\n%+v\n%+v", test.expected, got)
			}
		})
	}
}

func Test_FromTemplateMap(t *testing.T) {
	tests := []struct {
		name     string
		template map[string]string
		args     map[string]string
		expected map[string]string
	}{
		{
			name: "string map",
			template: map[string]string{
				"K1": "{{.Key1}}",
				"K2": "{{.Missing}}",
			},
			args: map[string]string{
				"Key1": "Val1",
				"Key2": "Val2",
				"Key3": "Val3",
			},
			expected: map[string]string{
				"K1": "Val1",
				"K2": "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FromTemplate(test.args, test.template)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !reflect.DeepEqual(test.expected, got) {
				t.Fatalf("Convert output does not match:\n%+v\n%+v", test.expected, got)
			}
		})
	}
}

func Test_FromTemplateSlice(t *testing.T) {
	tests := []struct {
		name     string
		template []string
		args     map[string]string
		expected []string
	}{
		{
			name: "string slice",
			template: []string{
				"{{.Key1}}",
				"{{.Missing}}",
				"{{.Key3}}",
			},
			args: map[string]string{
				"Key1": "Val1",
				"Key2": "Val2",
				"Key3": "Val3",
			},
			expected: []string{
				"Val1",
				"",
				"Val3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FromTemplate(test.args, test.template)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !reflect.DeepEqual(test.expected, got) {
				t.Fatalf("Convert output does not match:\n%+v\n%+v", test.expected, got)
			}
		})
	}
}
