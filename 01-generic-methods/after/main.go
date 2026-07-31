// 1.27: メソッド宣言が独自の型パラメータを持てる。
// 型の操作が型の名前空間に住み、メソッドチェーンで左から右に読める。
package main

import (
	"fmt"
	"strings"
)

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

// 1.26 まではコンパイルエラーだった宣言
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U] {
	out := &Stack[U]{}
	for _, v := range s.items {
		out.Push(f(v))
	}
	return out
}

func (s *Stack[T]) Fold[A any](init A, f func(A, T) A) A {
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

	// チェーンで左から右へ。型引数はリテラルから推論される
	joined := s.
		Map(func(v int) string { return fmt.Sprintf("<%d>", v) }).
		Fold("", func(acc, v string) string { return strings.TrimPrefix(acc+"-"+v, "-") })
	fmt.Println(joined)

	// 制限: interface のメソッドは型パラメータを宣言できない。
	// ジェネリックメソッドで interface を実装することもできない。
}
