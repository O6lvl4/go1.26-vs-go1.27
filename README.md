# Go 1.26 → 1.27 Before / After

[Miyazaki.go Go v1.27rc Sneak Peek #5](https://miyazaki-go.connpass.com/) (2026-07-31) 向け。
Go 1.27 の新機能を「**1.26 までどう書いていたか**」と並べて眺める。姉妹 repo: [go1.27-vs-almide](https://github.com/O6lvl4/go1.27-vs-almide)（他言語との比較はこちら）。

Go は後方互換なので `before/` のコードも 1.27 でそのまま動く。この repo の diff は「壊れる変更」ではなく「**消せるようになったコード**」の一覧。

**全サンプル実行検証済み**: `go1.27rc2 darwin/arm64`。

## 実行方法

```bash
# go.mod が go 1.27rc2 を指しているので、Go 1.21+ ならツールチェーンが自動取得される
go run ./01-generic-methods/before
go run ./01-generic-methods/after

# 全部まとめて
./demo.sh
```

## テーマ

| # | Before (〜1.26) | After (1.27) |
|---|---|---|
| [01](01-generic-methods/) | package スコープのジェネリック関数 | ジェネリックメソッド |
| [02](02-json/) | encoding/json v1（寛容がデフォルト） | encoding/json/v2（厳格がデフォルト） |
| [03](03-uuid/) | github.com/google/uuid に依存 | 標準 `uuid`（依存ゼロ） |
| [04](04-stdlib-bits/) | CutLast 手書き / 型別 IntN / URL 手動コピー | `strings.CutLast` / `Rand.N` / `URL.Clone` |
| [05](05-goroutine-leak/) | goroutine プロファイルを目視 | `goroutineleak` プロファイル |

## 01 — ジェネリックメソッド

### Before: 型の操作なのに package の名前空間に置くしかない

```go
// Stack 用の Map と Tree 用の Map は同じ package に共存できない
func Map[T, U any](s *Stack[T], f func(T) U) *Stack[U] { ... }
func Fold[T, A any](s *Stack[T], init A, f func(A, T) A) A { ... }

// 関数形式なのでチェーンできない。内側から外側へ読む
joined := Fold(
	Map(s, func(v int) string { return fmt.Sprintf("<%d>", v) }),
	"",
	func(acc, v string) string { return strings.TrimPrefix(acc+"-"+v, "-") },
)
```

### After: 型の名前空間に住み、左から右に読める

```go
// 1.26 まではコンパイルエラーだった宣言
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U] { ... }
func (s *Stack[T]) Fold[A any](init A, f func(A, T) A) A { ... }

joined := s.
	Map(func(v int) string { return fmt.Sprintf("<%d>", v) }).
	Fold("", func(acc, v string) string { return strings.TrimPrefix(acc+"-"+v, "-") })
```

制限は残る: interface のメソッドは型パラメータを宣言できず、ジェネリックメソッドで interface を実装することもできない。

## 02 — encoding/json v1 → v2

### Before: 寛容がデフォルト（そして 1.27 でも v1 API は互換維持）

```go
var m map[string]int
json.Unmarshal([]byte(`{"a":1,"a":2}`), &m)
// => map[a:2] <nil> — 重複キーは黙って後勝ち

json.Unmarshal([]byte(`{"NAME":"x"}`), &ev)
// => ev.Name == "x" — case-insensitive にマッチしてしまう

out, _ := json.MarshalIndent(ev, "", "  ") // 整形は別関数
```

### After: 厳格がデフォルト、緩めるのはオプトイン

```go
import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &m)
// => jsontext: duplicate object member name "a"

json.Unmarshal(data, &m, jsontext.AllowDuplicateNames(true))    // v1 の挙動が要るならこれ
json.Unmarshal(data, &ev, json.MatchCaseInsensitiveNames(true)) // case-insensitive もオプトイン

out, _ := json.Marshal(ev, jsontext.WithIndent("  ")) // 整形も Options
```

内部的には既存の `encoding/json` も v2 実装に置き換わった（`GOEXPERIMENT=nojsonv2` で旧実装に戻せる）。v1 API の挙動は互換維持なので、`before/` は 1.27 でも同じ出力を返す — 実測済み。

## 03 — uuid

### Before: google/uuid に依存を張るのが定石

```go
import "github.com/google/uuid" // go.mod に require が増える

v7, err := uuid.NewV7() // error を返すシグネチャ
// Compare メソッドがないのでソートは bytes.Compare を手書き
slices.SortFunc(ids, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
```

### After: import は一語、エラーも消える

```go
import "uuid" // 依存ゼロ (RFC 9562)

ids := []uuid.UUID{uuid.NewV7(), uuid.NewV7(), uuid.NewV7()} // error なし
slices.SortFunc(ids, uuid.UUID.Compare)                      // Compare が生えている
```

## 04 — stdlib 小ネタ

### Before

```go
// CutLast がないので LastIndex + スライスを毎回手書き
func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}

// Rand は型ごとに別メソッド
r.IntN(10), r.Int64N(100), uint8(r.Uint64N(50))

// URL の複製は手書きシャローコピー。*Userinfo が共有される罠つき
v := *u
v.RawQuery = "tab=2"
// u.User == v.User → true（深い複製は自己責任だった）
```

### After

```go
dir, file, _ := strings.CutLast("archive/2026/photo.tar.gz", "/")

// 1 つのジェネリックメソッドで全整数型（01 の言語変更の stdlib 実用例）
r.N(10), r.N(int64(100)), r.N(uint8(50))

v := u.Clone() // deep copy と明記。u.User == v.User → false（実測）
```

## 05 — goroutine リーク調査

### Before: 全 goroutine から目視で探す

```go
pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
```

```
goroutine profile: total 4
3 @ ... main.leak.func1        ← これがリーク（だと人間が判断する）
1 @ ... runtime/pprof.writeGoroutine ← 健全な goroutine も混ざる
```

### After: リークだけが出る

```go
pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
```

```
goroutineleak profile: total 3
3 @ ... main.leak.func1        ← 実行可能な goroutine から到達不能なものだけ
```

検出できるのは「ブロックしたまま到達不能」な goroutine。`for { time.Sleep(...) }` のような回り続けるタスクは映らない（詳しくは姉妹 repo の [06-adversarial](https://github.com/O6lvl4/go1.27-vs-almide#06--意地悪テスト-キャンセルは本物か)）。

## まとめ

| 観点 | 1.27 で変わったこと |
|---|---|
| 言語 | ジェネリクスの置き場所がメソッドに拡張（2018 年の設計議論から 8 年越し） |
| JSON | デフォルトの向きが寛容→厳格へ反転。v1 API は互換維持 |
| 依存 | uuid のような「全員が同じサードパーティを入れるもの」の取り込み |
| API 設計 | 型別メソッド群 → ジェネリックメソッド 1 つ、手書き複製 → 公式 Clone |
| 観測性 | 「全部見せるから探して」→「異常だけ報告する」プロファイル |
