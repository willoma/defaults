package defaults

import (
	"encoding"
	"reflect"
	"strconv"
)

func parseStruct(target reflect.Value, overwrite bool) (reflect.Value, []error) {
	var errs []error

	for i := range target.NumField() {
		typeField := target.Type().Field(i)
		field := target.Field(i)
		defaultValue, hasDefault := typeField.Tag.Lookup("default")

		value, err := parse(field, defaultValue, hasDefault, overwrite)
		if len(err) > 0 {
			errs = append(errs, addErrorsPrefixes(typeField.Name, err)...)

			continue
		}

		if value.IsValid() && (field.IsZero() || overwrite) {
			field.Set(value)
		}
	}

	return target, errs
}

func parse(target reflect.Value, value string, hasDefault, overwrite bool) (reflect.Value, []error) {
	// First, check if we have a specific parser for this type.
	if result, hasParser, errs := parseSpecific(target, value, hasDefault); hasParser {
		return result, errs
	}

	// Then check if it is an [encoding.TextUnmarshaler].
	if target.CanAddr() && target.Addr().Type().Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		result := reflect.New(target.Type()).Elem()
		resp := result.Addr().MethodByName("UnmarshalText").Call([]reflect.Value{reflect.ValueOf([]byte(value))})

		if errIface := resp[0].Interface(); errIface != nil {
			if err, ok := errIface.(error); ok {
				return reflect.Value{}, []error{err}
			}
		}

		return result, nil
	}

	// Then, parse according to the kind of the target.
	return parseKind(target, value, hasDefault, overwrite)
}

//nolint:funlen // We cannot make this shorter :-)
func parseKind(target reflect.Value, value string, hasDefault, overwrite bool) (reflect.Value, []error) {
	switch target.Kind() {
	case reflect.Bool:
		return parseWithError(target, strconv.ParseBool, value, hasDefault)

	case reflect.Int:
		return parseWithError(target, strconv.Atoi, value, hasDefault)

	case reflect.Int8:
		return parseWithErrorII(target, strconv.ParseInt, value, 10, 8, hasDefault)

	case reflect.Int16:
		return parseWithErrorII(target, strconv.ParseInt, value, 10, 16, hasDefault)

	case reflect.Int32:
		return parseWithErrorII(target, strconv.ParseInt, value, 10, 32, hasDefault)

	case reflect.Int64:
		return parseWithErrorII(target, strconv.ParseInt, value, 10, 64, hasDefault)

	case reflect.Uint:
		return parseWithErrorII(target, strconv.ParseUint, value, 10, strconv.IntSize, hasDefault)

	case reflect.Uint8:
		return parseWithErrorII(target, strconv.ParseUint, value, 10, 8, hasDefault)

	case reflect.Uint16:
		return parseWithErrorII(target, strconv.ParseUint, value, 10, 16, hasDefault)

	case reflect.Uint32:
		return parseWithErrorII(target, strconv.ParseUint, value, 10, 32, hasDefault)

	case reflect.Uint64:
		return parseWithErrorII(target, strconv.ParseUint, value, 10, 64, hasDefault)

	case reflect.Float32:
		return parseWithErrorI(target, strconv.ParseFloat, value, 32, hasDefault)

	case reflect.Float64:
		return parseWithErrorI(target, strconv.ParseFloat, value, 64, hasDefault)

	case reflect.Complex64:
		return parseWithErrorI(target, strconv.ParseComplex, value, 64, hasDefault)

	case reflect.Complex128:
		return parseWithErrorI(target, strconv.ParseComplex, value, 128, hasDefault)

	case reflect.Array:
		return parseArray(target, value, hasDefault)

	case reflect.Chan:
		return makeChan(target, value, hasDefault)

	case reflect.Map:
		return parseMap(target, value, hasDefault)

	case reflect.Pointer:
		return parsePointer(target, value, hasDefault, overwrite)

	case reflect.Slice:
		return parseSlice(target, value, hasDefault)

	case reflect.String:
		return reflect.ValueOf(value), nil

	case reflect.Struct:
		return parseStruct(target, overwrite)

	default:
		return reflect.Value{}, []error{ErrUnsupportedType}
	}
}

func parsePointer(target reflect.Value, value string, hasDefault, overwrite bool) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	pointedType := target.Type().Elem()

	if target.IsNil() {
		target = reflect.New(pointedType)
	} else if !overwrite {
		return reflect.Value{}, nil
	}

	result, errs := parse(target.Elem(), value, hasDefault, overwrite)
	if len(errs) > 0 {
		return reflect.Value{}, errs
	}

	resultValue := reflect.New(pointedType)
	resultValue.Elem().Set(result)

	return resultValue, nil
}

func makeChan(target reflect.Value, defaultValue string, hasDefault bool) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	if target.Type().ChanDir() != reflect.BothDir {
		return reflect.Value{}, []error{ErrUnsupportedType}
	}

	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		return reflect.Value{}, []error{err}
	}

	i = max(i, 0)

	return reflect.MakeChan(target.Type(), i), nil
}

func parseWithError[T any](
	target reflect.Value, fct func(string) (T, error), value string, hasDefault bool,
) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	data, err := fct(value)
	if err != nil {
		return reflect.Value{}, []error{err}
	}

	return reflect.ValueOf(data).Convert(target.Type()), nil
}

func parseWithErrorI[T any](
	target reflect.Value, fct func(string, int) (T, error), value string, arg int, hasDefault bool,
) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	data, err := fct(value, arg)
	if err != nil {
		return reflect.Value{}, []error{err}
	}

	return reflect.ValueOf(data).Convert(target.Type()), nil
}

func parseWithErrorII[T any](
	target reflect.Value, fct func(string, int, int) (T, error), value string, arg1, arg2 int, hasDefault bool,
) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	data, err := fct(value, arg1, arg2)
	if err != nil {
		return reflect.Value{}, []error{err}
	}

	return reflect.ValueOf(data).Convert(target.Type()), nil
}
