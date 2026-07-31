// 1.27 (実験的): ベクトル幅非依存・ポータブルな simd パッケージが追加。
// arm64 Neon / amd64 AVX / wasm で同じコードがハードウェア SIMD になる。
// ビルドには GOEXPERIMENT=simd が必要:
//
//	GOEXPERIMENT=simd go run ./13-simd/after

//go:build goexperiment.simd

package main

import (
	"fmt"
	"simd"
)

// y = a*x + y (axpy)。端数は LoadPart/StorePart がゼロ埋めで面倒を見る
func axpy(a float64, xs, ys []float64) {
	av := simd.BroadcastFloat64s(a)
	lanes := av.Len()
	for i := 0; i < len(xs); i += lanes {
		x, _ := simd.LoadFloat64sPart(xs[i:])
		y, _ := simd.LoadFloat64sPart(ys[i:])
		av.MulAdd(x, y).StorePart(ys[i:])
	}
}

func main() {
	xs := []float64{1, 2, 3, 4, 5, 6, 7}
	ys := []float64{10, 20, 30, 40, 50, 60, 70}
	axpy(2, xs, ys)
	fmt.Println(ys)
	fmt.Println("lanes:", simd.BroadcastFloat64s(0).Len(), "emulated:", simd.Emulated())
}
