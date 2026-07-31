#!/bin/sh
# 1.27 のツールチェーン系変更のデモ。コードの before/after ではなく
# コマンドの挙動が変わる系をここにまとめている。
set -e
cd "$(dirname "$0")"

echo "━━━━━━━━━━ go test が stdversion vet check を実行 ━━━━━━━━━━"
echo "\$ go test ./14-tools/stdversion  (module は go 1.23、コードは 1.24 の API)"
(cd stdversion && go test -count=1 . 2>&1) || true
echo

echo "━━━━━━━━━━ go mod tidy が require ブロックを統合 ━━━━━━━━━━"
tmp=$(mktemp -d)
cp tidy-merge/go.mod tidy-merge/main.go "$tmp/"
(cd "$tmp" && go mod tidy 2>/dev/null)
echo "--- before (手編集で増えた 2 ブロック)"
cat tidy-merge/go.mod
echo "--- after (direct/indirect の 2 ブロックに統合)"
cat "$tmp/go.mod"
rm -rf "$tmp"
echo

echo "━━━━━━━━━━ go fix の新 modernizer (atomictypes/embedlit/slicesbackward/unsafefuncs) ━━━━━━━━━━"
echo "\$ go fix -diff ./14-tools/fixdemo"
(cd .. && go fix -diff ./14-tools/fixdemo) || true
echo

echo "━━━━━━━━━━ go doc -ex / go doc pkg@version ━━━━━━━━━━"
echo "\$ go doc strings.ExampleCut  (example のソースをそのまま表示)"
go doc strings.ExampleCut
echo
echo "\$ go doc github.com/google/uuid@v1.2.0 New  (バージョン指定で doc を引く)"
go doc github.com/google/uuid@v1.2.0 New
