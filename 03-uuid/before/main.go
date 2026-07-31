// 〜1.26: UUID は標準にないので、事実上の標準だった github.com/google/uuid に
// 依存を張るのが定石だった。
package main

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

func main() {
	u := uuid.New()
	fmt.Println("New:  ", u)

	// V7 は error を返すシグネチャ
	ids := make([]uuid.UUID, 0, 3)
	for range 3 {
		v7, err := uuid.NewV7()
		if err != nil {
			panic(err)
		}
		ids = append(ids, v7)
	}
	// Compare メソッドがないので bytes.Compare を自分で書く
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
	for _, id := range ids {
		fmt.Println("V7:   ", id)
	}

	p, err := uuid.Parse(u.String())
	fmt.Println("Parse:", p, err == nil)
}
