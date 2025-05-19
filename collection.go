package defaults

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parseArray(target reflect.Value, value string, hasDefault bool) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	defaults := asList(value)

	if tgtLen, defLen := target.Len(), len(defaults); defLen != tgtLen {
		return reflect.Value{}, []error{fmt.Errorf("%w: expected %d values, got %d", ErrInvalidFormat, tgtLen, defLen)}
	}

	arrayType := target.Type()
	result := reflect.New(arrayType).Elem()
	errs := parseList(arrayType.Elem(), result, defaults)

	return result, errs
}

func parseSlice(target reflect.Value, value string, hasDefault bool) (reflect.Value, []error) {
	if !hasDefault {
		return reflect.Value{}, nil
	}

	defaults := asList(value)
	result := reflect.MakeSlice(target.Type(), len(defaults), len(defaults))
	errs := parseList(target.Type().Elem(), result, defaults)

	return result, errs
}

func parseList(itemsType reflect.Type, result reflect.Value, defaults []string) []error {
	var errs []error

	zero := reflect.Zero(itemsType)

	for i, def := range defaults {
		value, err := parse(zero, def, true, true)
		if len(err) > 0 {
			errs = append(errs, addErrorsPrefixes(strconv.Itoa(i), err)...)

			continue
		}

		result.Index(i).Set(value)
	}

	return errs
}

func parseMap(target reflect.Value, defaultValue string, hasDefaults bool) (reflect.Value, []error) {
	if !hasDefaults {
		return reflect.Value{}, nil
	}

	var errs []error

	defaults := asList(defaultValue)
	targetType := target.Type()
	result := reflect.MakeMapWithSize(targetType, len(defaults))
	keyType := targetType.Key()
	valueType := targetType.Elem()
	keyZero := reflect.Zero(keyType)
	valueZero := reflect.Zero(valueType)

	for i, def := range defaults {
		keyValue := strings.SplitN(def, ":", 2)

		if len(keyValue) != 2 {
			errs = append(
				errs,
				fmt.Errorf("%w: expected \"<key>:<value>\", got %q", ErrInvalidFormat, def),
			)

			continue
		}

		var fail bool

		key, err := parse(keyZero, keyValue[0], true, true)
		if len(err) > 0 {
			errs = append(errs, addErrorsPrefixes(strconv.Itoa(i), err)...)

			fail = true
		}

		value, err := parse(valueZero, keyValue[1], true, true)
		if len(err) > 0 {
			errs = append(errs, addErrorsPrefixes(strconv.Itoa(i), err)...)

			fail = true
		}

		if fail {
			continue
		}

		result.SetMapIndex(key, value)
	}

	return result, errs
}

// asList converts a comma-separated string to a list. A comma may be escaped with a backslash.
func asList(src string) []string {
	var result []string

	current := ""

	var escaped bool
	for _, char := range src {
		if escaped {
			escaped = false
			current += string(char)

			continue
		}

		if char == '\\' {
			escaped = true

			continue
		}

		if char == ',' {
			result = append(result, current)
			current = ""

			continue
		}

		current += string(char)
	}

	result = append(result, current)

	return result
}
