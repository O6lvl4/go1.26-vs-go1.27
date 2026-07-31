#!/bin/sh
# 全テーマの before/after を並べて実行する
set -e
cd "$(dirname "$0")"

for d in 01-generic-methods 02-json 03-uuid 04-stdlib-bits 05-goroutine-leak; do
  echo "━━━━━━━━━━ $d ━━━━━━━━━━"
  echo "--- before (〜1.26)"
  go run "./$d/before"
  echo "--- after (1.27)"
  go run "./$d/after"
  echo
done
