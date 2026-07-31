// 1.27: panic 時のトレースバックのヘッダ行に pprof の goroutine ラベルが
// 出るようになった (go directive が 1.27 以降のモジュールのみ)。
// 「どのテナントのどのジョブで落ちたか」がスタックだけで分かる。
// こちらもコードは同じで、GODEBUG で before/after を再現する:
//
//	GODEBUG=tracebacklabels=0 go run ./12-traceback-labels  # 〜1.26 相当
//	go run ./12-traceback-labels                            # 1.27 デフォルト
package main

import (
	"context"
	"runtime/pprof"
)

func main() {
	// pprof.Do は defer でラベルを復元してしまい panic 時には消えているので、
	// 常駐ラベルは SetGoroutineLabels で貼る
	ctx := pprof.WithLabels(context.Background(),
		pprof.Labels("job", "ingest", "tenant", "miyazaki"))
	pprof.SetGoroutineLabels(ctx)

	panic("boom")
}
