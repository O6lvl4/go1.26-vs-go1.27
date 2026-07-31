module tidydemo

go 1.27rc2

require github.com/google/uuid v1.6.0

// マージ衝突や手編集で増えがちな 2 個目の require ブロック。
// go 1.27 以降の module では go mod tidy が direct/indirect の 2 ブロックに統合する
require (
	github.com/cloudflare/circl v1.6.4
)
