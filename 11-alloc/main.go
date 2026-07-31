// 1.27: コンパイラがサイズ特化した malloc を呼ぶようになり、
// 80 byte 未満の小さいアロケーションが最大 30% 速くなった。
// コードは 1 文字も変わらないので、before/after はビルドフラグで再現する:
//
//	GOEXPERIMENT=nosizespecializedmalloc go run ./11-alloc  # 1.26 相当
//	go run ./11-alloc                                       # 1.27 デフォルト
package main

import (
	"fmt"
	"testing"
)

type node struct {
	key, val int64
	next     *node
}

var sink *node

func main() {
	r := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			sink = &node{key: 1}
		}
	})
	fmt.Printf("small alloc (24B): %s %s\n", r.String(), r.MemString())
}
