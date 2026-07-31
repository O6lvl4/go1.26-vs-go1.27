// 〜1.26: syscall.Errno は Plan 9 に存在せず、Errno に触れる移植性コードは
// Plan 9 向けビルドだけ壊れるので //go:build !plan9 での隔離が必要だった。
// 1.27: Plan 9 にも Errno 型が定義された (Plan 9 のシステムコールは文字列
// ErrorString を返すので、実際に Errno が返ることはない。ビルドを通すための定義)。
//
// この main.go は build tag なしで Plan 9 向けにもコンパイルが通る:
//
//	GOOS=plan9 GOARCH=amd64 go build ./17-plan9-errno  # 1.27: OK
//	(go1.26.5 実測: undefined: syscall.Errno でビルド不能)
package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func main() {
	_, err := os.Open("/no/such/file")

	// errno を見て分岐する定番の移植性コード。1.26 まではこの 1 行のせいで
	// Plan 9 ビルドが落ちた
	var errno syscall.Errno
	if errors.As(err, &errno) {
		fmt.Printf("errno=%d (%v)\n", uint(errno), error(errno))
	} else {
		fmt.Println("Errno が返らない環境:", err) // Plan 9 では ErrorString が返る
	}
}
