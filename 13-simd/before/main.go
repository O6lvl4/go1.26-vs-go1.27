// 〜1.26: SIMD を使いたければ cgo・アセンブリ・unsafe なライブラリのどれか。
// 純 Go で書くならスカラループで、ベクトル化はコンパイラの自動化任せ
// (Go のコンパイラは自動ベクトル化をしない)。
package main

import "fmt"

// y = a*x + y (axpy)
func axpy(a float64, xs, ys []float64) {
	for i := range xs {
		ys[i] = a*xs[i] + ys[i]
	}
}

func main() {
	xs := []float64{1, 2, 3, 4, 5, 6, 7}
	ys := []float64{10, 20, 30, 40, 50, 60, 70}
	axpy(2, xs, ys)
	fmt.Println(ys)
}
