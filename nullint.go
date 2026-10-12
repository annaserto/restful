// Copyright 2021-2025 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package restful

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// rejectNullIntegers reports when body assigns JSON null to a non-nullable integer of dest.
// Invalid JSON is ignored so the normal decoder reports it.
// A field tagged `nullable:"true"` may be null.
func rejectNullIntegers(body []byte, dest any) error {
	if len(bytes.TrimSpace(body)) == 0 || dest == nil {
		return nil
	}
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	return walkNullIntegers(reflect.TypeOf(dest), raw, "")
}

func walkNullIntegers(t reflect.Type, v any, path string) error {
	t = derefType(t)
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	for name, field := range jsonFieldIndex(t) {
		val, present := obj[name]
		if !present {
			continue
		}
		fieldPath := name
		if path != "" {
			fieldPath = path + "." + name
		}
		if val == nil {
			if isNonNullableInteger(field) {
				return fmt.Errorf("invalid null for integer field %s", fieldPath)
			}
			continue
		}
		ft := derefType(field.Type)
		switch ft.Kind() {
		case reflect.Struct:
			if err := walkNullIntegers(ft, val, fieldPath); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			items, ok := val.([]any)
			if !ok {
				continue
			}
			elem := derefType(ft.Elem())
			for _, item := range items {
				if err := walkNullIntegers(elem, item, fieldPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isNonNullableInteger(f reflect.StructField) bool {
	if f.Tag.Get("nullable") == "true" {
		return false
	}
	switch derefType(f.Type).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func jsonFieldIndex(t reflect.Type) map[string]reflect.StructField {
	fields := make(map[string]reflect.StructField)
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		t = derefType(t)
		if t == nil || t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if f.Anonymous {
				walk(f.Type)
				continue
			}
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name := f.Name
			if tag != "" {
				name, _, _ = strings.Cut(tag, ",")
				if name == "" {
					name = f.Name
				}
			}
			if _, exists := fields[name]; !exists {
				fields[name] = f
			}
		}
	}
	walk(t)
	return fields
}
