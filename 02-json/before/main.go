// 〜1.26 の encoding/json (v1): デフォルトが寛容。
// このコードは 1.27 でも同じ出力になる — encoding/json の内部は v2 実装に
// 置き換わったが、v1 API の挙動は互換維持されている。
package main

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Name string `json:"name"`
}

func main() {
	// 重複キーは黙って後勝ち。攻撃ベクタとして有名 (JSON interoperability 系 CVE の温床)
	var m map[string]int
	err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &m)
	fmt.Println("duplicate keys:", m, err)

	// フィールド名は case-insensitive にマッチしてしまう
	var ev Event
	_ = json.Unmarshal([]byte(`{"NAME":"x"}`), &ev)
	fmt.Println("case-insensitive match:", ev.Name)

	// 不正な UTF-8 は黙って U+FFFD に化ける。データ破壊がエラーにならない
	var s string
	err = json.Unmarshal([]byte("\"a\xffb\""), &s)
	fmt.Printf("invalid UTF-8: %q %v\n", s, err)

	// 整形はエンコーダの設定ではなく別関数 MarshalIndent
	out, _ := json.MarshalIndent(Event{Name: "miyazaki.go"}, "", "  ")
	fmt.Println(string(out))
}
