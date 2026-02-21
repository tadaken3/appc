# appc - App Store Connect CLI

[English](README.md) | [日本語](README.ja.md)

App Store Connect API からデータを取得するコマンドラインツールです。アプリ一覧、売上レポート、カスタマーレビュー、アナリティクスに対応しています。

## インストール

### go install

```bash
go install github.com/kenta-tanaka/appc@latest
```

### ソースからビルド

```bash
git clone https://github.com/kenta-tanaka/appc.git
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

## コマンド一覧

### `appc apps`

App Store Connect アカウントのアプリ一覧を取得します。

```bash
appc apps
appc apps --format csv
```

### `appc sales`

売上レポートをダウンロードします。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--date` | string | *（必須）* | レポート日付（`YYYY-MM-DD` または `YYYY-MM`） |
| `--type` | string | `SALES` | レポートタイプ: `SALES`, `PRE_ORDER`, `NEWSSTAND` |
| `--frequency` | string | `DAILY` | 頻度: `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY` |
| `--sub-type` | string | `SUMMARY` | サブタイプ: `SUMMARY`, `DETAILED`, `OPT_IN` |
| `--vendor` | string | *（設定値）* | ベンダー番号（設定値を上書き） |

```bash
appc sales --date 2025-01-15
appc sales --date 2025-01-15 --type SALES --frequency DAILY --format csv
appc sales --date 2025-01 --frequency MONTHLY --vendor 12345678
```

### `appc reviews`

アプリのカスタマーレビューを取得します。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--app` | string | *（必須）* | アプリ ID |
| `--rating` | int | `0` | 星評価でフィルタ（1〜5、0 = 全件） |
| `--limit` | int | `100` | 取得するレビューの最大件数 |

```bash
appc reviews --app <APP_ID>
appc reviews --app <APP_ID> --rating 5 --limit 50
appc reviews --app <APP_ID> --format csv > reviews.csv
```

### `appc analytics`

App Store のアナリティクスレポートをリクエスト・取得します。

| フラグ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `--app` | string | *（必須）* | アプリ ID |
| `--category` | string | | レポートカテゴリ（例: `APP_USAGE`, `APP_STORE_ENGAGEMENT`） |

```bash
appc analytics --app <APP_ID>
appc analytics --app <APP_ID> --category APP_USAGE
```

### `appc configure`

認証情報を対話形式で設定・更新します。既存の値がデフォルトとして表示され、Enter を押すとそのまま保持されます。

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
│   └── configure.go        #   appc configure
└── internal/
    ├── api/                # App Store Connect API クライアント
    │   ├── types.go        #   共通レスポンス・ページング型
    │   ├── apps.go         #   アプリ エンドポイント
    │   ├── sales.go        #   売上レポート エンドポイント
    │   ├── reviews.go      #   レビュー エンドポイント
    │   └── analytics.go    #   アナリティクス エンドポイント
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
