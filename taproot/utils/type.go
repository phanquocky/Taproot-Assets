package utils

import "fmt"

func ToPtr[T any](x T) *T {
	return &x
}

func PrintStruct(obj any) {
	fmt.Printf("%+v\n", obj)
}

func ToSliceAny[T any](arr []T) []any {
	docs := make([]any, 0)

	for _, i := range arr {
		docs = append(docs, i)
	}

	return docs
}
