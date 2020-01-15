package godiff

import "fmt"

type path []PathNode

func (p path) String() (res string) {
	for _, n := range p {
		res += n.String()
	}
	return
}

type PathNode interface {
	String() string
}

type StructField string

func (n StructField) String() string {
	return string(n) //fmt.Sprintf("%s", n)
}

type MapKey struct {
	Key interface{}
}

func (n MapKey) String() string {
	return fmt.Sprintf("[%#v]", n)
}

type SliceIndex int

func (n SliceIndex) String() string {
	return fmt.Sprintf("[%d]", n)
}
