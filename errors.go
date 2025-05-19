package defaults

import "errors"

type fieldError struct {
	err   error
	field string
}

var (
	// ErrInvalidFormat is returned when the default value is not valid for the target type.
	ErrInvalidFormat = errors.New("invalid format for default value")

	// ErrUnsupportedType is returned when the target type is not supported.
	ErrUnsupportedType = errors.New("unsupported type for defaults")

	// ErrMustBePointerToAStruct is returned when the target is not a pointer to a struct.
	ErrMustBePointerToAStruct = errors.New("target must be a pointer to a struct")
)

func (e fieldError) Error() string {
	return e.field + ": " + e.err.Error()
}

func (e fieldError) Unwrap() error {
	return e.err
}

func addErrorsPrefixes(prefix string, errs []error) []error {
	for i, err := range errs {
		var fErr fieldError
		if errors.As(err, &fErr) {
			fErr.field = prefix + "." + fErr.field
			errs[i] = fErr

			continue
		}

		errs[i] = fieldError{field: prefix, err: err}
	}

	return errs
}
