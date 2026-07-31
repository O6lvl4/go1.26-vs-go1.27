// 〜1.26: メソッドは型パラメータを持てないので、package スコープの
// ジェネリック関数として書くしかなかった。
package main

import (
	"fmt"
	"strings"
)

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

// 型の操作なのに package の名前空間を占有する。
// Stack 用の Map と Tree 用の Map は同じ package に共存できない。
func Map[T, U any](s *Stack[T], f func(T) U) *Stack[U] {
	out := &Stack[U]{}
	for _, v := range s.items {
		out.Push(f(v))
	}
	return out
}

func Fold[T, A any](s *Stack[T], init A, f func(A, T) A) A {
	acc := init
	for _, v := range s.items {
		acc = f(acc, v)
	}
	return acc
}

func main() {
	s := &Stack[int]{}
	s.Push(1)
	s.Push(2)
	s.Push(3)

	// 関数形式なのでチェーンできない。内側から外側へ読む
	joined := Fold(
		Map(s, func(v int) string { return fmt.Sprintf("<%d>", v) }),
		"",
		func(acc, v string) string { return strings.TrimPrefix(acc+"-"+v, "-") },
	)
	fmt.Println(joined)
}
