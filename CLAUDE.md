# CLAUDE.md

## プロジェクト概要

App Store Connect API の CLI ツール（Go 1.25 / cobra）

## 開発コマンド

```bash
# ビルド
go build ./...

# テスト（全パッケージ）
go test ./...

# テスト（詳細・レース検出・カバレッジ付き）
go test ./... -v -race -coverprofile=coverage.out

# lint（push 前に必ず実行すること）
golangci-lint run

# vet
go vet ./...
```

## 開発ワークフロー

1. `main` からブランチを切る
2. コードを修正する
3. `go test ./...` でテストが通ることを確認
4. `golangci-lint run` で lint エラーがないことを確認
5. コミット・push する
6. PR を作成し、CI（GitHub Actions）が通ることを確認

## CI（GitHub Actions）

`.github/workflows/ci.yml` で以下の3ジョブが実行される:

- **test**: ビルド + テスト（race detector・カバレッジ付き）
- **lint**: golangci-lint v2（errcheck 等の静的解析）
- **vet**: go vet

トリガー: `main` への push / PR

## プロジェクト構成

```
cmd/           CLI コマンド定義（cobra）
internal/
  api/         App Store Connect API クライアント
  auth/        JWT トークン生成
  client/      リトライ・認証付き HTTP クライアント
  config/      設定の読み込み・保存
  output/      JSON/CSV 出力フォーマット
```

## コーディング規約

- エラー戻り値は必ずチェックする（errcheck 対応）
  - 無視する場合は `_ =` で明示する
  - `defer f.Close()` → `defer func() { _ = f.Close() }()`
