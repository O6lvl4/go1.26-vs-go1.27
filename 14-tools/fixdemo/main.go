// go fix の 1.27 新 modernizer (atomictypes / embedlit / slicesbackward /
// unsafefuncs) が書き換える「古い書き方」を意図的に残した fixture。
//
//	go fix -diff ./14-tools/fixdemo   # 提案パッチを表示 (適用はしない)
package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Meta struct{ ID string }

type Server struct {
	Meta
	Port int
}

func main() {
	// atomictypes: 生の int64 + atomic 関数 → atomic.Int64
	var counter int64
	atomic.AddInt64(&counter, 1)
	fmt.Println(atomic.LoadInt64(&counter))

	// embedlit: 埋め込み型名経由のリテラル → 昇格フィールド直書き (06 の言語変更が前提)
	s := Server{Meta: Meta{ID: "srv-1"}, Port: 8080}
	fmt.Println(s.ID)

	// slicesbackward: 手書きの逆順ループ → slices.Backward
	xs := []int{1, 2, 3}
	for i := len(xs) - 1; i >= 0; i-- {
		fmt.Println(xs[i])
	}

	// unsafefuncs: ポインタ演算 → unsafe.Add
	arr := [4]int64{10, 20, 30, 40}
	p := unsafe.Pointer(uintptr(unsafe.Pointer(&arr[0])) + 8)
	fmt.Println(*(*int64)(p))
}
