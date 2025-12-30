package utils

import "strings"

// JoinByFunc 泛型函数，将 slice 中的元素通过 fn 映射成 string，再用 sep 拼接
func JoinByFunc[T any](elems []T, sep string, fn func(T) string) string {
	var strs = make([]string, len(elems))
	for idx, elem := range elems {
		strs[idx] = fn(elem)
	}
	return strings.Join(strs, sep)
}
