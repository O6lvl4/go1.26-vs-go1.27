// 〜1.26: goroutine リークの調査は "goroutine" プロファイルを目視するしかなかった。
// 健全な goroutine もリークも全部混ざって出てくるので、どれがリークかは人間が判断する。
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

func leak() {
	ch := make(chan int) // 誰も送信しないチャネル
	go func() { <-ch }() // この goroutine は永遠に回収されない
}

func main() {
	for range 3 {
		leak()
	}
	time.Sleep(100 * time.Millisecond)

	fmt.Println("=== goroutine profile (全部入り。リークの特定は目視) ===")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
}
