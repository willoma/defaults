/*
Package defaults provides a parser that extracts default values from the "default" struct tags
and unmarshals them into the struct.

The following types, as well as types that derive from them, can be directly parsed:

  - booleans: parsed with [strconv.ParseBool]
  - integers: parsed with [strconv.Atoi] or [strconv.ParseInt]
  - unsigned integers: parsed with [strconv.ParseUint]
  - floats: parsed with [strconv.ParseFloat]
  - complex numbers: parsed with [strconv.ParseComplex]
  - strings: copied as-is

When a type implements [encoding.TextUnmarshaler],
the default value is parsed by calling
its [encoding.TextUnmarshaler.UnmarshalText] method.

Among many other, the following well-known types
implement the [encoding.TextUnmarshaler] interface:

  - [big.Float]
  - [big.Int]
  - [big.Rat]
  - [net.IP]
  - [netip.Addr]
  - [netip.AddrPort]
  - [netip.Prefix]
  - [regexp.Regexp]
  - [slog.Level]
  - [slog.LevelVar]
  - [github.com/google/uuid.UUID]

The following types have specific parsers:

  - for bidirectional channels, the default value may be a buffer size,
    in which case the channel is created with this buffer size
    (if the value is 0 or negative, the channel is unbuffered).
  - [fs.FileMode] is parsed with [strconv.ParseUint]
    (in the octal notation, for example "644") or,
    if that fails, by trying to read permissions in the "rwxrwxrwx" format
    (for example "rw-r--r--").
  - [net.HardwareAddr] is parsed with [net.ParseMAC]
  - [time.Duration] is parsed with [time.ParseDuration]
  - [time.Time] is parsed with [time.Parse] with the following formats:
    [time.RFC3339] ("YYYY-MM-DDTHH:MM:SSZTZ"),
    [time.RFC3339Nano] ("YYYY-MM-DDTHH:MM:SS.NSZTZ"),
    [time.DateTime] ("YYYY-MM-DD HH:MM:SS"),
    [time.DateOnly] ("YYYY-MM-DD"),
    [time.TimeOnly] ("HH:MM:SS"),
    or "HH:MM"

Slices and arrays of all these types are supported,
with defaults values separated by commas
(for instance "first element,second,third\\, with a comma"
for "first element", "second" and "third, with a comma").

Maps of these types are supported,
with defaults values separated by commas,
each value being a key-value pair separated by a colon
(for instance "one:1,two:2,three:3" for {"one": 1, "two": 2, "three": 3}).

Defaults are unsupported for the following types:

  - uintptrs
  - unidirectional channels
  - functions
  - interfaces
  - unsafe pointers

Example:

	type gender int

	const (
		nonbinary gender = 0
		woman     gender = 1
		man       gender = 2
	)

	type wonderfulPerson struct {
		Name     string    `yaml:"name"     default:"Willow"`
		Birth    time.Time `yaml:"birth"    default:"1982-04-12T23:20:00+02:00"`
		Gender   gender    `yaml:"gender"   default:"1"`
		Passions []string  `yaml:"passions" default:"IT,dancing,videogames"`
	}
*/
package defaults

import (
	"errors"
	"reflect"
)

// Initialize unmarshals the tagged defaults and applies them, overwriting existing values.
func Initialize(target any) error { return apply(target, true) }

// Complete unmarshals the tagged defaults and applies them to unset values, leaving non-zero values untouched.
func Complete(target any) error { return apply(target, false) }

func apply(target any, overwrite bool) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return ErrMustBePointerToAStruct
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrMustBePointerToAStruct
	}

	_, errs := parseStruct(elem, overwrite)

	return errors.Join(errs...)
}
