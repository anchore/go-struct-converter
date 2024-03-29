package converter

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"text/template"
)

func FromTemplate[T any](args any, templateObject T) (T, error) {
	var out T
	err := Convert(templateObject, &out)
	if err == nil {
		err = applyTemplateArgs(args, reflect.ValueOf(&out), nil)
	}
	return out, err
}

func applyTemplateArgs(args any, v reflect.Value, path []reflect.Value) error {
	t := v.Type()

	for isPtr(t) {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	switch {
	case isString(t):
		return applyStringTemplateArgs(args, v, path)
	case isStruct(t):
		return applyStructTemplateArgs(args, v, path)
	case isSlice(t):
		return applySliceTemplateArgs(args, v, path)
	case isMap(t):
		return applyMapTemplateArgs(args, v, path)
	}

	return nil
}

var parseMutex sync.Mutex
var parsed = map[string]*template.Template{}

func execTemplate(tpl string, args any, path []reflect.Value) (string, error) {
	var err error
	t := parsed[tpl]
	if t == nil {
		parseMutex.Lock()
		defer parseMutex.Unlock()
		t = parsed[tpl]
		if t == nil {
			t = template.New("").Option("missingkey=zero")
			t, err = t.Parse(tpl)
			if err != nil {
				return "", fmt.Errorf("unable to parse template at %s: %w", printPath(path), err)
			}
			parsed[tpl] = t
		}
	}
	buf := bytes.Buffer{}
	err = t.Execute(&buf, args)
	if err != nil {
		return "", fmt.Errorf("unable to execute template at %s: %w", printPath(path), err)
	}
	return buf.String(), nil
}

func printPath(path []reflect.Value) string {
	out := ""
	for _, v := range path {
		if out != "" {
			out += "."
		}
		out += fmt.Sprintf("%v", v)
	}
	return out
}

func applyStringTemplateArgs(args any, v reflect.Value, path []reflect.Value) error {
	val, err := execTemplate(v.String(), args, path)
	if err != nil {
		return err
	}
	if !v.CanSet() {
		err = fmt.Errorf("unable to set value at %s", printPath(path))
		return err
	}
	v.SetString(val)
	return nil
}

func applyMapTemplateArgs(args any, v reflect.Value, path []reflect.Value) error {
	for _, k := range v.MapKeys() {
		item := v.MapIndex(k)
		switch item.Type().Kind() {
		// we cannot update certain types of things within maps by using, need to create a new one and set the value
		case reflect.String:
			val, err := execTemplate(item.String(), args, append(path, k))
			if err != nil {
				return err
			}
			v.SetMapIndex(k, reflect.ValueOf(val))
		case reflect.Struct:
			newVal := reflect.New(item.Type()).Elem()
			newVal.Set(item)
			err := applyTemplateArgs(args, newVal, append(path, k))
			if err != nil {
				return err
			}
			v.SetMapIndex(k, newVal)
		default:
			err := applyTemplateArgs(args, item, append(path, k))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func applySliceTemplateArgs(args any, v reflect.Value, path []reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		err := applyTemplateArgs(args, item, append(path, reflect.ValueOf(i)))
		if err != nil {
			return err
		}
	}
	return nil
}

func applyStructTemplateArgs(args any, v reflect.Value, path []reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		item := v.Field(i)
		err := applyTemplateArgs(args, item, append(path, reflect.ValueOf(field.Name)))
		if err != nil {
			return err
		}
	}
	return nil
}
