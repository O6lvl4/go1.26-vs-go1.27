// 1.27: goroutineleak プロファイルが追加。実行可能な goroutine から到達不能なまま
// ブロックしている goroutine だけを報告してくれる。
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

func leak() {
	ch := make(chan int)
	go func() { <-ch }()
}

func main() {
	for range 3 {
		leak()
	}
	time.Sleep(100 * time.Millisecond)

	fmt.Println("=== goroutineleak profile (リークだけが出る) ===")
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
