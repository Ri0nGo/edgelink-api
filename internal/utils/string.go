package utils

import (
	"math/rand"
	"strings"
	"time"
)

// JoinByFunc 泛型函数，将 slice 中的元素通过 fn 映射成 string，再用 sep 拼接
func JoinByFunc[T any](elems []T, sep string, fn func(T) string) string {
	var strs = make([]string, len(elems))
	for idx, elem := range elems {
		strs[idx] = fn(elem)
	}
	return strings.Join(strs, sep)
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandString 返回指定程度的随机字符串
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
