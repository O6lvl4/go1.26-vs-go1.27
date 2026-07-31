// 〜1.26 流: 実時間と実 TCP に依存したテスト。
// TTL 切れを time.Sleep で待つのでテストが実時間で 2 秒かかり、
// サーバは本物のポートを掴む。CI の負荷次第で flaky になる典型形。
package before

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExpiryAndPing(t *testing.T) {
	start := time.Now()

	expired := make(chan struct{})
	time.AfterFunc(2*time.Second, func() { close(expired) })
	<-expired // 実時間で 2 秒ブロックする

	// 実 TCP ポートを listen する本物のサーバ
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	}))
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	t.Logf("body=%s elapsed=%v (実時間)", body, time.Since(start).Round(time.Millisecond))
}
