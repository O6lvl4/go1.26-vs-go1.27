// 〜1.26: struct リテラルのキーはその struct が直接宣言したフィールド名だけ。
// 埋め込みで昇格したフィールドは、読むときは s.ID なのに
// 書くときだけ埋め込み型名を経由する非対称があった。
package main

import "fmt"

type Meta struct{ ID, Owner string }

type Server struct {
	Meta
	Host string
	Port int
}

func main() {
	s := Server{
		Meta: Meta{ID: "srv-1", Owner: "platform"}, // 型名がリテラルに漏れる
		Host: "example.com",
		Port: 8080,
	}
	fmt.Println(s.ID, s.Owner, s.Host, s.Port)

	// 埋め込みを 1 段深くするとリテラルも 1 段深くなる
	type Spec struct{ Server }
	sp := Spec{Server: Server{Meta: Meta{ID: "srv-2"}, Port: 9090}}
	fmt.Println(sp.ID, sp.Port)
}
