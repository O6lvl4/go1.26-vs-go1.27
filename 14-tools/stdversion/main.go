// go.mod は go 1.23 なのに 1.24 で入った strings.SplitSeq を使っている fixture。
// 1.27 からは go test が stdversion vet check をデフォルトで実行するので、
// この「新しすぎる標準ライブラリの使用」がテスト時に検出される:
//
//	cd 14-tools/stdversion && go test .
//	→ strings.SplitSeq requires go1.24 or later (module is go1.23)
//
// (リリース版 1.27 なら「go 1.26 の module で strings.CutLast」も同様に検出される)
package main

import (
	"fmt"
	"strings"
)

func main() {
	for part := range strings.SplitSeq("a,b,c", ",") {
		fmt.Println(part)
	}
}
