// 〜1.26: httptest.NewServer は 127.0.0.1 の本物の TCP ポートを listen する。
// URL は実行のたびに変わるランダムポート、Close は手動 (defer 忘れで fd リーク)、
// 大量・並列に立てると OS のポート/fd を実際に消費する。
package before

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

	srv := httptest.NewServer(h)
	defer srv.Close() // 後片付けは自分で

	res, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	t.Logf("URL=%s body=%s (本物の TCP ポート、毎回変わる)", srv.URL, body)

	// 100 台立てると OS のポートを 100 個掴む
	for range 100 {
		s := httptest.NewServer(h)
		defer s.Close()
	}
}
