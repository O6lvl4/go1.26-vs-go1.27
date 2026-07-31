// 1.27: encoding/json/v2 が正式入り。デフォルトが厳格に反転し、
// 緩めたい場合だけ Options でオプトインする。
package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

type Event struct {
	Name  string `json:"name"`
	Venue string `json:"venue,omitzero"` // omitempty より直感的な omitzero
}

func main() {
	// 重複キーはデフォルトでエラー
	var m map[string]int
	err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &m)
	fmt.Println("duplicate keys:", err)

	// v1 の挙動が必要ならオプトインで戻せる
	_ = json.Unmarshal([]byte(`{"a":1,"a":2}`), &m, jsontext.AllowDuplicateNames(true))
	fmt.Println("opt-in lenient:", m)

	// フィールド名は exact match がデフォルト
	var ev Event
	_ = json.Unmarshal([]byte(`{"NAME":"x"}`), &ev)
	fmt.Println("exact match (NAME ignored):", ev.Name == "")
	_ = json.Unmarshal([]byte(`{"NAME":"x"}`), &ev, json.MatchCaseInsensitiveNames(true))
	fmt.Println("opt-in case-insensitive:", ev.Name)

	// 不正な UTF-8 もデフォルトでエラー。黙ってデータが化けない
	var s string
	err = json.Unmarshal([]byte("\"a\xffb\""), &s)
	fmt.Println("invalid UTF-8:", err)
	_ = json.Unmarshal([]byte("\"a\xffb\""), &s, jsontext.AllowInvalidUTF8(true))
	fmt.Printf("opt-in lenient: %q\n", s)

	// 整形も Options として渡す
	out, _ := json.Marshal(Event{Name: "miyazaki.go"}, jsontext.WithIndent("  "))
	fmt.Println(string(out))
}
