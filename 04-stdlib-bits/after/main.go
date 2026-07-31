// 1.27: 手書きしていた 3 つがそのまま標準になった。
package main

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
)

func main() {
	dir, file, _ := strings.CutLast("archive/2026/photo.tar.gz", "/")
	base, ext, _ := strings.CutLast(file, ".")
	fmt.Println(dir, "|", base, "|", ext)

	// Rand.N は 1 つのジェネリックメソッドで全整数型をカバー。
	// これ自体が 01 のジェネリックメソッドで初めて書けるようになった API
	r := rand.New(rand.NewPCG(1, 2))
	fmt.Println(r.N(10), r.N(int64(100)), r.N(uint8(50)))

	// URL.Clone は deep copy と明記されている
	u, _ := url.Parse("https://user:pass@example.com/path?tab=1")
	v := u.Clone()
	v.RawQuery = "tab=2"
	fmt.Println("User pointer shared:", u.User == v.User) // false
	fmt.Println(u.String())
	fmt.Println(v.String())
}
