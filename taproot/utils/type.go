package utils

import "fmt"

// ToPtr return pointer of any type.
func ToPtr[T any](x T) *T {
	return &x
}

func PrintStruct(obj any) {
	fmt.Printf("%+v\n", obj)
}
