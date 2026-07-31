// 1.27: synctest.Sleep (Sleep+Wait のヘルパ) と httptest.NewTestServer
// (インメモリ偽ネットワーク) が追加。偽時計で 2 秒進めても実時間は一瞬、
// ネットワークも OS のポートを掴まない。
package after

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"
)

func TestExpiryAndPing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now() // バブル内の偽時計 (2000-01-01 00:00:00 UTC 起点)

		expired := make(chan struct{})
		time.AfterFunc(2*time.Second, func() { close(expired) })
		synctest.Sleep(2 * time.Second) // 偽時計を 2 秒進めて全 goroutine の完了を待つ
		<-expired                       // もう発火済み。実時間は経過していない

		// インメモリ偽ネットワーク上のサーバ。ポートを掴まない
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "pong")
		}))

		res, err := srv.Client().Get(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()

		t.Logf("body=%s elapsed=%v (偽時計)", body, time.Since(start).Round(time.Millisecond))
	})
}
