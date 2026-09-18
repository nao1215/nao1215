## OSS I maintain

メンテナンスを続けていて、他の人にも勧められるものだけを載せています。どれも MIT または Apache-2.0 で、README に導入手順があります。Stars は毎日自動で更新されます。

最初に試すなら次の4つです。

- [gup](https://github.com/nao1215/gup): `go install` で入れたツールを一括で更新し、別のマシンへ移せる
- [sqly](https://github.com/nao1215/sqly): CSV・Excel・JSON などのファイルに、対話シェルから SQL を実行できる
- [filesql](https://github.com/nao1215/filesql): ファイルを SQLite に読み込み、Go の `database/sql` で問い合わせられる
- [atago](https://github.com/nao1215/atago): CLI の振る舞いを YAML だけでテストできる。対話端末（PTY/TUI）も検証できる

### Go CLI

| OSS | 何ができるか | 導入 | Stars |
| :--- | :--- | :--- | ---: |
| [gup](https://github.com/nao1215/gup) | `$GOBIN` のバイナリを並列に更新する。一覧の書き出しと読み込みで、別のマシンへ同じツール群を移せる | `brew install nao1215/tap/gup` | 602 |
| [sqly](https://github.com/nao1215/sqly) | CSV・TSV・LTSV・JSON・JSONL・Parquet・Excel・ACH・Fedwire のファイルに SQL を実行する対話シェル。形式の違うファイル同士を JOIN でき、結果を書き戻せる | `brew install nao1215/tap/sqly` | 178 |
| [atago](https://github.com/nao1215/atago) | CLI のテストを YAML で書く。終了コード・出力・ファイル・スナップショット・対話端末を検証し、実行の記録からテストを作れる | `brew install --cask nao1215/tap/atago` | 20 |
| [jsonize](https://github.com/nao1215/jsonize) | コマンド出力・ファイル・引数から JSON を作る（コマンド名は `jz`）。`COMMAND \| jz \| jq` のようにパイプラインにつなげる | `brew install --cask nao1215/tap/jsonize` | 5 |
| [himorime](https://github.com/nao1215/himorime) | CLI の性能予算と性能劣化の検査を YAML で書く。実行時間・スループット・CPU 時間・メモリを測る | `brew install --cask nao1215/tap/himorime` | 1 |
| [mobilepkg](https://github.com/nao1215/mobilepkg) | APK・AAB・IPA などを SDK なしで検査し、メタデータ・署名・版の差分・セキュリティ上の指摘を取り出す。Go ライブラリとしても使える | `go install github.com/nao1215/mobilepkg/cmd/mobilepkg@latest` | 6 |
| [jose](https://github.com/nao1215/jose) | 鍵の生成（JWK）、署名と検証（JWS）、暗号化と復号（JWE）をシェルから行う | `brew install nao1215/tap/jose` | 15 |
| [iso8583tool](https://github.com/nao1215/iso8583tool) | 決済で使う ISO 8583 メッセージを解析・確認するデバッグ用 CLI | `go install github.com/nao1215/iso8583tool@latest` | 5 |
| [block](https://github.com/nao1215/block) | ブロックチェーン開発で使う CLI の版をロックファイルで固定し、開発者と CI でそろえる | `brew install --cask nao1215/tap/block` | 2 |
| [omokage](https://github.com/nao1215/omokage) | 文章が特定の書き手の文体にどれだけ近いかを測る。日本語と英語に対応し、手元だけで動く | `go install github.com/nao1215/omokage@latest` | 17 |
| [career](https://github.com/nao1215/career) | 1つの YAML から、日本語の履歴書・職務経歴書と英語の CV を PDF で作る | `go install github.com/nao1215/career@latest` | 4 |
| [mimixbox](https://github.com/nao1215/mimixbox) | BusyBox に倣った Linux 向けのコマンド集。1つのバイナリに多数のコマンドを収める | `brew install nao1215/tap/mimixbox` | 39 |
| [onionscan](https://github.com/nao1215/onionscan) | Tor の onion サービスを調査し、設定ミスや情報漏えいの手がかりを探す | `brew install nao1215/tap/onionscan` | 18 |

### Go ライブラリ

| OSS | 何ができるか | 導入 | Stars |
| :--- | :--- | :--- | ---: |
| [filesql](https://github.com/nao1215/filesql) | CSV・JSON・Parquet・Excel などを SQLite に読み込み、`database/sql` で問い合わせる。圧縮ファイルも読め、前処理とバリデーションの `prep` パッケージが付く | `go get github.com/nao1215/filesql` | 383 |
| [markdown](https://github.com/nao1215/markdown) | Markdown と Mermaid 図を Go のコードから組み立てる。v1 で API が安定している | `go get github.com/nao1215/markdown` | 141 |
| [prompt](https://github.com/nao1215/prompt) | 補完と履歴を持つ対話型プロンプト・REPL を作る。go-prompt の後継を目指している | `go get github.com/nao1215/prompt` | 10 |
| [tornago](https://github.com/nao1215/tornago) | Go から Tor を使うクライアント・サーバーライブラリ | `go get github.com/nao1215/tornago` | 8 |
| [imaging](https://github.com/nao1215/imaging) | リサイズ・切り抜き・回転・ぼかしなどの画像処理。disintegration/imaging のフォークを保守している | `go get github.com/nao1215/imaging` | 10 |
| [sensitive](https://github.com/nao1215/sensitive) | テキスト中のクレジットカード番号やメールアドレスなどを検出し、必要ならマスクする | `go get github.com/nao1215/sensitive` | 2 |

### Gleam パッケージ

すべて Hex で公開しています。多くは Erlang と JavaScript の両方で動きます。

| OSS | 何ができるか | 導入 | Stars |
| :--- | :--- | :--- | ---: |
| [oaspec](https://github.com/nao1215/oaspec) | OpenAPI 3.x の仕様から、型付きのサーバースタブとクライアント SDK を生成する | `gleam add oaspec` | 8 |
| [sqlode](https://github.com/nao1215/sqlode) | SQL のスキーマとクエリから型安全なコードを生成する（sqlc 風）。PostgreSQL・MySQL・SQLite に対応 | `gleam add sqlode` | 5 |
| [metamon](https://github.com/nao1215/metamon) | プロパティベーステストとメタモルフィックテストを書くためのライブラリ | `gleam add --dev metamon` | 5 |
| [dataprep](https://github.com/nao1215/dataprep) | 入力の前処理とバリデーションを小さな部品の組み合わせで書く | `gleam add dataprep` | 5 |
| [datastream](https://github.com/nao1215/datastream) | 遅延評価のストリームを組み合わせて処理する。資源の後始末まで扱う | `gleam add datastream` | 5 |
| [ssevents](https://github.com/nao1215/ssevents) | Server-Sent Events のイベントを組み立て・解析する | `gleam add ssevents` | 2 |
| [multipartkit](https://github.com/nao1215/multipartkit) | multipart 形式（ファイルアップロードなど）を解析・生成する | `gleam add multipartkit` | 3 |
| [mimetype](https://github.com/nao1215/mimetype) | 拡張子から MIME タイプを引き、ファイル先頭のバイト列から種類を判定する | `gleam add mimetype` | 3 |
| [yabase](https://github.com/nao1215/yabase) | Base64・Base32 などの符号化を1つの型安全な API で扱う | `gleam add yabase` | 4 |
| [automata](https://github.com/nao1215/automata) | cron と RRULE の予定計算、リトライ、ファイルシステムのイベント、有限オートマトン | `gleam add automata` | 4 |
| [textmetrics](https://github.com/nao1215/textmetrics) | 編集距離・類似度・最長共通部分列（LCS）・diff を計算する | `gleam add textmetrics` | 5 |
| [finanza](https://github.com/nao1215/finanza) | 10 進数演算、通貨の書式化、金融計算、カード番号の検証 | `gleam add finanza` | 4 |
| [geokit](https://github.com/nao1215/geokit) | 2 点間の距離・方位・geohash・polyline・Web メルカトルなどの地理計算 | `gleam add geokit` | 7 |
| [qrkit](https://github.com/nao1215/qrkit) | QR コードを生成し、端末表示・SVG・PNG で出力する | `gleam add qrkit` | 11 |
| [sparklinekit](https://github.com/nao1215/sparklinekit) | 小さな折れ線グラフを、端末向けの文字・SVG・PNG で描く | `gleam add sparklinekit` | 1 |
| [packkit](https://github.com/nao1215/packkit) | gzip・zstd などの圧縮と、tar・zip などのアーカイブを外部依存なしで扱う。まだ発展途上 | `gleam add packkit` | 2 |

### Rust

| OSS | 何ができるか | 導入 | Stars |
| :--- | :--- | :--- | ---: |
| [truss](https://github.com/nao1215/truss) | 画像変換を CLI・HTTP サーバー・WASM から使える。署名付き URL と SSRF 対策を備え、AVIF・WebP・SVG に対応する | `cargo install truss-image` | 9 |

CI で CLI を入れるための GitHub Action も用意しています: [setup-atago](https://github.com/nao1215/setup-atago)、[setup-himorime](https://github.com/nao1215/setup-himorime)、[setup-iso8583tool](https://github.com/nao1215/setup-iso8583tool)、[setup-block](https://github.com/nao1215/setup-block)。v0 系のものは、マイナーバージョンの更新で API や設定形式が変わることがあります。
