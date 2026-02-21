# CLAUDE.md

## WHAT: プロジェクト概要

App Store Connect API の CLI ツール（Go 1.25 / cobra）。

個人開発者が App Store の売上・レビュー・アナリティクスデータを取得し、
Claude Code などの AI ツールにパイプして分析・グロースに活用するためのツール。

- データは **stdout**、エラー・ステータスは **stderr** に出力（パイプ連携前提）
- 認証は JWT（ES256）で行い、トークンは自動キャッシュ・更新

```
cmd/           CLI コマンド定義（cobra）
internal/
  api/         App Store Connect API クライアント
  auth/        JWT トークン生成
  client/      リトライ・認証付き HTTP クライアント
  config/      設定の読み込み・保存
  output/      JSON/CSV 出力フォーマット
```

## WHY: 設計方針

- **internal/ に実装を閉じる**: cmd/ はコマンド定義のみ、ビジネスロジックは internal/ に置く
- **エラーは握りつぶさない**: errcheck 対応。無視する場合は `_ =` で明示する
- **stdout/stderr の分離**: jq やファイルリダイレクトと組み合わせやすくする

## HOW: 検証コマンド

```bash
go build ./...
go test ./...
golangci-lint run   # push前に必ず実行
go vet ./...
```

## ブランチ戦略（Git Flow）

- `develop` からブランチを切る（`feat/`, `fix/`, `docs/` など）
- PR は `develop` へ向ける
- リリース時のみ `release/vX.X.X` → `main` へマージしてタグを打つ

詳細は [CONTRIBUTING.md](CONTRIBUTING.md) を参照。
