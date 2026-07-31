// 1.27: maphash.Hasher[T] が「hash ベースのコンテナと要素の間の契約」として
// 標準化された。ハッシュと等価関係が 1 つの型にまとまり、comparable 型は
// ComparableHasher でそのまま使える。
package main

import (
	"fmt"
	"hash/maphash"
	"strings"
)

type Bloom[T any] struct {
	seed   maphash.Seed
	bits   [4096]bool
	hasher maphash.Hasher[T] // 標準の契約を受け取るだけ
}

func NewBloom[T any](h maphash.Hasher[T]) *Bloom[T] {
	return &Bloom[T]{seed: maphash.MakeSeed(), hasher: h}
}

func (b *Bloom[T]) probe(v T, i byte) uint64 {
	var h maphash.Hash
	h.SetSeed(b.seed)
	h.WriteByte(i)
	b.hasher.Hash(&h, v)
	return h.Sum64() % uint64(len(b.bits))
}

func (b *Bloom[T]) Add(v T) {
	b.bits[b.probe(v, 0)] = true
	b.bits[b.probe(v, 1)] = true
}

func (b *Bloom[T]) Contains(v T) bool {
	return b.bits[b.probe(v, 0)] && b.bits[b.probe(v, 1)]
}

// 独自の等価関係も Hasher として一級市民。Hash と Equal の整合が型で束ねられる
type CaseInsensitive struct{}

func (CaseInsensitive) Hash(h *maphash.Hash, s string) { h.WriteString(strings.ToLower(s)) }
func (CaseInsensitive) Equal(x, y string) bool         { return strings.ToLower(x) == strings.ToLower(y) }

func main() {
	// comparable 型は ComparableHasher で無料 (== と整合するハッシュ)
	names := NewBloom[string](maphash.ComparableHasher[string]{})
	names.Add("Gopher")
	fmt.Println(names.Contains("Gopher"), names.Contains("gopher"))

	loose := NewBloom[string](CaseInsensitive{})
	loose.Add("Gopher")
	fmt.Println(loose.Contains("Gopher"), loose.Contains("gopher"))
}
