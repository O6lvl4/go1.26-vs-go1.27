// 〜1.26: hash ベースの自作コンテナ (hash table, Bloom filter, ...) に要素を
// 入れるための標準の契約がなく、コンテナごとにハッシュ関数の受け渡し方を
// 発明していた。関数値を渡すのが定番で、等価関係が要るならもう 1 本増える。
package main

import (
	"fmt"
	"hash/maphash"
	"strings"
)

type Bloom[T any] struct {
	seeds [2]maphash.Seed
	bits  [4096]bool
	hash  func(maphash.Seed, T) uint64 // このコンテナ独自の契約
}

func NewBloom[T any](hash func(maphash.Seed, T) uint64) *Bloom[T] {
	return &Bloom[T]{
		seeds: [2]maphash.Seed{maphash.MakeSeed(), maphash.MakeSeed()},
		hash:  hash,
	}
}

func (b *Bloom[T]) Add(v T) {
	for _, s := range b.seeds {
		b.bits[b.hash(s, v)%uint64(len(b.bits))] = true
	}
}

func (b *Bloom[T]) Contains(v T) bool {
	for _, s := range b.seeds {
		if !b.bits[b.hash(s, v)%uint64(len(b.bits))] {
			return false
		}
	}
	return true
}

func main() {
	// string 用のハッシュ関数を書いて渡す
	names := NewBloom(func(seed maphash.Seed, s string) uint64 {
		return maphash.String(seed, s)
	})
	names.Add("Gopher")
	fmt.Println(names.Contains("Gopher"), names.Contains("gopher"))

	// case-insensitive にしたければ正規化入りの別関数。
	// 「ハッシュと等価関係はセット」という不変条件は書き手のマナー頼み
	loose := NewBloom(func(seed maphash.Seed, s string) uint64 {
		return maphash.String(seed, strings.ToLower(s))
	})
	loose.Add("Gopher")
	fmt.Println(loose.Contains("Gopher"), loose.Contains("gopher"))
}
