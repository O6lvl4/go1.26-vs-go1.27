// 〜1.26 の細かい不便たち。
package main

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
)

// CutLast がないので LastIndex + スライスを毎回手書き
func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}

func main() {
	dir, file, _ := cutLast("archive/2026/photo.tar.gz", "/")
	base, ext, _ := cutLast(file, ".")
	fmt.Println(dir, "|", base, "|", ext)

	// Rand は型ごとに別メソッド。型が変わるたびに呼び分ける
	r := rand.New(rand.NewPCG(1, 2))
	fmt.Println(r.IntN(10), r.Int64N(100), uint8(r.Uint64N(50)))

	// URL の複製は手書きシャローコピー。*Userinfo が共有される罠つき
	u, _ := url.Parse("https://user:pass@example.com/path?tab=1")
	v := *u
	v.RawQuery = "tab=2"
	fmt.Println("User pointer shared:", u.User == v.User) // true — 深い複製は自己責任だった
	fmt.Println(u.String())
	fmt.Println(v.String())
}
