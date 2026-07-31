// 1.27: 手書きしていた 3 つがそのまま標準になった。
package main

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand/v2"
	"net/url"
	"strings"
)

func main() {
	dir, file, _ := strings.CutLast("archive/2026/photo.tar.gz", "/")
	base, ext, _ := strings.CutLast(file, ".")
	fmt.Println(dir, "|", base, "|", ext)

	// bytes 側にも同じものが入った
	host, port, _ := bytes.CutLast([]byte("[::1]:8080"), []byte(":"))
	fmt.Printf("%s | %s\n", host, port)

	// Rand.N は 1 つのジェネリックメソッドで全整数型をカバー。
	// これ自体が 01 のジェネリックメソッドで初めて書けるようになった API
	r := rand.New(rand.NewPCG(1, 2))
	fmt.Println(r.N(10), r.N(int64(100)), r.N(uint8(50)))

	// Divide は丸めモード (Trunc/Floor/Round/Ceil) を指定して一発
	quo, rem := new(big.Int).Divide(big.NewInt(-7), big.NewInt(2), new(big.Int), big.Floor)
	fmt.Println("floor(-7/2):", quo, rem)

	// URL.Clone は deep copy と明記されている
	u, _ := url.Parse("https://user:pass@example.com/path?tab=1")
	v := u.Clone()
	v.RawQuery = "tab=2"
	fmt.Println("User pointer shared:", u.User == v.User) // false
	fmt.Println(u.String())
	fmt.Println(v.String())
}
