---
allowed-tools: Bash(gh pr comment:*),Bash(gh pr diff:*),Bash(gh pr view:*),Bash(gh pr list:*),Bash(gh api:*),Bash(jq:*),WebFetch,mcp__github_inline_comment__create_inline_comment
description: 5つの観点からPRを包括的にレビューする（Go プロジェクト特化）
argument-hint: [owner/repo] [pr-number]
---

## ステップ 1: プロジェクトルールの読み込み

CLAUDE.md を Read ツールで読み込み、設計方針を把握する:
- `internal/` パッケージへの実装集約ルール（`cmd/` はコマンド定義のみ）
- errcheck 対応ルール（エラー無視は `_ =` で明示）
- stdout/stderr 分離ルール

## ステップ 2: PR 情報の取得

引数から REPO と PR_NUMBER を取得:
```bash
gh pr view $PR_NUMBER --repo $REPO --json title,body,baseRefName,headRefName,files
gh pr diff $PR_NUMBER --repo $REPO
```

## ステップ 3: 5つのサブエージェントを並列実行

以下を並列起動。各エージェントは CLAUDE.md のルールと PR diff を前提知識として受け取る。

### code-quality-reviewer
- `cmd/` にビジネスロジックが混入していないか
- errcheck 対応（エラー無視は `_ =` で明示）
- エラーラップに `fmt.Errorf("context: %w", err)` パターン
- 命名規則（Go 慣習）
- DRY/単一責任原則

### performance-reviewer
- `resp.Body` の `Close()` 漏れ（リソースリーク）
- ページネーション処理のメモリ効率
- `context.Context` の伝搬
- HTTP クライアントのタイムアウト設定

### test-coverage-reviewer
- 新しい `internal/` 関数のテスト有無
- テーブルドリブンテストパターン
- エラーパスのテスト
- テスト間の状態共有がないか

### documentation-accuracy-reviewer
- cobra コマンドの `Use`/`Short`/`Long` と実際の動作の整合性
- フラグの説明文の正確性
- README との整合性（新機能追加時）

### security-code-reviewer
- JWT 秘密鍵・API トークンのログ漏れ
- エラーメッセージへの認証情報混入
- 入力値バリデーション（App ID 等）

各エージェントは問題発見時に `mcp__github_inline_comment__create_inline_comment` で該当行にコメント。

## ステップ 4: フィードバック統合・PR コメント投稿

```bash
gh pr comment $PR_NUMBER --repo $REPO --body "..."
```

コメント構成:
1. レビュー総評
2. 重要度別サマリー（Critical / Warning / Info）
3. Go プロジェクト規約チェックリスト
   - [ ] `internal/` への実装集約
   - [ ] errcheck 対応
   - [ ] stdout/stderr 分離
4. 各エージェントの主要指摘（インライン済みは除く）
5. 承認 / 要修正の推奨
