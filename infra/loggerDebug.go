package infra

import (
	"fmt"
	"math"
	"reflect"
)

func checkForNaN(name string, val interface{}) {
	v := reflect.ValueOf(val)
	checkRecursive(name, v)
}

func checkRecursive(path string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			checkRecursive(path, v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldName := field.Name
			checkRecursive(path+"."+fieldName, v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			checkRecursive(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			checkRecursive(fmt.Sprintf("%s[%v]", path, key.Interface()), v.MapIndex(key))
		}
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) {
			fmt.Printf("NaN detected at path: %s\n", path)
		}
	}
}
