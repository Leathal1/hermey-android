// Package models defines data types for the hermes-webui API.
//
// Lenient decoding helpers
// ------------------------
// The upstream server may return fields as different types depending on the
// model version (string vs number, missing vs null, etc.).  Every exported
// model type implements a custom json.Unmarshaler that first tries a strict
// decode and, if that fails, falls back to a field-by-field decode with
// coercion.  Non-essential fields are ignored when malformed instead of
// failing the whole response.
package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FlexString accepts either a JSON string or the raw JSON null value.
type FlexString string

// UnmarshalJSON implements lenient string decoding.
func (f *FlexString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	}
	// Coerce numbers/bools to string.
	*f = FlexString(string(data))
	return nil
}

// FlexInt accepts JSON numbers encoded as numbers or strings.
type FlexInt int64

// UnmarshalJSON implements lenient int decoding.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Try float truncation.
			fv, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			v = int64(fv)
		}
		*f = FlexInt(v)
		return nil
	}
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		var fv float64
		if err2 := json.Unmarshal(data, &fv); err2 != nil {
			return err
		}
		v = int64(fv)
	}
	*f = FlexInt(v)
	return nil
}

// FlexBool accepts booleans or common string representations.
type FlexBool bool

// UnmarshalJSON implements lenient bool decoding.
func (f *FlexBool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = false
		return nil
	}
	val := string(data)
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		val = s
	}
	switch strings.ToLower(val) {
	case "true", "1", "yes", "on":
		*f = true
		return nil
	case "false", "0", "no", "off":
		*f = false
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*f = FlexBool(b)
	return nil
}

// FlexTime accepts RFC3339 strings and numeric Unix timestamps (seconds or millis).
type FlexTime time.Time

// UnmarshalJSON implements lenient time decoding.
func (f *FlexTime) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = FlexTime(time.Time{})
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*f = FlexTime(time.Time{})
			return nil
		}
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02",
		} {
			if t, err := time.Parse(layout, s); err == nil {
				*f = FlexTime(t)
				return nil
			}
		}
		// Try numeric string.
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			if v > 1e12 {
				*f = FlexTime(time.UnixMilli(v))
			} else {
				*f = FlexTime(time.Unix(v, 0))
			}
			return nil
		}
		return fmt.Errorf("unparseable time: %q", s)
	}
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v > 1e12 {
		*f = FlexTime(time.UnixMilli(v))
	} else {
		*f = FlexTime(time.Unix(v, 0))
	}
	return nil
}

// Time returns the underlying time.Time value.
func (f FlexTime) Time() time.Time { return time.Time(f) }

// MarshalJSON serializes FlexTime as RFC3339.
func (f FlexTime) MarshalJSON() ([]byte, error) {
	t := time.Time(f)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

// LenientUnmarshal first tries a strict unmarshal of the target value
// using a type alias to avoid re-entering the target's own UnmarshalJSON.
// If strict unmarshal fails, it falls back to a map-based decode with coercion.
func LenientUnmarshal(data []byte, target interface{}) error {
	if err := unmarshalStrict(target, data); err == nil {
		return nil
	}
	return fallbackUnmarshal(data, target)
}

// unmarshalStrict decodes data into target without invoking any custom
// UnmarshalJSON defined on target's type.
func unmarshalStrict(target interface{}, data []byte) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return json.Unmarshal(data, target)
	}
	// Create an anonymous struct type with the same fields.
	et := v.Elem().Type()
	fields := make([]reflect.StructField, 0, et.NumField())
	for i := 0; i < et.NumField(); i++ {
		f := et.Field(i)
		// Copy field but clear the tag? Keep json tag so standard decoder honors it.
		fields = append(fields, f)
	}
	alias := reflect.StructOf(fields)
	pv := reflect.New(alias)
	if err := json.Unmarshal(data, pv.Interface()); err != nil {
		return err
	}
	v.Elem().Set(pv.Elem())
	return nil
}

func fallbackUnmarshal(data []byte, target interface{}) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	return fillStructFromMap(raw, target)
}

func fillStructFromMap(raw map[string]json.RawMessage, target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		name := jsonName(field)
		data, ok := raw[name]
		if !ok {
			continue
		}
		fv := v.Field(i)
		if err := setField(fv, data); err != nil {
			// Ignore malformed non-critical fields.
			continue
		}
	}
	return nil
}

func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return field.Name
	}
	if i := strings.Index(tag, ","); i >= 0 {
		tag = tag[:i]
	}
	if tag == "" {
		return field.Name
	}
	return tag
}

func setField(fv reflect.Value, data []byte) error {
	if len(data) == 0 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	if !fv.CanSet() {
		return nil
	}

	// If the field itself implements json.Unmarshaler, prefer it.
	if fv.CanAddr() {
		if u, ok := fv.Addr().Interface().(json.Unmarshaler); ok {
			return u.UnmarshalJSON(data)
		}
	}

	switch fv.Kind() {
	case reflect.String:
		var s FlexString
		if err := s.UnmarshalJSON(data); err != nil {
			return err
		}
		fv.SetString(string(s))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x FlexInt
		if err := x.UnmarshalJSON(data); err != nil {
			return err
		}
		fv.SetInt(int64(x))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x FlexInt
		if err := x.UnmarshalJSON(data); err != nil {
			return err
		}
		fv.SetUint(uint64(x))
	case reflect.Bool:
		var b FlexBool
		if err := b.UnmarshalJSON(data); err != nil {
			return err
		}
		fv.SetBool(bool(b))
	case reflect.Float32, reflect.Float64:
		var f float64
		if err := json.Unmarshal(data, &f); err != nil {
			// Try string-encoded float.
			var s string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			f2, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			f = f2
		}
		fv.SetFloat(f)
	case reflect.Struct:
		return LenientUnmarshal(data, fv.Addr().Interface())
	case reflect.Slice:
		return json.Unmarshal(data, fv.Addr().Interface())
	case reflect.Ptr:
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return LenientUnmarshal(data, fv.Interface())
	default:
		return json.Unmarshal(data, fv.Addr().Interface())
	}
	return nil
}

// StringSlice is a slice alias so gomobile can pass arrays of strings.
type StringSlice []string
