// Package goenv provides a simple way to load environment variables into a struct.
package goenv

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	durationType       = reflect.TypeOf(time.Duration(0))
	timeType           = reflect.TypeOf(time.Time{})
	textUnmarshalerTyp = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

type Options struct {
	// DotEnvPath is an optional development-only fallback file.
	// Runtime environment variables always take precedence.
	DotEnvPath string
}

// Load populates cfg from environment variables and optional dotenv fallback values.
func Load(cfg any, opts Options) error {
	var dotEnv map[string]string

	if opts.DotEnvPath != "" {
		parsed, err := parseDotEnv(opts.DotEnvPath)
		if err != nil {
			return err
		}
		dotEnv = parsed
	}

	v := reflect.ValueOf(cfg)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return errors.New("cfg must be pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if fieldType.PkgPath != "" {
			continue // unexported field
		}

		key := fieldType.Tag.Get("env")
		if key == "" {
			continue
		}

		value, ok := resolveValue(key, fieldType, dotEnv)

		required := tagTruthy(fieldType, "required")
		allowEmpty := tagTruthy(fieldType, "allow_empty", "allowempty")

		if !ok && required {
			return fmt.Errorf("missing required env: %s", key)
		}
		if required && value == "" && !allowEmpty {
			return fmt.Errorf("required env is empty: %s", key)
		}

		if !ok {
			continue
		}

		if err := setField(field, value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	return nil
}

func tagTruthy(fieldType reflect.StructField, keys ...string) bool {
	for _, key := range keys {
		if strings.EqualFold(fieldType.Tag.Get(key), "true") {
			return true
		}
	}

	return false
}

func resolveValue(key string, fieldType reflect.StructField, dotEnv map[string]string) (string, bool) {
	if value, ok := os.LookupEnv(key); ok {
		return value, true
	}

	if dotEnv != nil {
		if value, ok := dotEnv[key]; ok {
			return value, true
		}
	}

	if value, ok := fieldType.Tag.Lookup("default"); ok {
		return value, true
	}

	return "", false
}

func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch {
	case field.Type() == durationType:
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(d))
		return nil
	case field.Type() == timeType:
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			t, err = time.Parse(time.RFC3339, value)
			if err != nil {
				return err
			}
		}
		field.Set(reflect.ValueOf(t))
		return nil
	case field.CanAddr() && field.Addr().Type().Implements(textUnmarshalerTyp):
		return field.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
	case field.Type() == reflect.TypeOf([]byte(nil)):
		field.SetBytes([]byte(value))
		return nil
	case field.Kind() == reflect.Slice:
		return setSlice(field, value)
	}

	return setScalar(field, value)
}

func setSlice(field reflect.Value, value string) error {
	if strings.TrimSpace(value) == "" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	parts := strings.Split(value, ",")
	slice := reflect.MakeSlice(field.Type(), 0, len(parts))

	for _, part := range parts {
		elem := reflect.New(field.Type().Elem()).Elem()
		if err := setField(elem, strings.TrimSpace(part)); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}

	field.Set(slice)
	return nil
}

func setScalar(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}
