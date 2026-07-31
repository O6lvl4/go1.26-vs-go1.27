// 〜1.26: 耐量子署名 (ML-DSA / FIPS 204) を使うなら cloudflare/circl 等の
// サードパーティに依存を張るしかなかった。03-uuid と同じ構図。
package main

import (
	"crypto/rand"
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65" // go.mod に require が増える
)

func main() {
	pub, priv, err := mldsa65.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	msg := []byte("miyazaki.go #5")
	sig := make([]byte, mldsa65.SignatureSize)
	if err := mldsa65.SignTo(priv, msg, nil, true, sig); err != nil {
		panic(err)
	}

	fmt.Println("sig bytes:", len(sig))
	fmt.Println("verify:   ", mldsa65.Verify(pub, msg, nil, sig))
	fmt.Println("tampered: ", mldsa65.Verify(pub, []byte("tampered"), nil, sig))
}
