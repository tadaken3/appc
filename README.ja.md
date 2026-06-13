# appc - App Store Connect CLI

[English](README.md) | [日本語](README.ja.md)

個人開発者が App Store Connect のデータ（売上レポート・レビュー・アナリティクス）を取得し、Claude Code などの AI ツールにパイプしてアプリのグロースに活用するための CLI ツールです。

## インストール

### Homebrew（推奨）

```bash
brew tap tadaken3/appc
brew install appc
```

### go install

```bash
go install github.com/tadaken3/appc@latest
```

### ソースからビルド

```bash
git clone https://github.com/tadaken3/appc.git
cd appc
go build -o appc .
```

## セットアップ

### 1. API キーの作成

1. [App Store Connect](https://appstoreconnect.apple.com/) を開き、**ユーザとアクセス > 統合 > App Store Connect API** に移動します。
2. **API キーを生成** をクリックし、役割（管理者、財務など）を選択します。
3. `.p8` 秘密鍵ファイルをダウンロードします。**ダウンロードは一度しかできません。**
4. ページに表示される **Key ID** と **Issuer ID** を控えておきます。

### 2. appc の設定

対話形式のセットアップを実行します:

```bash
appc configure
```

以下の項目を入力します:

| 項目 | 説明 |
|---|---|
| **Issuer ID** | App Store Connect > ユーザとアクセス > キー に表示される UUID |
| **Key ID** | 作成した API キーの英数字 ID |
| **Private Key Path** | `.p8` ファイルのパス（`~` 展開に対応） |
| **Vendor Number** | 売上レポートに使用するベンダー番号 |

設定は `~/.config/appc/config.json` にパーミッション `0600` で保存されます。

### 環境変数（CI・クラウド環境）

`appc configure` の実行や `.p8` ファイルの配置が難しい環境（CI、コンテナ、
Claude Code on the web など）では、設定を環境変数だけで渡すこともできます。
環境変数が設定されている場合は設定ファイルの値を**上書き**するため、設定ファイル
は任意です。

| 環境変数 | 説明 |
|---|---|
| `APPC_ISSUER_ID` | Issuer ID |
| `APPC_KEY_ID` | Key ID |
| `APPC_PRIVATE_KEY` | `.p8` 秘密鍵の PEM 内容（インライン） |
| `APPC_PRIVATE_KEY_PATH` | `.p8` ファイルのパス（`APPC_PRIVATE_KEY` の代替） |
| `APPC_VENDOR_NUMBER` | 売上レポート用のベンダー番号 |

`APPC_PRIVATE_KEY` は `APPC_PRIVATE_KEY_PATH` より優先されます。例:

```bash
export APPC_ISSUER_ID="..."
export APPC_KEY_ID="..."
export APPC_VENDOR_NUMBER="..."
export APPC_PRIVATE_KEY="$(cat AuthKey_XXXX.p8)"
appc configure --validate   # 認証情報が正しく解決できるか確認
```

## コマンド一覧

### `appc apps`

App Store Connect アカウントのアプリ一覧を取得します。iTunes Lookup API から各アプリの App Store 評価を自動的に付与します。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--country` | string | `jp` | 評価取得に使用する国コード |

```bash
appc apps
appc apps --country us --format csv
```

### `appc lookup`

iTunes Lookup API からアプリの評価・メタデータを取得します。複数の App ID をカンマ区切りで指定でき、競合アプリの調査にも使えます（App Store Connect の認証不要）。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--app` | string | *（必須）* | アプリ ID（カンマ区切りで複数指定可） |
| `--country` | string | `jp` | ストア検索に使用する国コード |

```bash
appc lookup --app 6745560143 --country jp
appc lookup --app 6745560143,123456789 --country us    # 競合調査
appc lookup --app 6745560143 --format csv
```

### `appc sales`

売上レポートをダウンロードします。データのない日付（HTTP 404）はエラーではなく空結果として扱われます。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--date` | string | *（必須）* | レポート日付（`YYYY-MM-DD` または `YYYY-MM`） |
| `--from` | string | | 開始日付（`YYYY-MM-DD`、`--to` と併用、日次のみ） |
| `--to` | string | | 終了日付（`YYYY-MM-DD`、`--from` と併用、日次のみ） |
| `--type` | string | `SALES` | レポートタイプ: `SALES`, `PRE_ORDER`, `NEWSSTAND` |
| `--frequency` | string | `DAILY` | 頻度: `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY` |
| `--sub-type` | string | `SUMMARY` | サブタイプ: `SUMMARY`, `DETAILED`, `OPT_IN` |
| `--vendor` | string | *（設定値）* | ベンダー番号（設定値を上書き） |

*\* `--date` または `--from`/`--to` のいずれかが必須です。*

```bash
appc sales --date 2025-01-15
appc sales --date 2025-01-15 --type SALES --frequency DAILY --format csv
appc sales --date 2025-01 --frequency MONTHLY --vendor 12345678
appc sales --from 2025-01-01 --to 2025-01-31
```

### `appc reviews`

アプリのカスタマーレビューを取得します。`--summary` を指定すると、iTunes の全体評価を含む評価分布サマリーを表示します。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--app` | string | *（必須）* | アプリ ID |
| `--rating` | int | `0` | 星評価でフィルタ（1〜5、0 = 全件） |
| `--limit` | int | `100` | 取得するレビューの最大件数 |
| `--summary` | bool | `false` | 個別レビューの代わりに評価サマリーを表示 |
| `--country` | string | `jp` | 評価取得に使用する国コード（`--summary` 時に使用） |

```bash
appc reviews --app <APP_ID>
appc reviews --app <APP_ID> --rating 5 --limit 50
appc reviews --app <APP_ID> --summary                    # App Store 全体評価付き
appc reviews --app <APP_ID> --summary --country us
appc reviews --app <APP_ID> --format csv > reviews.csv
```

### `appc analytics`

App Store のアナリティクスレポートをリクエスト・取得します。セグメント CSV をダウンロードしてパースした結果を出力します。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--app` | string | *（必須）* | アプリ ID |
| `--category` | string | | レポートカテゴリ |
| `--snapshot` | bool | `false` | ONE_TIME_SNAPSHOT で過去データを一括取得 |

利用可能なカテゴリ: `APP_USAGE`, `APP_STORE_ENGAGEMENT`, `COMMERCE`, `FRAMEWORK_USAGE`, `PERFORMANCE`

```bash
# 最新のアナリティクスを取得（ONGOING モード — リクエスト作成日以降のデータ）
appc analytics --app <APP_ID> --category APP_USAGE
appc analytics --app <APP_ID> --category APP_STORE_ENGAGEMENT --format csv

# 過去データを一括取得（アプリ作成日〜リクエスト日）
appc analytics --app <APP_ID> --category APP_USAGE --snapshot

# Claude にパイプして分析
appc analytics --app <APP_ID> --category APP_USAGE | claude "インストールのトレンドを分析して"
```

> **注意:** 初回実行時は Apple 側でレポートインスタンスの生成に時間がかかります（ONGOING は 1〜2 日、スナップショットは数時間）。2 回目以降はすぐにデータが取得できます。

### `appc configure`

認証情報を対話形式で設定・更新します。既存の値がデフォルトとして表示され、Enter を押すとそのまま保持されます。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--validate` | bool | `false` | 設定を変更せずにバリデーションのみ実行 |

## グローバルフラグ

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--format` | string | `json` | 出力形式: `json` または `csv` |

すべてのコマンドに適用されます。

## 出力

- **データ** は **stdout** に出力されます（パイプ連携に対応）。
- **ステータスメッセージやエラー** は **stderr** に出力されます。

```bash
# jq でフィルタ
appc apps | jq '.[].name'

# CSV にエクスポート
appc sales --date 2025-01-15 --format csv > sales.csv

# Claude Code にパイプして分析
appc reviews --app <APP_ID> --format json | claude "これらのレビューの感情を要約して"
```

## プロジェクト構成

```
appc/
├── main.go                 # エントリーポイント
├── cmd/                    # CLI コマンド定義（cobra）
│   ├── root.go             #   ルートコマンドとグローバルフラグ
│   ├── apps.go             #   appc apps
│   ├── sales.go            #   appc sales
│   ├── reviews.go          #   appc reviews
│   ├── analytics.go        #   appc analytics
│   ├── lookup.go           #   appc lookup
│   └── configure.go        #   appc configure
└── internal/
    ├── api/                # App Store Connect API クライアント
    │   ├── types.go        #   共通レスポンス・ページング型
    │   ├── apps.go         #   アプリ エンドポイント
    │   ├── sales.go        #   売上レポート エンドポイント
    │   ├── reviews.go      #   レビュー エンドポイント
    │   ├── analytics.go    #   アナリティクス エンドポイント
    │   └── lookup.go       #   iTunes Lookup API
    ├── auth/               # JWT（ES256）トークン生成
    │   └── jwt.go
    ├── client/             # リトライ・認証付き HTTP クライアント
    │   └── client.go
    ├── config/             # 設定の読み込み・保存
    │   └── config.go
    └── output/             # JSON/CSV 出力フォーマット
        └── formatter.go
```

## 開発・テスト

```bash
# 全テスト実行
go test ./...

# 詳細出力
go test ./... -v
```

## 認証の仕組み

appc は **ES256（ECDSA P-256 + SHA-256）で署名された JWT** を使用して App Store Connect API に認証します。

- **Issuer ID**、**Key ID**、`.p8` **秘密鍵** からトークンを生成します。
- クレームには `iss`、`iat`、`exp`、`aud`（`appstoreconnect-v1`）が含まれます。
- トークンの **有効期間は 20 分** で、自動的にキャッシュされます。
- 有効期限の 1 分前にキャッシュが更新されます。
- トークン生成はスレッドセーフです。
