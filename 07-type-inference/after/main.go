// 1.27: 関数型推論が一般化され、ジェネリック関数を「一致する関数型」に
// 代入・変換するすべての文脈で型引数が推論されるようになった。
package main

import "fmt"

func double[T ~int | ~float64](v T) T { return v * 2 }

func main() {
	var f func(int) int = double
	fmt.Println(f(21))

	// 1.26 までは cannot use generic function double without instantiation だった
	ops := map[string]func(int) int{"double": double}
	fmt.Println(ops["double"](2))

	ch := make(chan func(float64) float64, 1)
	ch <- double
	fmt.Println((<-ch)(1.5))

	g := (func(int) int)(double)
	fmt.Println(g(4))
}
