package godiff

import "reflect"

type Comparer interface {
	CompareHash() []byte
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func hard(k reflect.Kind) bool {
	switch k {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return true
	}
	return false
}
