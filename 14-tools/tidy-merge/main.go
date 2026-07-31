// go mod tidy の require ブロック統合デモ用 fixture。
// 実行は ../demo.sh 経由 (go.mod を一時ディレクトリにコピーして tidy し、diff を表示)。
package main

import (
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/google/uuid"
)

func main() {
	fmt.Println(uuid.Max, mldsa65.SignatureSize)
}
