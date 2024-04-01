package converter

import (
	"fmt"
	"reflect"
	"strconv"
)

// Convert takes two objects, e.g. v2_1.Document and &v2_2.Document{} and attempts to map all the properties from one
// to the other. After the automatic mapping, if a struct implements the ConvertFrom interface, this is called to
// perform any additional conversion logic necessary.
func (c funcChain) convert(fromValue reflect.Value, toValuePtr reflect.Value) error {
	toTypePtr := toValuePtr.Type()

	if !isPtr(toTypePtr) {
		return fmt.Errorf("TO value provided was not a pointer, unable to set value: %+v", toValuePtr)
	}

	toValue, err := c.getValue(fromValue, toTypePtr)
	if err != nil {
		return err
	}

	// don't set nil values
	if toValue == nilValue {
		return nil
	}

	// toValuePtr is the passed-in pointer, toValue is also the same type of pointer
	toValuePtr.Elem().Set(toValue.Elem())
	return nil
}

func (c funcChain) getValue(fromValue reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	var err error

	fromType := fromValue.Type()

	var toValue reflect.Value

	// handle incoming pointer Types
	if isPtr(fromType) {
		if fromValue.IsNil() {
			return nilValue, nil
		}
		fromValue = fromValue.Elem()
		if !fromValue.IsValid() || fromValue.IsZero() {
			return nilValue, nil
		}
		fromType = fromValue.Type()
	}

	baseTargetType := targetType
	if isPtr(targetType) {
		baseTargetType = targetType.Elem()
	}

	switch {
	case isStruct(fromType) && isStruct(baseTargetType):
		// this always creates a pointer type
		toValue = reflect.New(baseTargetType)
		toValue = toValue.Elem()

		for i := 0; i < fromType.NumField(); i++ {
			fromField := fromType.Field(i)
			fromFieldValue := fromValue.Field(i)

			toField, exists := baseTargetType.FieldByName(fromField.Name)
			if !exists {
				continue
			}
			toFieldType := toField.Type

			toFieldValue := toValue.FieldByName(toField.Name)

			newValue, err := c.getValue(fromFieldValue, toFieldType)
			if err != nil {
				return nilValue, err
			}

			if newValue == nilValue {
				continue
			}

			toFieldValue.Set(newValue)
		}

		// check for custom convert functions from previous/next version struct

		if c.funcs[fromType] != nil && c.funcs[fromType][baseTargetType] != nil {
			convertFunc := c.funcs[fromType][baseTargetType]
			err = convertFunc(fromValue, toValue.Addr())
			if err != nil {
				return nilValue, fmt.Errorf("an error occurred calling %s.%s: %v", baseTargetType.Name(), convertFromName, err)
			}
		}
	case isSlice(fromType) && isSlice(baseTargetType):
		if fromValue.IsNil() {
			return nilValue, nil
		}

		length := fromValue.Len()
		targetElementType := baseTargetType.Elem()
		toValue = reflect.MakeSlice(baseTargetType, length, length)
		for i := 0; i < length; i++ {
			v, err := c.getValue(fromValue.Index(i), targetElementType)
			if err != nil {
				return nilValue, err
			}
			if v.IsValid() {
				toValue.Index(i).Set(v)
			}
		}
	case isMap(fromType) && isMap(baseTargetType):
		if fromValue.IsNil() {
			return nilValue, nil
		}

		keyType := baseTargetType.Key()
		elementType := baseTargetType.Elem()
		toValue = reflect.MakeMap(baseTargetType)
		for _, fromKey := range fromValue.MapKeys() {
			fromVal := fromValue.MapIndex(fromKey)
			k, err := c.getValue(fromKey, keyType)
			if err != nil {
				return nilValue, err
			}
			v, err := c.getValue(fromVal, elementType)
			if err != nil {
				return nilValue, err
			}
			if k == nilValue || v == nilValue {
				continue
			}
			if v == nilValue {
				continue
			}
			if k.IsValid() && v.IsValid() {
				toValue.SetMapIndex(k, v)
			}
		}
	default:
		toValue = fromValue
	}

	// handle non-pointer returns -- the reflect.New earlier always creates a pointer
	if !isPtr(baseTargetType) {
		toValue = fromPtr(toValue)
	}

	toValue, err = c.convertValueTypes(toValue, baseTargetType)

	if err != nil {
		return nilValue, err
	}

	// handle elements which are now pointers
	if isPtr(targetType) {
		toValue = toPtr(toValue)
	}

	return toValue, nil
}

// convertValueTypes takes a value and a target type, and attempts to convert
// between the Types - e.g. string -> int. when this function is called the value
func (c funcChain) convertValueTypes(value reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	typ := value.Type()
	switch {
	// if the Types are the same, just return the value
	case typ == targetType:
		return value, nil
	case typ.Kind() == targetType.Kind() && typ.ConvertibleTo(targetType):
		return value.Convert(targetType), nil
	case value.IsZero() && isPrimitive(targetType):
		// do nothing, will return nilValue
	case isPrimitive(typ) && isPrimitive(targetType):
		// get a string representation of the value
		str := fmt.Sprintf("%v", value.Interface()) // TODO is there a better way to get a string representation?
		var err error
		var out interface{}
		switch {
		case isString(targetType):
			out = str
		case isBool(targetType):
			out, err = strconv.ParseBool(str)
		case isInt(targetType):
			out, err = strconv.Atoi(str)
		case isUint(targetType):
			out, err = strconv.ParseUint(str, 10, 64)
		case isFloat(targetType):
			out, err = strconv.ParseFloat(str, 64)
		}

		if err != nil {
			return nilValue, err
		}

		v := reflect.ValueOf(out)

		v = v.Convert(targetType)

		return v, nil
	case isSlice(typ) && isSlice(targetType):
		// this should already be handled in getValue
	case isSlice(typ):
		// this may be lossy
		if value.Len() > 0 {
			v := value.Index(0)
			v, err := c.convertValueTypes(v, targetType)
			if err != nil {
				return nilValue, err
			}
			return v, nil
		}
		return c.convertValueTypes(nilValue, targetType)
	case isSlice(targetType):
		elementType := targetType.Elem()
		v, err := c.convertValueTypes(value, elementType)
		if err != nil {
			return nilValue, err
		}
		if v == nilValue {
			return v, nil
		}
		slice := reflect.MakeSlice(targetType, 1, 1)
		slice.Index(0).Set(v)
		return slice, nil
	}

	return nilValue, fmt.Errorf("unable to convert from: %v to %v", value.Interface(), targetType.Name())
}
