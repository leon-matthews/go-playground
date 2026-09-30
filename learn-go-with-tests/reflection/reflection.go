package main

import (
	"reflect"
)

// walk calls fn for all string fiels inside x
func walk(x any, fn func(input string)) {
	val := getValue(x)
	walkValue := func(value reflect.Value) {
		walk(value.Interface(), fn)
	}

	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			walkValue(val.Index(i))
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			walkValue(val.MapIndex(key))
		}
	case reflect.String:
		fn(val.String())
	case reflect.Struct:
		for _, v := range val.Fields() {
			walkValue(v)
		}
	}
}

// getValue automatically dereferences pointer values
func getValue(x any) reflect.Value {
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	return val
}
