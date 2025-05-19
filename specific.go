package defaults

import (
	"fmt"
	"io/fs"
	"net"
	"reflect"
	"strconv"
	"time"
)

func parseSpecific(
	target reflect.Value, value string, hasDefault bool,
) (result reflect.Value, hasParser bool, errs []error) {
	switch target.Type() {
	case reflect.TypeOf(fs.FileMode(0)):
		hasParser = true

		if hasDefault {
			var mode fs.FileMode
			mode, errs = parseFileMode(value)
			result = reflect.ValueOf(mode)
		}

	case reflect.TypeOf(net.HardwareAddr{}):
		hasParser = true

		if hasDefault {
			mac, err := net.ParseMAC(value)
			if err != nil {
				errs = []error{err}
			}

			result = reflect.ValueOf(mac)
		}

	case reflect.TypeOf(time.Duration(0)):
		hasParser = true

		if hasDefault {
			dur, err := time.ParseDuration(value)
			if err != nil {
				errs = []error{err}
			}

			result = reflect.ValueOf(dur)
		}

	case reflect.TypeOf(time.Time{}):
		hasParser = true

		if hasDefault {
			var stamp time.Time

			stamp, errs = parseTime(value)
			result = reflect.ValueOf(stamp)
		}
	}

	return result, hasParser, errs
}

func parseFileMode(value string) (fs.FileMode, []error) {
	uintMode, uintErr := strconv.ParseUint(value, 8, 32)
	if uintErr == nil {
		return fs.FileMode(uintMode), nil
	}

	const rwxFormatLength = 9

	if len(value) != rwxFormatLength {
		return 0, []error{
			uintErr,
			fmt.Errorf("%w: rwxrwxrwx expects %d characters, got %d", ErrInvalidFormat, rwxFormatLength, len(value)),
		}
	}

	return parseFileModeRwxrwxrwx(value), nil
}

func parseFileModeRwxrwxrwx(value string) fs.FileMode {
	var mode fs.FileMode

	const (
		read    = 4
		write   = 2
		execute = 1
		user    = 6
		group   = 3
		other   = 0
	)

	if value[0] == 'r' {
		mode |= (read << user)
	}

	if value[1] == 'w' {
		mode |= (write << user)
	}

	if value[2] == 'x' {
		mode |= (execute << user)
	}

	if value[3] == 'r' {
		mode |= (read << group)
	}

	if value[4] == 'w' {
		mode |= (write << group)
	}

	if value[5] == 'x' {
		mode |= (execute << group)
	}

	if value[6] == 'r' {
		mode |= (read << other)
	}

	if value[7] == 'w' {
		mode |= (write << other)
	}

	if value[8] == 'x' {
		mode |= (execute << other)
	}

	return mode
}

// parseTime parses a string value as a [time.Time], supporting the following formats:
//
//   - [time.RFC3339] ("YYYY-MM-DDTHH:MM:SSZTZ")
//   - [time.RFC3339Nano] ("YYYY-MM-DDTHH:MM:SS.NSZTZ")
//   - [time.DateTime] ("YYYY-MM-DD HH:MM:SS")
//   - [time.DateOnly] ("YYYY-MM-DD")
//   - [time.TimeOnly] ("HH:MM:SS")
//   - "HH:MM"
//
// If none of the formats match, all parsing errors are joined and returned.
func parseTime(value string) (time.Time, []error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}

	errs := []error{err}

	parsed, err = time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, nil
	}

	errs = append(errs, err)

	parsed, err = time.Parse(time.DateTime, value)
	if err == nil {
		return parsed, nil
	}

	errs = append(errs, err)

	parsed, err = time.Parse(time.DateOnly, value)
	if err == nil {
		return parsed, nil
	}

	errs = append(errs, err)

	parsed, err = time.Parse(time.TimeOnly, value)
	if err == nil {
		return parsed, nil
	}

	errs = append(errs, err)

	parsed, err = time.Parse("15:04", value)
	if err == nil {
		return parsed, nil
	}

	errs = append(errs, err)

	return time.Time{}, errs
}
