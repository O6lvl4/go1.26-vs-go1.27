// 1.27: crypto/mldsa が標準入り。crypto.Signer を実装し、
// x509/TLS 1.3 からもそのまま使える。依存ゼロ。
package main

import (
	"crypto/mldsa"
	"fmt"
)

func main() {
	priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		panic(err)
	}

	msg := []byte("miyazaki.go #5")
	sig, err := priv.Sign(nil, msg, &mldsa.Options{})
	if err != nil {
		panic(err)
	}

	pub := priv.PublicKey()
	fmt.Println("sig bytes:", len(sig))
	fmt.Println("verify:   ", mldsa.Verify(pub, msg, sig, &mldsa.Options{}) == nil)
	fmt.Println("tampered: ", mldsa.Verify(pub, []byte("tampered"), sig, &mldsa.Options{}) == nil)
}
