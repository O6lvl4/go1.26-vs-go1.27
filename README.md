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

# 全部まとめて (14 のツール系以外)
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
| [15](15-httptest/) | httptest.NewServer が実 TCP ポートを listen | `NewTestServer` が偽ネットワーク上に（URL は常に example.com） |
| [16](16-secret/) | secret mode は goroutine に漏れ穴 | spawn した goroutine も secret mode を継承 |
| [17](17-plan9-errno/) | Errno コードは Plan 9 ビルドで壊れる | Plan 9 にも `syscall.Errno` が定義され build tag 不要 |

## 01 — ジェネリックメソッド

### そもそも「ジェネリックメソッド」って何？

プログラムの関数はふつう「int（整数）用」「string（文字列）用」と型を決めて書く。同じ処理を型違いで何度も書くのは無駄なので、**ジェネリクス**という「型の部分を穴あきにした設計図」が Go 1.18 で入った。穴 `[T]` にはあとで好きな型がはまる。

**メソッド**は型にくっついた関数のこと。`s.Push(1)` のように「そのデータ自身に命令する」形で書ける。ところが 1.26 までは「穴あきの関数」を型にくっつけることが文法上できず、型の外に転がった関数として書くしかなかった。1.27 でくっつけられるようになり、`s.Map(...).Fold(...)` と左から右へ読める自然な形になった。

### 何が嬉しい？

- ライブラリ作者: 型にまつわる操作を型の名前空間に置けるので、エディタで `s.` と打てば補完に全部並ぶ。package 直下に `MapStack` / `MapTree` が同居できず改名するしかなかった名前衝突問題が消える。
- 利用者: 内側から外側へ読む入れ子呼び出しが、左から右へ流れるチェーンになる。型引数は推論されるので書くものも減る。
- 2018 年の設計議論から 8 年越しのジェネリクス最後の大物で、stdlib 自身が早速 `Rand.N`（[04](04-stdlib-bits/)）で使っている。

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

### そもそも「JSON の寛容と厳格」って何？

**JSON** はプログラム同士がデータをやりとりするときの共通の書き方（`{"name": "miyazaki"}` のようなテキスト）。世界中の Web サービスがこれで会話している。

いままでの Go の JSON 読み取りは「多少変なデータでも顔パスで通す警備員」だった。同じ名前が 2 回出てきても黙って後の方を採用、大文字小文字の違いも見逃し、壊れた文字は勝手に置き換えて素通し。一見親切だけど、「気づかないうちに攻撃者の仕込んだ別のデータを受け入れていた」という事故の温床だった。v2 は「まず全部ちゃんと調べて、怪しければ止める警備員」。緩くしたいときだけ、許可証（Options）を明示的に渡す方式に変わった。

### 何が嬉しい？

- 上げるだけで速い: 既存コードは書き換えゼロで中身が v2 実装になり、Unmarshal が大幅に高速化する。
- 事故がデフォルトで止まる: 重複キーや大文字小文字のすり抜けは「パーサ間の解釈差を突く」実在の攻撃手法。いままで各自が自衛していたものが標準で弾かれる。
- 移行は強制されない: v1 API は互換維持なので、新規コードから v2 を選べばいい。

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

### そもそも「UUID」って何？

**UUID** は「世界中の誰とも相談せずに作っても、他人と絶対かぶらない番号」。128 bit のランダムな数で、かぶる確率は宝くじ 1 等に連続で当たるよりはるかに低い。データベースの ID など、あらゆる場面の「名前の重複を避けたいもの」に使われている。

いままでは Google 製ライブラリを**依存**（= 他人の部品を借りてくること。go.mod という借り物リストに 1 行増える）として入れるのがお約束だった。ほぼ全員が同じ部品を借りているなら言語に最初から入れよう、というのが 1.27。

### 何が嬉しい？

- go.mod から依存が 1 つ消える = バージョン追従・脆弱性監視・ライセンス確認の対象が 1 つ減る。サプライチェーン攻撃の入口も減る。
- 絶対に失敗しない `NewV7()` の error 処理や手書きの Compare など、全員が書いていた儀式コードが消える。
- チームの「どの uuid ライブラリを使う？」という選定・統一の議論が終了する。

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

### そもそも「stdlib（標準ライブラリ）」って何？

**標準ライブラリ**は言語に最初から付いてくる道具箱。道具箱に無いものは各自が手作りする（= 同じような数行のコードを世界中の人が書く）ことになる。みんなが同じ道具を手作りしているなら標準に入れた方がいい、ということで 1.27 では「最後の区切り文字で切る」「型を問わない乱数」「余りの丸め方を選べる割り算」「URL の完全コピー」などの定番手作り品が公式の道具になった。

### 何が嬉しい？

- 「全員が一度は書いた 5 行」が 1 行になる。手書き cutLast のオフバイワンや、URL シャローコピーで `*Userinfo` が共有される罠のような定番バグが構造的に消える。
- `Divide` の丸めモード指定は、負数が絡む日付・ページング計算のたびに「Go の `/` と `%` は負数でどっちに丸まるんだっけ」と調べていた問題への公式回答。

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

## 番外 — stdlib と「v2」の物語（rand/v2 → json/v2）

### そもそも「v2 パッケージ」って何？

Go には「一度公開した API は壊さない」という強い約束がある（Go 1 互換性保証）。だから設計を根本からやり直したくなっても、既存パッケージを作り替えることはできない。代わりに「隣に新しい棟（v2）を建てて、住みたい人から引っ越してもらう。旧棟も取り壊さない」方式をとる。

- **第 1 号: math/rand/v2**（Go 1.22、2024 年 2 月）。乱数はプログラムの「サイコロ」で、v1 のサイコロは古いアルゴリズム（出目を予測されやすい・種 Seed の管理が事故りがち）だった。v2 で暗号品質の ChaCha8 がデフォルトになり、`Intn` → `IntN` などの命名も整理された。
- **第 2 号: encoding/json/v2**（1.27、本 repo の [02](02-json/)）。

つまり rand/v2 自体は 1.26→1.27 の差分ではない。**1.27 の rand の差分は `Rand.N` ただ 1 つ**（[04](04-stdlib-bits/)）。ただしこの 1 つが面白くて、トップレベル関数の `rand.N` は 1.22 からあったのに、同じもののメソッド版は「ジェネリックメソッドが書けない」という言語制約で 5 年間作れなかった。1.27 で言語が追いつき（[01](01-generic-methods/)）、即 stdlib が採用した — **言語の変更が API 設計を変える**一番小さな実例。

json v2 からはもう一つの学びがある: 1.27 で旧 `encoding/json` の中身は v2 実装に置き換わり、v1 API はその上の互換レイヤになった（[02](02-json/) 参照）。「v2 が出る = v1 が捨てられる」ではなく、**v1 の見た目のまま中身だけ新しくなる**。互換性の約束と設計のやり直しを両立させる Go 流の答えがこの v2 方式で、次の候補として議論されているのが go/ast などの古参たち。

### 何が嬉しい？

- 利用者にとっての v2 方式の意味は「大改修が来ても既存コードは 1 行も壊れない。でもコンパイラを上げるだけで中身は新しくなる」。Go のアップグレードが怖くないのはこの規律のおかげ。

## 05 — goroutine リーク調査

### そもそも「goroutine リーク」って何？

**goroutine** は Go の軽量な作業員。何千人でも同時に雇えるのが Go の強み。ただし「連絡が来るまで待ってて」と指示した相手から永遠に連絡が来ないと、その作業員は待ちぼうけのまま居座り続ける。これが**リーク**で、積もり積もるとメモリを食いつぶしてサーバが倒れる。

いままでの点呼（**プロファイル** = 実行中プログラムの健康診断）は「全員分の名簿を出すから、待ちぼうけの人を目で探してね」方式。1.27 の `goroutineleak` は「もう誰にも起こしてもらえないことが確定した人」だけを自動で選んで報告してくれる。

### 何が嬉しい？

- 本番の「メモリがじわじわ増える」調査が、数千行の goroutine 名簿とのにらめっこから「確定犯だけのリスト」に変わる。`/debug/pprof/goroutineleak` を叩くだけ。
- 「リーク 0 件」をテストや監視で機械的にチェックする運用が組める。
- Uber の大規模実運用（数千サービス）で鍛えられた検出手法がそのまま標準で手に入る。

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

### そもそも「埋め込みと昇格」って何？

**struct** はデータをまとめる箱（`Server` という箱に Host と Port を入れる、のような）。Go には箱の中に別の箱を名前なしでそのまま入れる**埋め込み**という機能があり、読むときは `s.ID` のように中の箱を意識せず直接取り出せる（これを**昇格**と呼ぶ）。

ところが箱を作る（初期化する）ときだけは `Server{Meta: Meta{ID: ...}}` と入れ子で書く決まりで、読みと書きが非対称だった。1.27 からは作るときも `Server{ID: ...}` と読むときと同じ名前で書ける。

### 何が嬉しい？

- 埋め込みを多用する設定 struct・オプション struct の初期化が浅くなり、利用側から埋め込み構造が見えなくなる。
- リファクタ耐性が上がる: フィールドを共通の埋め込み型へ移動しても、読み取り（`s.ID`）だけでなくリテラルも書き換え不要になる。いままでは移動した瞬間、全初期化箇所の修正が必要だった。
- 既存コードの書き換えは `go fix`（embedlit、[14](14-tools/)）が自動でやってくれる。

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

### そもそも「型推論」って何？

**型推論**は、コンパイラが文脈から型を察してくれる機能。`x := 1` と書くだけで x が整数になるのもこれ。ジェネリック関数（01 参照）を使うときも本来は穴 `[T]` を埋める必要があるが、「代入先が `func(int) int` なんだから T は int でしょ」と察せる場面では省略できる。

ただし 1.26 までは察しの良さに場面ごとのムラがあった。変数への代入では察してくれるのに、map リテラルの中や型変換では「ちゃんと `[int]` と書け」と怒られる。1.27 でこのムラがなくなり、「一致する関数型に入れる場面なら常に察する」に統一された。

### 何が嬉しい？

- ハンドラを map に並べるディスパッチテーブルや、コールバック登録の場面から `[int]` のノイズが消える。ジェネリック関数が「普通の関数と同じ感覚で値として配れる」ようになる。
- 「代入なら省けるのにリテラルだと書かされる」という説明不能な非一貫性が消え、覚えるルールが 1 つになる。

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

### そもそも「耐量子署名」って何？

**デジタル署名**は、データに押す「偽造できないハンコ」。「このファイルは確かに本人が作った」「途中で 1 バイトも書き換えられていない」ことをコンピュータが確認できる仕組みで、アプリの自動アップデートも HTTPS の証明書も全部これで守られている。

ハンコが偽造できない理由は、土台に「解くのに何億年もかかる数学パズル」があるから。いまの署名（RSA や楕円曲線）が使うパズルは「巨大な数を素数のかけ算に戻す」系統。15 = 3 × 5 は一瞬で分かるけど、600 桁の数になると今のコンピュータでは宇宙の年齢より長くかかる。

ところが**量子コンピュータ**は、よりによってこの系統のパズルにだけ効く「近道」を持つことが 1994 年から分かっている（ショアのアルゴリズム）。十分大きな量子コンピュータが完成した日に、いまのハンコは全部偽造し放題になる。しかも「完成してから乗り換える」では遅い — いまのうちに通信を録りためておき、将来まとめて解読・偽造する攻撃（harvest now, decrypt later）があるから、完成前の今から乗り換えが始まっている。

ML-DSA はその乗り換え先のひとつで、**格子**（何百次元の空間にびっしり並んだ点の中から一番近い点を探す、量子コンピュータでも近道が見つかっていないパズル）を土台にした署名。アメリカの標準化機関 NIST が 2024 年に **FIPS 204** という標準として制定した（元の名前は CRYSTALS-Dilithium）。代償は大きさで、いままでの Ed25519 の署名が 64 bytes なのに対して ML-DSA-65 は 3309 bytes（下の実測どおり）。それでも「量子コンピュータが来ても偽造されないハンコ」の値段としては安い、というのが世界の結論。

### 何が嬉しい？

- 依存ゼロで耐量子署名を試せて、`crypto/x509` と TLS 1.3 に直結しているので「証明書やサーバを ML-DSA 対応にする」実験まで stdlib だけで完結する。
- 政府調達系の PQC 移行期限（CNSA 2.0 など）への対応が始めやすくなる。録りため攻撃（harvest now, decrypt later）への対策は「量子コンピュータが完成してから」では手遅れなので、始める時期は今。

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

### そもそも「ハッシュ」って何？

**ハッシュ**はデータから作る短い「指紋」。同じデータからは必ず同じ指紋が出る。**ハッシュテーブル**（Go の map の中身）は、指紋の値で保管場所（引き出し）を決めることで、何百万件の中からでも一瞬で見つけられるデータ構造。この repo の例で使っている **Bloom フィルタ**も親戚で、「リストに絶対無いことだけは即答できる」超省メモリな出席簿。

こうした自作コンテナに要素を入れるには「指紋の取り方」と「どれとどれを同じとみなすか」をコンテナに教える必要があるが、教え方の流儀がコンテナごとにバラバラだった。1.27 で「この 2 つをセットで持つ型を渡す」という標準の取り決め `Hasher[T]` ができた。

### 何が嬉しい？

- スライスや map を抱えた struct のような「組み込み map のキーにできない型」をキーにするコンテナが、正攻法で書けるようになる。
- 契約が標準化されたので、将来 stdlib に hash コンテナ本体（#70471 で議論中）が入ったとき、いま書いた Hasher がそのまま挿さる。ライブラリ間の相互運用も効く。
- Seed を混ぜる前提の設計なので、hash flooding（同じ指紋のデータを大量に投げて map を遅くする DoS 攻撃）への耐性が最初から組み込まれている。

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

### そもそも「偽時計でテスト」って何？

**テスト**は「プログラムが正しく動くか自動で確かめる小プログラム」。「2 秒経ったらキャッシュが消えること」を確かめるのに本物の 2 秒を待っていると、テストが千個に増えたとき悲惨だし、混み合った CI ではタイミングがずれて「たまに落ちる」不安定テストになる。

1.27 の道具は**偽の時計**（ゲーム内時間のように、待っている人が揃った瞬間に 2 秒進む）と**偽のネットワーク**（OS の本物の通信ポートを使わない、メモリ内の配線）をくれる。プログラム側は本物と同じコードのまま、実測どおり 2 秒のテストが 0 秒で終わる。

### 何が嬉しい？

- テストが速くなるだけでなく**決定的**になる。CI の「たまに落ちるから再実行」文化と、そのたびに失われる信頼と時間を削れる。
- リトライ・タイムアウト・TTL・デバウンスのような「時間のロジック」は、いままで実質テスト不能（sleep 挟んで祈る）だった。それが普通の単体テストになる。

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

実測: before 2.00s / after 0.00s。`synctest` 本体は 1.25 実験→1.26 正式化で、1.27 は `Sleep` ヘルパと httptest 連携が追加分。httptest 側の掘り下げは [15](15-httptest/)。

## 11 — サイズ特化 malloc（コード変更ゼロで速くなる）

### そもそも「malloc（アロケーション）」って何？

プログラムが「新しいデータの置き場所ください」とメモリをもらう操作を**アロケーション**（malloc）と呼ぶ。Go のプログラムはこれを毎秒何百万回もやるので、窓口が少し速くなるだけで全体が効く。1.27 はこの窓口に「よくある小さいサイズ専用の高速レーン」を作った。スーパーの「お会計 10 点まで専用レジ」と同じ発想で、コードは 1 文字も変えずに速くなる。

### 何が嬉しい？

- 何もしなくても速くなる。1.27 にビルドし直すだけで、アロケーションの多い API サーバほど効く（謳い文句は小アロケーション最大 30%・全体 〜1%、下の実測では 24B が ほぼ半減）。
- 「小さい struct をポンポン作る Go らしい書き方」の実行時ペナルティが下がり、可読性と性能のトレードオフが緩む。

コンパイラが 80 byte 未満の小アロケーションにサイズ特化ルーチンを生成するようになった。コードは同じなので `GOEXPERIMENT` で新旧を比べる:

```bash
GOEXPERIMENT=nosizespecializedmalloc go run ./11-alloc  # 1.26 相当: 10.76 ns/op
go run ./11-alloc                                       # 1.27:      5.72 ns/op (実測 M4 Pro)
```

リリースノートの謳い文句は「最大 30%・実プログラム全体で〜1%」。バイナリは約 60KB 増える。opt-out は 1.28 で削除予定。

## 12 — panic トレースバックに pprof ラベル

### そもそも「トレースバック」って何？

プログラムが想定外の状況でクラッシュ（**panic**）すると、Go は「その瞬間、どの関数のどの行で何をしていたか」の現場写真（**スタックトレース**）を残してくれる。事故調査はこの写真から始まる。

いままでの写真には「場所」しか写らなかった。1.27 からは、goroutine にあらかじめ貼っておいた**名札**（pprof ラベル: どの利用者の・どの種類の仕事か）も写真に写る。「サーバが落ちた。で、誰のどの処理で？」が写真 1 枚で分かるようになった。

### 何が嬉しい？

- 障害対応の初動が変わる: パニックのスタックだけで「どのテナントのどのジョブか」まで分かるので、ログを掘って再現条件を探す工程を丸ごと飛ばせることがある。マルチテナント SaaS では実質必須級。
- ラベルは pprof のプロファイルと共通なので、平常時の性能観測と障害時の証拠が同じ名札で繋がる。すでにラベルを貼っている運用なら、1.27 に上げるだけで恩恵が出る。

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

### そもそも「SIMD」って何？

CPU には「数を 1 個ずつ計算する命令」のほかに、「並べた数をまとめて一発で計算する命令」がある。これが **SIMD**（Single Instruction, Multiple Data）。レジで商品を 1 個ずつスキャンするか、カゴごと一括スキャンするかの違いで、画像処理・ゲーム・AI の高速化はだいたいこれが支えている。

いままで Go からこの命令を使うには C 言語やアセンブリ（機械語一歩手前の言語）の世界に降りる必要があった。1.27 では実験的に、Go のまま書けるようになった。しかも CPU の種類（Apple の arm64 / Intel の amd64 / ブラウザの wasm）を問わず同じコードで動き、SIMD の無い環境では自動で普通の計算に切り替わる。

### 何が嬉しい？

- 数値処理・画像処理・ML の前処理などを、cgo の呼び出しコストとビルドの複雑さ（クロスコンパイル崩壊）を抱えずに高速化する道が開く。
- CPU ごとにコードを分けなくていい。arm64 の Mac で書いて amd64 のサーバや wasm にそのまま持っていける。「Go は SIMD が使えないから C で」という理由が一つ消えていく（まだ実験的なので本採用は 1.28 以降の様子見で）。

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

### そもそも「ツールチェーン」って何？

Go は言語本体（コンパイラ）だけでなく、テストを回す `go test`、借り物リストを整頓する `go mod tidy`、古い書き方を新しい書き方に自動で直す `go fix`、説明書を引く `go doc` といった**道具一式（ツールチェーン）**が付いてくる。1.27 ではこの道具たちも賢くなった: 「宣言しているバージョンより新しい API をうっかり使ったら教える」「散らかった借り物リストを畳む」「新しい書き方への引っ越しを自動化する」など。言語の新機能（06 など）が出ると、`go fix` がその新しい書き方への書き換えまで面倒を見てくれる、という連携になっている。

### 何が嬉しい？

- stdversion: 「ライブラリ作者が go.mod の宣言より新しい API をうっかり使い、古い Go の利用者のビルドだけ壊れる」という発覚が遅い定番事故を、手元の `go test` の時点で捕まえてくれる。
- tidy のブロック統合: マージ衝突や手編集で荒れた go.mod が、次の tidy で勝手に綺麗になる。レビューの diff も安定する。
- go fix: 「新しい書き方に追従するコスト」を道具側が引き受けてくれるので、新機能が「知ってる人だけ使う方言」にならない。

| デモ | 変更 |
|---|---|
| [stdversion/](14-tools/stdversion/) | `go test` が stdversion vet check をデフォルト実行。「go.mod の版より新しい stdlib API」がテストで落ちる |
| [tidy-merge/](14-tools/tidy-merge/) | go 1.27+ module では `go mod tidy` が散らばった require ブロックを direct/indirect の 2 つに統合（コメントは保持） |
| [fixdemo/](14-tools/fixdemo/) | `go fix` の新 modernizer: `atomictypes` / `embedlit` / `slicesbackward` / `unsafefuncs`。`go fix -diff` でパッチ確認 |
| (demo.sh 内) | `go doc pkg@version` と `go doc -ex`（example のソース表示） |

補足: stdversion の fixture が「go 1.23 module + 1.24 API」なのは、rc 版ツールチェーンでは 1.27 シンボル（例: go 1.26 module での `strings.CutLast`）が拾われなかったため（rc の版数比較によるものと思われる。リリース版 1.27 では拾われるはず）。

## 15 — httptest の偽ネットワーク

### そもそも「httptest」って何？

Web サーバのコード（ハンドラ）をテストするには相手役のサーバが要る。**httptest** は「テスト用の使い捨てサーバを 1 行で立てる」標準の道具で、いままでは OS から本物の TCP ポートを 1 つ借りて立てていた。ポート番号は毎回変わり、テストを何千個も並列で回す現場ではポートや fd（ファイルディスクリプタ）の取り合いが起きる。

1.27 の `NewTestServer` は、配線がメモリの中で完結する**偽のネットワーク**の上にサーバを立てる。OS に一切触らないので住所は常に `http://example.com` で固定、後片付けも `t` に登録されて自動で行われる。

### 何が嬉しい？

- URL が常に同じなので、スナップショットテストやログ比較に「毎回変わるポート番号」のノイズが混ざらない。
- ポート枯渇・並列テスト同士の衝突・Close し忘れの fd リークが構造的に消える。100 台立てても OS リソース消費ゼロ。
- [10](10-synctest/) の偽時計と組み合わせると「時間もネットワークも偽物」になり、HTTP クライアントのリトライやタイムアウトまで完全に決定的にテストできる。

### Before: 本物の TCP ポートを 1 つ借りる

```go
srv := httptest.NewServer(h)
defer srv.Close() // 後片付けは自分で。忘れると fd リーク

srv.Client().Get(srv.URL)
// URL=http://127.0.0.1:57073 — 本物のポート。実行のたびに変わる (実測)
```

### After: 偽ネットワーク上に立てる

```go
srv := httptest.NewTestServer(t, h) // Close は t.Cleanup に自動登録

srv.Client().Get(srv.URL)
// URL=http://example.com — 固定アドレス。ポート消費ゼロ (実測)
```

注意: 偽ネットワークへの配線を持つのは `srv.Client()` だけ。素の `http.Get` を使うと本物の example.com に行ってしまう。synctest バブル内で HTTP テストができるのは、この偽ネットワークが OS のブロッキング I/O を経由しないおかげ。

## 16 — secret mode の goroutine 継承（実験的）

### そもそも「secret mode」って何？

暗号の秘密鍵のような「絶対に漏らしたくないデータ」は、使い終わったあともメモリに残骸が残る。プログラムがクラッシュしたときのダンプ（メモリの記念写真）や、同じメモリを後から再利用する別のコードから、その残骸を読まれてしまうことがある。**runtime/secret**（1.26 で実験導入）の `secret.Do(f)` は、f が使ったレジスタ・スタック・ヒープを、使い終わりしだい確実にゼロで上書きしてくれる「シュレッダー付きの作業部屋」。

ところが 1.26 では、部屋の中から `go` で新しい作業員（goroutine）を呼ぶと、**その作業員は部屋の外で働いていた**。鍵を扱う処理を並行化した瞬間、シュレッダーの保証が静かに消える罠があった。1.27 で「部屋の中で生まれた作業員も部屋の中」に直った。

### 何が嬉しい？

- 暗号ライブラリの作者が「Do の中では go を書くな」という守りにくい規約を背負わなくてよくなる。並行化しても forward secrecy（あとで鍵が漏れても過去の通信は守られる性質）の前提が崩れない。
- ただし対応は linux/amd64・linux/arm64 のみ、`GOEXPERIMENT=runtimesecret` 必須の実験段階。いまは「こういう部屋が標準に入りつつある」と知っておくのが正しい距離感。

### 挙動

```go
secret.Do(func() {
	fmt.Println(secret.Enabled()) // 親
	go func() { fmt.Println(secret.Enabled()) }() // 子 ← ここが 1.27 の差分
})
```

| 環境 | 親 | 子 |
|---|---|---|
| linux + 1.26 | true | **false**（保護が漏れる） |
| linux + 1.27 | true | **true** |
| darwin など未対応 OS | false | false（Do は素通し・実測） |

linux の挙動は stdlib 自身の `TestSecretInheritance` が保証している。実装差分も 1 行で分かりやすく、1.27 の `runtime/proc.go` の goroutine 生成部に「親が secret mode なら `newg.secret = 1`」が追加された（1.26 にはこの行が無い）。

## 17 — Plan 9 にも syscall.Errno

### そもそも「Plan 9」って何？

**Plan 9 from Bell Labs** は、UNIX を作った当のベル研チームが 1980 年代末に「UNIX の設計をやり直すなら」と作った研究用 OS。名前の元ネタは「史上最低の映画」と名高い B 級 SF『Plan 9 from Outer Space』(1959) で、当人たちが大真面目にふざけて付けた。マスコットはうさぎの **Glenda**。作者は Renée French — **後に Go の Gopher を描くことになる同じ人**である。

思想は「UNIX の『すべてはファイル』を極限まで」。ネットワーク接続も、画面上のウィンドウも、よそのマシンの CPU さえもファイルとして見える。OS 自体は天下を取れなかったが、アイデアの方は大脱走した: Linux の `/proc`、コンテナの祖先にあたる名前空間、WSL2 や VM でファイル共有に使われる 9P プロトコル、そして極めつけは **UTF-8** — いま世界中の文字がやりとりできるあのエンコーディングは、1992 年に Ken Thompson と Rob Pike が Plan 9 のためにダイナーの紙ナプキンで設計したものだ。

そしてこの 2 人 + Robert Griesemer が 2007 年に Google で作ったのが **Go**。初期の Go コンパイラは Plan 9 の C コンパイラの直系で、Go のアセンブラは今も Plan 9 流の文法、goroutine の源流も Pike が Plan 9 時代に作った言語群 (Newsqueak/Alef/Limbo) にある。つまり Go にとって Plan 9 は実家であり、2026 年の今も公式ポートとして生かされ続けている。

### で、Errno がどうした？

OS への依頼（システムコール）が失敗すると、ほとんどの OS は**番号**で理由を返す。これが **errno**（2 = ファイルがない、13 = 権限がない、…）で、Go では `syscall.Errno` 型。ところが Plan 9 だけは番号ではなく**文字列**で理由を返す流儀を貫いており（ここでも独自路線）、Go の Plan 9 版には `syscall.Errno` が概念ごと存在しなかった。結果、「errno を見て分岐する」移植性コードは Plan 9 向けビルドだけコンパイルが壊れ、`//go:build !plan9` で隔離する必要があった。

1.27 は「Plan 9 でも型としては定義する。ただし実際に返ることはない」という割り切りでこれを解消した。実家の流儀は変えず、外の世界のコードも通す。

### 何が嬉しい？

- `errors.As(err, &errno)` で EINTR リトライや ENOENT 判定をする定番コードから、Plan 9 隔離用の build tag が消せる。
- 実務インパクトは今回の網羅表で最小級。ただ「**実際には使われない型を定義してでも、移植性コードのビルドを守る**」という Go の互換性哲学の一番小さな見本で、故郷の OS を 2026 年でも見捨てない話としても味わい深い。

### 実測

```console
$ go run ./17-plan9-errno                              # darwin で実行
errno=2 (no such file or directory)

$ GOOS=plan9 GOARCH=amd64 go build ./17-plan9-errno    # 1.27: コンパイル成功
$ GOOS=plan9 GOARCH=amd64 go1.26.5 build .             # 1.26 実測:
./main.go:12:20: undefined: syscall.Errno              # ← ビルド不能だった
```

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
| net/http/httptest NewTestServer | [10](10-synctest/) / [15](15-httptest/) |
| net/url URL.Clone / Values.Clone | [04](04-stdlib-bits/) |
| runtime/secret: secret モードの goroutine 継承 | [16](16-secret/) |
| strings.CutLast | [04](04-stdlib-bits/) |
| syscall: Plan 9 の Errno 定義 | [17](17-plan9-errno/) |
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
