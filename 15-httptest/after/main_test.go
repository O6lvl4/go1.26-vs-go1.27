// 1.27: httptest.NewTestServer はインメモリの偽ネットワーク上にサーバを立てる。
// OS のポートを 1 つも使わず、URL は常に http://example.com で決定的。
// t を受け取るので Close は t.Cleanup に自動登録され、synctest バブル内でも使える (10 参照)。
package after

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	})

	srv := httptest.NewTestServer(t, h) // defer 不要

	// 偽ネットワークへの配線を持つのは srv.Client() だけ。
	// 素の http.Get を使うと本物の example.com に行ってしまうので注意
	res, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	t.Logf("URL=%s body=%s (偽ネットワーク、ポート消費ゼロ)", srv.URL, body)

	// 100 台立てても OS リソースは食わない
	for range 100 {
		httptest.NewTestServer(t, h)
	}
}
