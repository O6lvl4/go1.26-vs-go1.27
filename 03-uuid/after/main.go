// 1.27: uuid が標準ライブラリ入り (RFC 9562)。依存ゼロ、import は "uuid" 一語。
package main

import (
	"fmt"
	"slices"
	"uuid"
)

func main() {
	u := uuid.New()
	fmt.Println("New:  ", u)

	// V7 は error を返さない。Compare もメソッドとして生えている
	ids := []uuid.UUID{uuid.NewV7(), uuid.NewV7(), uuid.NewV7()}
	slices.SortFunc(ids, uuid.UUID.Compare)
	for _, id := range ids {
		fmt.Println("V7:   ", id)
	}

	p, err := uuid.Parse(u.String())
	fmt.Println("Parse:", p, err == nil)
}
