// 1.27: secret mode (runtime/secret、実験的) を goroutine が継承するようになった。
// 対応は linux/amd64・linux/arm64 のみで、それ以外の OS では Do は f を呼ぶだけの素通し。
//
//	GOEXPERIMENT=runtimesecret go run ./16-secret
//
// linux での挙動 (stdlib の TestSecretInheritance と runtime 実装差分より):
//
//	1.26: 親 true / 子 false ← 並行化した瞬間に保護が漏れていた
//	1.27: 親 true / 子 true  (runtime/proc.go に newg.secret = 1 が追加された)
//
// darwin では親子とも false (未対応 OS の素通し挙動の実測)

//go:build goexperiment.runtimesecret

package main

import (
	"fmt"
	"runtime"
	"runtime/secret"
)

func main() {
	fmt.Printf("%s/%s (対応 OS は linux のみ)\n", runtime.GOOS, runtime.GOARCH)
	done := make(chan struct{})
	secret.Do(func() {
		fmt.Println("親 (secret.Do 内):          secret mode =", secret.Enabled())
		go func() {
			defer close(done)
			fmt.Println("子 goroutine (Do 内 spawn): secret mode =", secret.Enabled())
		}()
	})
	<-done
}
