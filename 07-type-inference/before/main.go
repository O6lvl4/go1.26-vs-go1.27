// 〜1.26: ジェネリック関数を関数型の値として使うときの型推論は
// 変数への代入・return・関数引数どまり（1.21 で入った範囲）。
// 合成リテラル・チャネル送信・型変換では明示実体化が必要だった。
package main

import "fmt"

func double[T ~int | ~float64](v T) T { return v * 2 }

func main() {
	// ここは 1.21 から推論が効いていた
	var f func(int) int = double
	fmt.Println(f(21))

	// ここから下は [T] を書かないとコンパイルエラーだった
	ops := map[string]func(int) int{"double": double[int]}
	fmt.Println(ops["double"](2))

	ch := make(chan func(float64) float64, 1)
	ch <- double[float64]
	fmt.Println((<-ch)(1.5))

	g := (func(int) int)(double[int])
	fmt.Println(g(4))
}
