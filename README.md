# Go 1.26 → 1.27 Before / After

[Miyazaki.go Go v1.27rc Sneak Peek #5](https://miyazaki-go.connpass.com/) (2026-07-31) 向け。
Go 1.27 の新機能を「**1.26 までどう書いていたか**」と並べて眺める。姉妹 repo: [go1.27-vs-almide](https://github.com/O6lvl4/go1.27-vs-almide)（他言語との比較はこちら）。

Go は後方互換なので `before/` のコードも 1.27 でそのまま動く。この repo の diff は「壊れる変更」ではなく「**消せるようになったコード**」の一覧。

**全サンプル実行検証済み**: `go1.27rc2 darwin/arm64`。コードで示しにくい変更も末尾の[網羅表](#go-127-リリースノート網羅表)で全項目拾っている。

## 実行方法

```bash
# go.mod が go 1.27rc2 を指しているので、Go 1.21+ ならツールチェーンが自動取得される
go run ./01-generic-methods/before
go run ./01-generic-methods/after

# 全部まとめて (01〜13)
./demo.sh

# ツールチェーン系 (go test の stdversion / go mod tidy / go fix / go doc)
./14-tools/demo.sh
```

## テーマ

| # | Before (〜1.26) | After (1.27) |
|---|---|---|
| [01](01-generic-methods/) | package スコープのジェネリック関数 | ジェネリックメソッド |
| [02](02-json/) | encoding/json v1（寛容がデフォルト） | encoding/json/v2（厳格がデフォルト） |
| [03](03-uuid/) | github.com/google/uuid に依存 | 標準 `uuid`（依存ゼロ） |
| [04](04-stdlib-bits/) | CutLast 手書き / 型別 IntN / URL 手動コピー / QuoRem 手補正 | `CutLast` / `Rand.N` / `URL.Clone` / `Int.Divide` |
| [05](05-goroutine-leak/) | goroutine プロファイルを目視 | `goroutineleak` プロファイル |
| [06](06-struct-literal-keys/) | 埋め込みは型名経由のネストしたリテラル | 昇格フィールド名で直接初期化 |
| [07](07-type-inference/) | 合成リテラル・変換では明示実体化 | 関数型推論が全文脈に一般化 |
| [08](08-mldsa/) | 耐量子署名は cloudflare/circl に依存 | 標準 `crypto/mldsa`（依存ゼロ） |
| [09](09-maphash-hasher/) | コンテナごとに独自のハッシュ契約 | 標準 `maphash.Hasher[T]` |
| [10](10-synctest/) | 実時間 sleep + 実 TCP のテスト | `synctest.Sleep` + `httptest.NewTestServer` |
| [11](11-alloc/) | (コードは同じ) | サイズ特化 malloc で小アロケーション高速化 |
| [12](12-traceback-labels/) | panic のスタックに文脈情報なし | traceback ヘッダに pprof ラベル |
| [13](13-simd/) | SIMD は cgo/アセンブリ/自動ベクトル化なし | 実験的 `simd` パッケージ（ポータブル） |
| [14](14-tools/) | — | ツールチェーンの挙動変更いろいろ |

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

json.Unmarshal([]byte("\"a\xffb\""), &s)
// => "a�b" <nil> — 不正 UTF-8 が黙って U+FFFD に化ける。データ破壊がエラーにならない

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

err = json.Unmarshal([]byte("\"a\xffb\""), &s)
// => jsontext: invalid UTF-8 after offset 2

json.Unmarshal(data, &m, jsontext.AllowDuplicateNames(true))    // v1 の挙動が要るならこれ
json.Unmarshal(data, &ev, json.MatchCaseInsensitiveNames(true)) // case-insensitive もオプトイン
json.Unmarshal(data, &s, jsontext.AllowInvalidUTF8(true))       // UTF-8 も同様

out, _ := json.Marshal(ev, jsontext.WithIndent("  ")) // 整形も Options
```

内部的には既存の `encoding/json` も v2 実装に置き換わった（`GOEXPERIMENT=nojsonv2` で旧実装に戻せる）。v1 API の挙動は互換維持なので、`before/` は 1.27 でも同じ出力を返す — 実測済み。Unmarshal 性能は大幅に向上。

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
// CutLast がないので LastIndex + スライスを毎回手書き (strings も bytes も)
func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}

// Rand は型ごとに別メソッド
r.IntN(10), r.Int64N(100), uint8(r.Uint64N(50))

// big.Int の floor 除算は QuoRem からの手補正が定番
quo, rem := new(big.Int).QuoRem(x, y, new(big.Int))
if rem.Sign() != 0 && (rem.Sign() < 0) != (y.Sign() < 0) {
	quo.Sub(quo, big.NewInt(1))
	rem.Add(rem, y)
}

// URL の複製は手書きシャローコピー。*Userinfo が共有される罠つき
v := *u
v.RawQuery = "tab=2"
// u.User == v.User → true（深い複製は自己責任だった）
```

### After

```go
dir, file, _ := strings.CutLast("archive/2026/photo.tar.gz", "/")
host, port, _ := bytes.CutLast([]byte("[::1]:8080"), []byte(":")) // bytes 側にも同じものが入った

// 1 つのジェネリックメソッドで全整数型（01 の言語変更の stdlib 実用例）
r.N(10), r.N(int64(100)), r.N(uint8(50))

// 丸めモード (Trunc/Floor/Round/Ceil) を指定して一発
quo, rem := new(big.Int).Divide(x, y, new(big.Int), big.Floor)

v := u.Clone() // deep copy と明記。u.User == v.User → false（実測）。Values.Clone も追加
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

検出できるのは「ブロックしたまま到達不能」な goroutine。`for { time.Sleep(...) }` のような回り続けるタスクは映らない（詳しくは姉妹 repo の [06-adversarial](https://github.com/O6lvl4/go1.27-vs-almide#06--意地悪テスト-キャンセルは本物か)）。1.26 で実験入りしていたものが 1.27 で正式化（`net/http/pprof` の `/debug/pprof/goroutineleak` も利用可）。

## 06 — struct リテラルのセレクタキー

### Before: 読むときは昇格するのに、書くときだけ型名経由

```go
type Meta struct{ ID, Owner string }
type Server struct {
	Meta
	Host string
	Port int
}

s := Server{
	Meta: Meta{ID: "srv-1", Owner: "platform"}, // 型名がリテラルに漏れる
	Host: "example.com",
}
fmt.Println(s.ID) // 読みは昇格して s.ID なのに、書きだけ非対称
```

### After: リテラルのキーに有効なフィールドセレクタが書ける

```go
// 1.26 までは unknown field X in struct literal エラーだった
s := Server{ID: "srv-1", Owner: "platform", Host: "example.com"}

// 昇格が何段深くても、セレクタとして有効なら書ける
type Spec struct{ Server }
sp := Spec{ID: "srv-2", Port: 9090}
```

`go fix` の新 modernizer `embedlit`（[14](14-tools/) 参照）が旧形式を自動で書き換えてくれる。
キーに書けるのは昇格フィールドの名前そのもので、`Meta.ID: "x"` のようなドット付きパスは 1.27 でも書けない（`invalid field name` — 実測）。

## 07 — 関数型推論の一般化

### Before: 文脈によって推論されたりされなかったり

```go
func double[T ~int | ~float64](v T) T { return v * 2 }

var f func(int) int = double // 変数代入は 1.21 から推論 OK

// 合成リテラル・チャネル送信・型変換は明示実体化が必須だった
ops := map[string]func(int) int{"double": double[int]}
ch <- double[float64]
g := (func(int) int)(double[int])
```

### After: 「一致する関数型への代入・変換」なら全文脈で推論

```go
// 1.26 までは cannot use generic function double without instantiation
ops := map[string]func(int) int{"double": double}
ch <- double
g := (func(int) int)(double)
```

## 08 — 耐量子署名 ML-DSA (FIPS 204)

### そもそも「耐量子署名」って何？（中学生向け解説）

**デジタル署名**は、データに押す「偽造できないハンコ」。「このファイルは確かに本人が作った」「途中で 1 バイトも書き換えられていない」ことをコンピュータが確認できる仕組みで、アプリの自動アップデートも HTTPS の証明書も全部これで守られている。

ハンコが偽造できない理由は、土台に「解くのに何億年もかかる数学パズル」があるから。いまの署名（RSA や楕円曲線）が使うパズルは「巨大な数を素数のかけ算に戻す」系統。15 = 3 × 5 は一瞬で分かるけど、600 桁の数になると今のコンピュータでは宇宙の年齢より長くかかる。

ところが**量子コンピュータ**は、よりによってこの系統のパズルにだけ効く「近道」を持つことが 1994 年から分かっている（ショアのアルゴリズム）。十分大きな量子コンピュータが完成した日に、いまのハンコは全部偽造し放題になる。しかも「完成してから乗り換える」では遅い — いまのうちに通信を録りためておき、将来まとめて解読・偽造する攻撃（harvest now, decrypt later）があるから、完成前の今から乗り換えが始まっている。

ML-DSA はその乗り換え先のひとつで、**格子**（何百次元の空間にびっしり並んだ点の中から一番近い点を探す、量子コンピュータでも近道が見つかっていないパズル）を土台にした署名。アメリカの標準化機関 NIST が 2024 年に **FIPS 204** という標準として制定した（元の名前は CRYSTALS-Dilithium）。代償は大きさで、いままでの Ed25519 の署名が 64 bytes なのに対して ML-DSA-65 は 3309 bytes（下の実測どおり）。それでも「量子コンピュータが来ても偽造されないハンコ」の値段としては安い、というのが世界の結論。

### Before: cloudflare/circl 等に依存を張る（03-uuid と同じ構図）

```go
import "github.com/cloudflare/circl/sign/mldsa/mldsa65" // go.mod に require が増える

pub, priv, err := mldsa65.GenerateKey(rand.Reader)
sig := make([]byte, mldsa65.SignatureSize)
err = mldsa65.SignTo(priv, msg, nil, true, sig)
ok := mldsa65.Verify(pub, msg, nil, sig)
```

### After: 標準 crypto/mldsa。x509 / TLS 1.3 からもそのまま使える

```go
import "crypto/mldsa" // 依存ゼロ

priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
sig, err := priv.Sign(nil, msg, &mldsa.Options{}) // crypto.Signer 実装
err = mldsa.Verify(priv.PublicKey(), msg, sig, &mldsa.Options{})
```

署名は 3309 bytes（ML-DSA-65、実測一致）。`crypto/x509` が ML-DSA 鍵/署名に、`crypto/tls` が TLS 1.3 の `MLDSA44/65/87` SignatureScheme に対応。

## 09 — hash コンテナの標準契約 maphash.Hasher

### Before: コンテナごとに独自のハッシュ契約を発明

```go
type Bloom[T any] struct {
	hash func(maphash.Seed, T) uint64 // このコンテナ独自の契約
	...
}

// 呼び出し側はコンテナごとの流儀で関数値を渡す。
// 等価関係(Equal)まで要るコンテナなら関数がもう 1 本増える
b := NewBloom(func(seed maphash.Seed, s string) uint64 { return maphash.String(seed, s) })
```

### After: Hasher[T] が標準の契約。ハッシュと等価関係が 1 つの型に束なる

```go
type Hasher[T any] interface { // hash/maphash に標準で入った
	Hash(*maphash.Hash, T)
	Equal(x, y T) bool
}

b := NewBloom[string](maphash.ComparableHasher[string]{}) // comparable 型は無料

type CaseInsensitive struct{}                             // 独自の等価関係も一級市民
func (CaseInsensitive) Hash(h *maphash.Hash, s string) { h.WriteString(strings.ToLower(s)) }
func (CaseInsensitive) Equal(x, y string) bool         { return strings.ToLower(x) == strings.ToLower(y) }
```

非 comparable な型を将来の hash ベースコンテナ (hash table / Bloom filter / #70471) のキーにするための土台。go/types にも同じ発想の `Hasher` が入った。

## 10 — テストの偽時計と偽ネットワーク

### Before: 実時間で 2 秒待ち、実 TCP ポートを掴む

```go
time.AfterFunc(2*time.Second, func() { close(expired) })
<-expired // テストが実時間で 2 秒かかる。CI 負荷で flaky に

srv := httptest.NewServer(handler) // 本物のポートを listen
```

### After: 偽時計は一瞬で進み、ネットワークはインメモリ

```go
synctest.Test(t, func(t *testing.T) {
	time.AfterFunc(2*time.Second, func() { close(expired) })
	synctest.Sleep(2 * time.Second) // 1.27 新 API: Sleep + Wait のヘルパ
	<-expired                       // もう発火済み。実時間ゼロ

	srv := httptest.NewTestServer(t, handler) // 1.27 新 API: 偽ネットワーク上のサーバ
	res, _ := srv.Client().Get(srv.URL)
})
```

実測: before 2.00s / after 0.00s。`synctest` 本体は 1.25 実験→1.26 正式化で、1.27 は `Sleep` ヘルパと httptest 連携が追加分。

## 11 — サイズ特化 malloc（コード変更ゼロで速くなる）

コンパイラが 80 byte 未満の小アロケーションにサイズ特化ルーチンを生成するようになった。コードは同じなので `GOEXPERIMENT` で新旧を比べる:

```bash
GOEXPERIMENT=nosizespecializedmalloc go run ./11-alloc  # 1.26 相当: 10.76 ns/op
go run ./11-alloc                                       # 1.27:      5.72 ns/op (実測 M4 Pro)
```

リリースノートの謳い文句は「最大 30%・実プログラム全体で〜1%」。バイナリは約 60KB 増える。opt-out は 1.28 で削除予定。

## 12 — panic トレースバックに pprof ラベル

```go
ctx := pprof.WithLabels(context.Background(), pprof.Labels("job", "ingest", "tenant", "miyazaki"))
pprof.SetGoroutineLabels(ctx)
panic("boom")
```

```
# 〜1.26 相当 (GODEBUG=tracebacklabels=0)
goroutine 1 [running]:

# 1.27 (go directive が 1.27 以降の module でデフォルト有効)
goroutine 1 [running] {job: ingest, tenant: miyazaki}:
```

「どのテナントのどのジョブで落ちたか」がスタックトレースだけで分かる。注意: `pprof.Do` は defer でラベルを復元するので、panic 時に見せたい常駐ラベルは `SetGoroutineLabels` で貼る。

## 13 — 実験的 simd パッケージ

### Before: 純 Go はスカラループのみ（自動ベクトル化なし）

```go
func axpy(a float64, xs, ys []float64) {
	for i := range xs {
		ys[i] = a*xs[i] + ys[i]
	}
}
```

### After: ベクトル幅非依存のポータブル SIMD（要 GOEXPERIMENT=simd）

```go
import "simd"

av := simd.BroadcastFloat64s(a)
for i := 0; i < len(xs); i += av.Len() {
	x, _ := simd.LoadFloat64sPart(xs[i:]) // 端数はゼロ埋めで面倒を見てくれる
	y, _ := simd.LoadFloat64sPart(ys[i:])
	av.MulAdd(x, y).StorePart(ys[i:])
}
```

arm64 (Neon) 実測: `lanes: 2, emulated: false` — ハードウェア SIMD で動く。amd64 AVX / wasm も同じコードで動き、非対応環境は純 Go エミュレーション。アーキ固有 API が欲しければ `simd/archsimd`（1.26 実験開始、1.27 で arm64/wasm 対応追加）。

## 14 — ツールチェーンの挙動変更（`./14-tools/demo.sh`）

| デモ | 変更 |
|---|---|
| [stdversion/](14-tools/stdversion/) | `go test` が stdversion vet check をデフォルト実行。「go.mod の版より新しい stdlib API」がテストで落ちる |
| [tidy-merge/](14-tools/tidy-merge/) | go 1.27+ module では `go mod tidy` が散らばった require ブロックを direct/indirect の 2 つに統合（コメントは保持） |
| [fixdemo/](14-tools/fixdemo/) | `go fix` の新 modernizer: `atomictypes` / `embedlit` / `slicesbackward` / `unsafefuncs`。`go fix -diff` でパッチ確認 |
| (demo.sh 内) | `go doc pkg@version` と `go doc -ex`（example のソース表示） |

補足: stdversion の fixture が「go 1.23 module + 1.24 API」なのは、rc 版ツールチェーンでは 1.27 シンボル（例: go 1.26 module での `strings.CutLast`）が拾われなかったため（rc の版数比較によるものと思われる。リリース版 1.27 では拾われるはず）。

## Go 1.27 リリースノート網羅表

コードで示せる項目は例に、示しにくい項目は注記で全項目カバー。

| リリースノート項目 | 扱い |
|---|---|
| **言語**: ジェネリックメソッド | [01](01-generic-methods/) |
| **言語**: struct リテラルのセレクタキー | [06](06-struct-literal-keys/) |
| **言語**: 関数型推論の一般化 | [07](07-type-inference/) |
| **ツール**: response file (@file) 対応 | 注記のみ: compile/link/asm/cgo/cover/pack が GCC 互換の @file を受理。ビルドシステム連携向け |
| **go command**: bzr サポート削除 | 注記のみ |
| **go command**: 削除済み GODEBUG 設定の受理 | 注記のみ: 最終デフォルト値なら go.mod / //go:debug に残っていてもビルド可 |
| **go test**: stdversion vet check | [14-tools/stdversion](14-tools/stdversion/) |
| **go test**: -json の OutputType | 注記のみ |
| **go doc**: pkg@version / -ex | [14-tools/demo.sh](14-tools/demo.sh) |
| **go fix**: 新 modernizer 4 種 | [14-tools/fixdemo](14-tools/fixdemo/)（fmtappendf は削除、waitgroup → waitgroupgo に改名） |
| **go mod tidy**: require ブロック統合 | [14-tools/tidy-merge](14-tools/tidy-merge/) |
| **trace**: -http がデフォルト localhost | 注記のみ: pprof と挙動統一 |
| **runtime**: traceback に pprof ラベル | [12](12-traceback-labels/) |
| **runtime**: asynctimerchan 削除 | 注記のみ: time のチャネルは常に同期（unbuffered）に確定 |
| **runtime**: サイズ特化 malloc | [11](11-alloc/) |
| **runtime**: goroutineleak プロファイル正式化 | [05](05-goroutine-leak/) |
| **コンパイラ**: //line 相対パス解決の統一 | 注記のみ: go/scanner と同じ解釈に |
| **コンパイラ**: クロージャ名の単純化 | 注記のみ: 関数リテラルのコードポインタ比較に依存していると挙動が変わりうる |
| **リンカ**: -macos / -macsdk | 注記のみ |
| **stdlib**: encoding/json/v2 + jsontext | [02](02-json/) |
| **stdlib**: crypto/mldsa | [08](08-mldsa/) |
| **stdlib**: uuid | [03](03-uuid/) |
| **stdlib**: simd / simd/archsimd（実験的） | [13](13-simd/) |
| bytes/strings.CutLast | [04](04-stdlib-bits/) |
| compress/flate 高速化 | 注記のみ: 出力バイト列も変わる。実測: 同一 15500B 入力 → 1.26.5 は 93B、1.27rc2 は 90B（zip/gzip/zlib/png にも波及） |
| crypto.MLDSAMu | 注記のみ: External μ ML-DSA 署名のシグナル用 |
| crypto/ecdsa Sign のハッシュ長チェック | 注記のみ |
| crypto/tls（MLKEM1024 / LocalCertificate / QUICConfig.ClientHelloInfoConn / Config.Rand 非推奨 / GODEBUG 5 種削除） | 注記のみ: PQC ハイブリッド関連は 08 の文脈 |
| crypto/x509（pkix パース拡大 / RawSignatureAlgorithm / SSL_CERT_FILE・SSL_CERT_DIR が Win/mac でも有効） | 注記のみ |
| crypto/x509/pkix RDNSequence.String | 注記のみ |
| database/sql ConvertAssign / driver.RowsColumnScanner | 注記のみ: ドライバ作者向け |
| go/constant StringLen / go/scanner Scanner.End / go/token File.String | 注記のみ: ツール作者向け |
| go/types Hasher（+ gotypesalias GODEBUG 削除） | 注記のみ: [09](09-maphash-hasher/) と同じ Hasher 設計の go/types 版 |
| hash/maphash Hasher / ComparableHasher | [09](09-maphash-hasher/) |
| math/big Int.Divide | [04](04-stdlib-bits/) |
| math/rand/v2 Rand.N | [04](04-stdlib-bits/) |
| net UnixConn の io.EOF 直接返却 | 注記のみ |
| net/http（ユーザ提供 Conn での ALPN / HTTP/2 priority RFC 9218 / Response.Body 自動 drain / MaxHeaderValueCount） | 注記のみ: Body 自動 drain は接続再利用の改善。MaxIdleConns=0 運用は要確認 |
| net/http/httptest NewTestServer | [10](10-synctest/) |
| net/url URL.Clone / Values.Clone | [04](04-stdlib-bits/) |
| runtime/secret: secret モードの goroutine 継承 | 注記のみ |
| strings.CutLast | [04](04-stdlib-bits/) |
| syscall: Plan 9 の Errno 定義 | 注記のみ |
| testing/synctest Sleep | [10](10-synctest/) |
| unicode 15 → 17 | 注記のみ: テーブル更新（Unicode 16/17 の 2 世代分） |
| **ports**: macOS 13 Ventura 以降必須 | 注記のみ |
| **ports**: ppc64 の ELFv2 移行（cgo/PIE 対応） | 注記のみ |

## まとめ

| 観点 | 1.27 で変わったこと |
|---|---|
| 言語 | ジェネリックメソッド（8 年越し）+ リテラルのセレクタキー + 推論の一般化。3 つとも「書けなかった自然な書き方が書けるようになる」方向 |
| JSON | デフォルトの向きが寛容→厳格へ反転。v1 API は互換維持 |
| 依存 | uuid・ML-DSA のような「全員が同じサードパーティを入れるもの」の取り込み |
| API 設計 | 型別メソッド群 → ジェネリックメソッド 1 つ、手書き複製 → 公式 Clone、独自ハッシュ契約 → 標準 Hasher[T] |
| 観測性 | 「全部見せるから探して」→「異常だけ報告する」プロファイル + panic に文脈ラベル |
| 性能 | コード変更ゼロで malloc・flate・json が速くなる。SIMD は実験的に手動の道も |
| ツール | 書き方の現代化（go fix）と「新しすぎる API」の検出（stdversion）が両輪に |
