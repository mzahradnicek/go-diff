package godiff

import (
	"bytes"
	"reflect"
	"time"
	"unsafe"
)

type update struct {
	Old, New interface{}
}

type visit struct {
	a1, a2 unsafe.Pointer
	typ    reflect.Type
}

type arlist map[*path]interface{}
type modlist map[*path]update

type diff struct {
	Added    arlist         `json:"added"`
	Removed  arlist         `json:"removed"`
	Modified modlist        `json:"modified"`
	visited  map[visit]bool `json:"-"`
}

func (d *diff) diff(aVal, bVal reflect.Value, p path, opts *opts) bool {
	localPath := make(path, len(p))
	copy(localPath, p)

	if !aVal.IsValid() && !bVal.IsValid() {
		return false
	}

	if !bVal.IsValid() {
		d.Modified[&localPath] = update{Old: aVal.Interface(), New: nil}
		return false
	} else if !aVal.IsValid() {
		d.Modified[&localPath] = update{Old: nil, New: bVal.Interface()}
		return false
	}

	if aVal.Type() != bVal.Type() {
		d.Modified[&localPath] = update{Old: aVal.Interface(), New: bVal.Interface()}
		return false
	}

	kind := aVal.Kind()

	// Borrowed from the reflect package to handle recursive data structures.
	if aVal.CanAddr() && bVal.CanAddr() && hard(kind) {
		addr1 := unsafe.Pointer(aVal.UnsafeAddr())
		addr2 := unsafe.Pointer(bVal.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := aVal.Type()
		v := visit{addr1, addr2, typ}
		if d.visited[v] {
			return true
		}

		// Remember for later.
		d.visited[v] = true
	}
	// End of borrowed code.

	// check for Comparer interface !!!!!!
	avc, aok := aVal.Interface().(Comparer)
	bvc, bok := bVal.Interface().(Comparer)

	if aok && bok {
		if !bytes.Equal(avc.CompareHash(), bvc.CompareHash()) {
			d.Modified[&localPath] = update{Old: aVal.Interface(), New: bVal.Interface()}
			return false
		}
		return true
	}

	equal := true
	switch kind {
	case reflect.Map, reflect.Ptr, reflect.Func, reflect.Chan, reflect.Slice:
		if aVal.IsNil() && bVal.IsNil() {
			return true
		}
		if aVal.IsNil() || bVal.IsNil() {
			d.Modified[&localPath] = update{Old: aVal.Interface(), New: bVal.Interface()}
			return false
		}
	}

	switch kind {
	case reflect.Array, reflect.Slice:
		aLen := aVal.Len()
		bLen := bVal.Len()
		for i := 0; i < min(aLen, bLen); i++ {
			localPath := append(localPath, SliceIndex(i))
			if eq := d.diff(aVal.Index(i), bVal.Index(i), localPath, opts); !eq {
				equal = false
			}
		}
		if aLen > bLen {
			for i := bLen; i < aLen; i++ {
				localPath := append(localPath, SliceIndex(i))
				d.Removed[&localPath] = aVal.Index(i).Interface()
				equal = false
			}
		} else if aLen < bLen {
			for i := aLen; i < bLen; i++ {
				localPath := append(localPath, SliceIndex(i))
				d.Added[&localPath] = bVal.Index(i).Interface()
				equal = false
			}
		}
	case reflect.Map:
		for _, key := range aVal.MapKeys() {
			aI := aVal.MapIndex(key)
			bI := bVal.MapIndex(key)
			localPath := append(localPath, MapKey{key.Interface()})
			if !bI.IsValid() {
				d.Removed[&localPath] = aI.Interface()
				equal = false
			} else if eq := d.diff(aI, bI, localPath, opts); !eq {
				equal = false
			}
		}
		for _, key := range bVal.MapKeys() {
			aI := aVal.MapIndex(key)
			if !aI.IsValid() {
				bI := bVal.MapIndex(key)
				localPath := append(localPath, MapKey{key.Interface()})
				d.Added[&localPath] = bI.Interface()
				equal = false
			}
		}

	case reflect.Struct:
		typ := aVal.Type()
		// If the field is time.Time, use Equal to compare
		if typ.String() == "time.Time" {
			aTime := aVal.Interface().(time.Time)
			bTime := bVal.Interface().(time.Time)
			if !aTime.Equal(bTime) {
				d.Modified[&localPath] = update{Old: aTime.String(), New: bTime.String()}
				equal = false
			}
		} else {
			for i := 0; i < typ.NumField(); i++ {
				index := []int{i}
				field := typ.FieldByIndex(index)

				if field.Tag.Get("testdiff") == "ignore" { // skip fields marked to be ignored
					continue
				}
				if field.PkgPath != "" {
					continue
				}
				if _, skip := opts.ignoreFields[field.Name]; skip {
					continue
				}
				localPath := append(localPath, StructField(field.Name))
				if eq := d.diff(aVal.FieldByIndex(index), bVal.FieldByIndex(index), localPath, opts); !eq {
					equal = false
				}
			}
		}
	case reflect.Ptr:
		equal = d.diff(aVal.Elem(), bVal.Elem(), localPath, opts)
	default:
		if !reflect.DeepEqual(aVal.Interface(), bVal.Interface()) {
			d.Modified[&localPath] = update{Old: aVal.Interface(), New: bVal.Interface()}
			equal = false
		}
	}

	return equal
}

func NewDiff() *diff {
	return &diff{
		Added:    make(arlist),
		Removed:  make(arlist),
		Modified: make(modlist),
		visited:  make(map[visit]bool),
	}
}

func DeepDiff(a, b interface{}, options ...option) (*diff, bool) {
	d := NewDiff()
	opts := &opts{}

	for _, o := range options {
		o.apply(opts)
	}

	return d, d.diff(reflect.ValueOf(a), reflect.ValueOf(b), nil, opts)
}
