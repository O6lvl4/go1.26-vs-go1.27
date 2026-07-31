// 1.27: struct リテラルのキーに有効なフィールドセレクタが書けるようになった。
// 埋め込みで昇格したフィールドを、読むときと同じ名前でそのまま初期化できる。
package main

import "fmt"

type Meta struct{ ID, Owner string }

type Server struct {
	Meta
	Host string
	Port int
}

func main() {
	// 1.26 までは unknown field X in struct literal エラーだった書き方
	s := Server{
		ID:    "srv-1",
		Owner: "platform",
		Host:  "example.com",
		Port:  8080,
	}
	fmt.Println(s.ID, s.Owner, s.Host, s.Port)

	// 昇格が何段深くても、セレクタとして有効なら書ける
	type Spec struct{ Server }
	sp := Spec{ID: "srv-2", Port: 9090}
	fmt.Println(sp.ID, sp.Port)
}
