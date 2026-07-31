#!/bin/sh
# 全テーマの before/after を並べて実行する
set -e
cd "$(dirname "$0")"

for d in 01-generic-methods 02-json 03-uuid 04-stdlib-bits 05-goroutine-leak \
	06-struct-literal-keys 07-type-inference 08-mldsa 09-maphash-hasher; do
  echo "━━━━━━━━━━ $d ━━━━━━━━━━"
  echo "--- before (〜1.26)"
  go run "./$d/before"
  echo "--- after (1.27)"
  go run "./$d/after"
  echo
done

echo "━━━━━━━━━━ 10-synctest ━━━━━━━━━━"
echo "--- before (実時間 + 実 TCP: 2 秒かかる)"
go test -count=1 -v ./10-synctest/before 2>&1 | grep -E "body=|^ok"
echo "--- after (偽時計 + 偽ネットワーク: 一瞬)"
go test -count=1 -v ./10-synctest/after 2>&1 | grep -E "body=|^ok"
echo

echo "━━━━━━━━━━ 11-alloc (サイズ特化 malloc) ━━━━━━━━━━"
echo "--- before (GOEXPERIMENT=nosizespecializedmalloc = 1.26 相当)"
GOEXPERIMENT=nosizespecializedmalloc go run ./11-alloc
echo "--- after (1.27 デフォルト)"
go run ./11-alloc
echo

echo "━━━━━━━━━━ 12-traceback-labels ━━━━━━━━━━"
echo "--- before (GODEBUG=tracebacklabels=0 = 〜1.26 相当)"
GODEBUG=tracebacklabels=0 go run ./12-traceback-labels 2>&1 | head -3 || true
echo "--- after (1.27 デフォルト: ヘッダにラベルが出る)"
go run ./12-traceback-labels 2>&1 | head -3 || true
echo

echo "━━━━━━━━━━ 13-simd (実験的) ━━━━━━━━━━"
echo "--- before (スカラループ)"
go run ./13-simd/before
echo "--- after (GOEXPERIMENT=simd: ハードウェア SIMD)"
GOEXPERIMENT=simd go run ./13-simd/after
echo

echo "━━━━━━━━━━ 15-httptest ━━━━━━━━━━"
echo "--- before (実 TCP: URL は毎回違うポート)"
go test -count=1 -v ./15-httptest/before 2>&1 | grep -E "URL=|^ok"
echo "--- after (偽ネットワーク: URL は常に example.com)"
go test -count=1 -v ./15-httptest/after 2>&1 | grep -E "URL=|^ok"
echo

echo "ツールチェーン系 (stdversion / tidy / go fix / go doc) は ./14-tools/demo.sh で。"
