# Contributing to appc

## ブランチ戦略（Git Flow）

```
main        本番リリース済みのコード。タグ（v0.0.x）が打たれる。
develop     次のリリースに向けた開発ブランチ。
feature/*   新機能・改善。develop から切って develop にマージ。
docs/*      ドキュメント変更。develop から切って develop にマージ。
fix/*       バグ修正。develop から切って develop にマージ。
hotfix/*    本番の緊急バグ修正。main から切って main と develop にマージ。
release/*   リリース前の最終調整。develop から切って main にマージ。
```

## 開発フロー

### 通常の機能開発・バグ修正

```bash
# 1. develop を最新化
git switch develop
git pull origin develop

# 2. ブランチを切る
git switch -c feat/your-feature   # 機能追加
git switch -c fix/your-bug        # バグ修正
git switch -c docs/your-doc       # ドキュメント

# 3. コードを修正
# 4. テスト・lint を通す
go test ./...
golangci-lint run

# 5. コミット・push
git push origin feat/your-feature

# 6. develop への PR を作成
```

### ホットフィックス（本番の緊急修正）

```bash
# 1. main から切る
git switch main
git switch -c hotfix/critical-bug

# 2. 修正・コミット
# 3. main と develop の両方に PR を作成
```

## リリース手順

### 1. release ブランチを作成

```bash
git switch develop
git pull origin develop
git switch -c release/v0.0.3
```

### 2. バージョンに関する最終調整

- CHANGELOG・バージョン確認など必要な修正をこのブランチで行う

### 3. main にマージ

release ブランチから main への PR を作成・マージする。

### 4. タグを打つ

```bash
git switch main
git pull origin main
git tag v0.0.3
git push origin v0.0.3
```

タグを push すると GitHub Actions（`.github/workflows/release.yml`）が自動的に起動し、以下が実行される：

- macOS/Linux × amd64/arm64 向けバイナリのビルド
- GitHub Releases へのアップロード
- `tadaken3/homebrew-appc` の Formula 自動更新

### 5. develop にもマージ

release ブランチを develop にもマージして差分をなくす。

```bash
git switch develop
git merge release/v0.0.3
git push origin develop
```

## コーディング規約

- エラー戻り値は必ずチェックする（無視する場合は `_ =` で明示）
- `golangci-lint run` でエラーがないことを確認してからコミット

## CI

PR 作成時・main/develop への push 時に以下が自動実行される：

| ジョブ | 内容 |
|---|---|
| test | ビルド + テスト（race detector・カバレッジ付き） |
| lint | golangci-lint |
| vet | go vet |
