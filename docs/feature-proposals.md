# Feature Proposals for appc

現在の実装（`apps`, `sales`, `reviews`, `analytics`, `configure`）を踏まえた機能提案。
優先度は **P0**（すぐに欲しい）> **P1**（あると便利）> **P2**（将来的に）の3段階。

---

## 1. データ取得系（新しい API エンドポイント対応）

### 1-1. `appc subscriptions` — サブスクリプション情報の取得 [P0]

App Store Connect API v1 の `/v1/subscriptionGroups` および `/v1/subscriptions` を利用し、
サブスクリプショングループ・個別プランの一覧を取得する。

```
appc subscriptions --app <APP_ID>
appc subscriptions --app <APP_ID> --group <GROUP_ID>
```

**出力フィールド例**: group_id, group_name, subscription_id, name, product_id, state

**理由**: サブスクアプリの個人開発者にとって、プラン構成の把握は最重要。
売上データと組み合わせてプラン別の収益分析に使える。

---

### 1-2. `appc finance` — 財務レポートの取得 [P0]

`/v1/financeReports` エンドポイントを利用して、支払い・財務レポートを取得する。

```
appc finance --date 2026-01 --vendor <VENDOR> --region ZZ
appc finance --from 2025-07 --to 2026-01
```

**出力フィールド例**: region, currency, units, amount, proceeds, exchange_rate

**理由**: `sales` は販売データだが、実際の入金額（Apple 手数料控除後）の確認には
財務レポートが必要。確定申告や収益管理に直結する。

---

### 1-3. `appc versions` — アプリバージョン・ビルド情報 [P1]

`/v1/apps/{id}/appStoreVersions` を利用して、バージョン履歴とレビュー状態を取得する。

```
appc versions --app <APP_ID>
appc versions --app <APP_ID> --state READY_FOR_SALE
```

**出力フィールド例**: version, state, release_type, created_date, platform

**理由**: バージョンのリリース状況を CLI から確認でき、リリース自動化スクリプトに組み込める。
「今どのバージョンが審査中か」を素早く把握できる。

---

### 1-4. `appc iap` — アプリ内課金の一覧 [P1]

`/v1/apps/{id}/inAppPurchasesV2` を利用して、IAP の一覧と状態を取得する。

```
appc iap --app <APP_ID>
appc iap --app <APP_ID> --state APPROVED
```

**出力フィールド例**: id, name, product_id, type (CONSUMABLE/NON_CONSUMABLE/AUTO_RENEWABLE), state

**理由**: IAP の構成確認・棚卸しに使える。売上データと product_id で JOIN して
商品別の収益分析が可能。

---

### 1-5. `appc pricing` — 価格帯・地域別価格の取得 [P2]

`/v1/apps/{id}/appPriceSchedule` で価格設定を取得する。

```
appc pricing --app <APP_ID>
appc pricing --app <APP_ID> --territory JP
```

**出力フィールド例**: territory, currency, price, proceeds

**理由**: 地域別の価格設定を一覧で把握でき、価格改定の判断材料になる。

---

## 2. データ分析・利便性向上

### 2-1. `appc digest` — 日次ダイジェスト生成 [P0]

売上・レビュー・アナリティクスを1コマンドでまとめて取得し、AI に渡しやすい
構造化サマリーを出力する。

```
appc digest --app <APP_ID> --date 2026-02-20
appc digest --app <APP_ID> --period last-7d
```

**出力例** (JSON):
```json
{
  "date": "2026-02-20",
  "sales": { "units": 42, "proceeds": 12800, "currency": "JPY" },
  "reviews": { "total": 5, "average": 4.2, "new_count": 2 },
  "analytics": { "impressions": 1500, "downloads": 42, "conversion": 0.028 }
}
```

**理由**: 個人開発者が毎朝「昨日どうだった？」を1コマンドで把握し、
そのまま `| claude` にパイプして分析できる。このツールの核心的ユースケース。

---

### 2-2. 日付プリセット `--period` フラグ [P0]

全コマンドに `--period` フラグを追加し、よく使う日付範囲を簡単に指定できるようにする。

```
appc sales --period yesterday
appc sales --period last-week
appc sales --period last-month
appc sales --period last-7d
appc sales --period last-30d
appc reviews --app <APP_ID> --period last-7d
```

**実装**: `internal/sales/daterange.go` を拡張して、プリセット名から
`from`/`to` を自動計算する関数を追加する。

**理由**: 毎回 `--from 2026-02-14 --to 2026-02-20` と入力するのは面倒。
特にシェルスクリプトや cron での定期実行で便利。

---

### 2-3. `appc compare` — 期間比較 [P1]

2つの期間の売上・レビューを比較し、差分を出力する。

```
appc compare sales --period last-7d --vs prev-period
appc compare sales --from 2026-02-01 --to 2026-02-14 --vs 2026-01-18..2026-01-31
```

**出力例**:
```json
{
  "current":  { "units": 300, "proceeds": 90000 },
  "previous": { "units": 250, "proceeds": 75000 },
  "change":   { "units": "+20.0%", "proceeds": "+20.0%" }
}
```

**理由**: グロース分析の基本は前期比較。手作業で2回コマンドを叩いて diff する手間を省く。

---

### 2-4. 売上の集計オプション `--group-by` [P1]

`appc sales` に集計機能を追加する。

```
appc sales --from 2026-01-01 --to 2026-01-31 --group-by country
appc sales --from 2026-01-01 --to 2026-01-31 --group-by date
appc sales --from 2026-01-01 --to 2026-01-31 --group-by product
```

**理由**: 大量の生データを返すよりも、国別・日別・商品別のサマリーの方が
AI に渡すトークン数も減り、分析しやすい。

---

## 3. 出力・フォーマット

### 3-1. Markdown テーブル出力 `--format table` [P0]

人間が直接読む場合や、Markdown ドキュメントに貼り付ける場合に使える
テーブル形式の出力を追加する。

```
appc apps --format table
```

```
| ID         | Name       | Bundle ID          | SKU    |
|------------|------------|--------------------|--------|
| 1234567890 | MyApp      | com.example.myapp  | MYAPP1 |
```

**実装**: `internal/output/formatter.go` に `WriteTable` を追加。
`CSVRecord` インターフェースをそのまま利用できる。

**理由**: ターミナルで直接確認したいケースは多い。
JSON は jq なしだと読みにくく、CSV もカラムがずれる。

---

### 3-2. `--output` フラグ（ファイル出力） [P1]

stdout をリダイレクトする代わりに、直接ファイルに書き出すフラグ。

```
appc sales --date 2026-02-20 --output sales-0220.json
appc sales --from 2026-01-01 --to 2026-01-31 --output sales-jan.csv --format csv
```

**理由**: stdout への出力を基本とし、`--output` はシェルリダイレクト（`> file.json`）の
薄いラッパーとして位置づける。CLAUDE.md の「データは stdout、ステータスは stderr」方針を維持する。
`--output auto` の自動命名は予測不可能な副作用を生むため、採用しない。

---

### 3-3. TSV 出力 `--format tsv` [P2]

スプレッドシートにコピペしやすい TSV 形式を追加する。

**理由**: CSV はカンマを含むフィールドでクォートが必要だが、TSV はそのまま
Google Sheets / Excel にペーストできる。

---

## 4. 開発者体験（DX）

### 4-1. シェル補完 `appc completion` [P0]

Cobra の組み込み機能を利用して、bash/zsh/fish/powershell の補完スクリプトを生成する。

```
appc completion bash > /etc/bash_completion.d/appc
appc completion zsh > "${fpath[1]}/_appc"
```

**実装**: Cobra の `GenBashCompletion` 等を呼ぶだけでほぼ完成。
カスタム補完（`--app` に既存アプリ ID を候補表示）も可能。

**理由**: CLI ツールとして基本的な DX 向上。フラグ名を覚えなくてよくなる。

---

### 4-2. `appc profile` — マルチアカウント対応 [P1]

複数の App Store Connect アカウントを切り替えて使えるようにする。

```
appc profile add personal --issuer-id XXX --key-id YYY --key-path ~/.keys/p.p8
appc profile add work --issuer-id AAA --key-id BBB --key-path ~/.keys/w.p8
appc profile use personal
appc sales --date 2026-02-20 --profile work
```

**設定ファイル**: `~/.config/appc/config.json` を拡張。

```json
{
  "current_profile": "personal",
  "profiles": {
    "personal": { "issuer_id": "...", "key_id": "...", ... },
    "work": { "issuer_id": "...", "key_id": "...", ... }
  }
}
```

**理由**: 受託開発者やチームで複数アカウントを管理する場合に必須。

**セキュリティ上の注意**:
- `appc profile add` 時に `.p8` ファイルのパーミッションが `0600` でない場合は警告を出す
- `config.json` には `key_path`（ファイルパス）のみを保持し、秘密鍵の内容は保存しない
- 将来的には macOS Keychain / Linux libsecret への保存も検討する

---

### 4-3. `appc doctor` — 診断コマンド [P1]

環境や設定の問題を自動診断する。

```
appc doctor
```

```
[OK] Config file: ~/.config/appc/config.json
[OK] Private key: AuthKey_XXXXX.p8 (valid ECDSA P-256)
[OK] JWT generation: success
[OK] API connection: authenticated (issuer: XXXXXXXX-...)
[WARN] golangci-lint: not found (optional, for development)
```

**理由**: セットアップ時のトラブルシューティングを大幅に簡略化。
API 疎通確認まで一発でできるのは初心者に優しい。

**出力に関する注意**:
- 全出力を `os.Stderr` に向ける（CLAUDE.md の「ステータスは stderr」原則に従う）
- issuer ID は先頭8文字 + `...` に省略する（例: `XXXXXXXX-...`）
- 秘密鍵のパスはファイル名のみ表示し、フルパスは出力しない（バグレポート送信時の情報漏洩を防止）

---

### 4-4. `--verbose` / `--quiet` フラグ [P2]

stderr 出力の詳細度を制御する。

```
appc sales --from ... --to ... --quiet    # stderr 非表示
appc sales --date ... --verbose           # リクエスト/レスポンス詳細を表示
```

**理由**: パイプラインに組み込む際は `--quiet` で余計な出力を抑制。
デバッグ時は `--verbose` で HTTP リクエストの詳細を確認。

**セキュリティ上の注意**:
`--verbose` で HTTP リクエスト/レスポンスの詳細を出力する際、
`Authorization` ヘッダに含まれる JWT Bearer トークンが平文で出力されないようにする。
- `Authorization` ヘッダは `Bearer [REDACTED]` に置換してから出力する
- `internal/client` に `redactHeader(name string) bool` ヘルパーを追加してテストで検証する
- ユーザーが `2>&1 | tee debug.log` でログ保存した場合や CI/CD ログへの漏洩を防止する

---

## 5. AI 連携強化

### 5-1. `appc prompt` — AI 向けプロンプト生成 [P1]

データ取得結果に分析指示を添えたプロンプトを生成し、そのまま AI にパイプできるようにする。

```
appc prompt growth --app <APP_ID> --period last-30d | claude
appc prompt reviews --app <APP_ID> --period last-7d | claude
```

**出力例**:
```
以下は App Store Connect から取得した直近30日間の売上データです。

[JSON データ]

このデータを分析して、以下の観点でレポートを作成してください：
1. 売上トレンドと前月比較
2. 国別の売上構成比
3. 改善のための具体的なアクション提案
```

**理由**: appc のコンセプト「AI ツールにパイプして分析」を最も直接的に実現する機能。
データ取得 → プロンプト生成 → AI 分析のパイプラインをワンライナーで完結させる。

---

### 5-2. レビューのセンチメント分類フィールド [P2]

`appc reviews` の出力にレビュー本文から推定したセンチメント
（positive/negative/neutral）を付加する。

```
appc reviews --app <APP_ID> --sentiment
```

**理由**: レビュー数が多い場合に、ネガティブレビューだけ抽出して
改善点を洗い出すワークフローに使える。ただし外部 AI 呼び出しが必要なため
オプトインにする。

---

## 6. 自動化・運用

### 6-1. `appc watch` — 定期ポーリング [P2]

指定間隔でデータを取得し続け、変更があれば通知する。

```
appc watch reviews --app <APP_ID> --interval 1h
appc watch sales --date yesterday --interval 6h
```

> **注**: Apple の Sales Reports API は日次レポートを当日深夜（Pacific Time）以降に
> 翌日公開するため、`--date today` は機能しない。常に前日以前のデータを指定する。

**理由**: cron を設定しなくても、ターミナルを開いたまま新着レビューを
監視できる。Webhook 連携（Slack 通知等）の基盤にもなる。

**実装上の必須要件**:
- `signal.NotifyContext` で `SIGINT`/`SIGTERM` を受け取りコンテキストをキャンセルする
- 既存 `client.go` は `context.Context` を受け取る設計のため、ポーリングループもこのコンテキストを引き継ぐ
- ポーリング間隔は「前回完了からN分後」方式とし、定義を明確にする

```go
// 推奨パターン
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

---

### 6-2. キャッシュ機構 [P1]

過去のレポートデータをローカルにキャッシュし、同じリクエストの再実行を高速化する。

```
appc sales --date 2026-01-15          # API 呼び出し（キャッシュ保存）
appc sales --date 2026-01-15          # キャッシュから即座に返す
appc sales --date 2026-01-15 --fresh  # キャッシュを無視して再取得
```

**保存先**: `${XDG_CACHE_HOME:-~/.cache}/appc/`（XDG Base Directory 準拠）

**理由**: 過去の売上データはレポート確定後は不変なので、キャッシュとの相性が良い。
日付範囲の一括取得（`--from --to`）で API レート制限に引っかかるリスクも軽減できる。

**キャッシュ設計の注意点**:
- **データ確定タイミング**: Apple のレポートは公開後 24〜48 時間は未確定の場合がある。未確定期間中のデータには TTL を設定し、自動的に再取得する
- **キャッシュキー**: エンドポイント・日付・vendor 番号・レポートタイプなど全パラメータをキーに含める（不足するとコリジョン、過剰だとヒットしない）
- **アトミック書き込み**: クラッシュ時の破損防止のため `os.Rename` による一時ファイル経由で書き込む
- **`--fresh` フラグ**: キャッシュを無視して再取得するオプション

---

## 実装優先順についてのまとめ

| 優先度 | 提案 | カテゴリ |
|--------|------|----------|
| P0 | `appc digest` (日次ダイジェスト) | 分析 |
| P0 | `--period` プリセット | 利便性 |
| P0 | `--format table` | 出力 |
| P0 | `appc completion` (シェル補完) | DX |
| P0 | `appc finance` (財務レポート) | データ取得 |
| P0 | `appc subscriptions` | データ取得 |
| P1 | `appc compare` (期間比較) | 分析 |
| P1 | `--group-by` 集計 | 分析 |
| P1 | `appc prompt` (AI 向けプロンプト) | AI 連携 |
| P1 | `appc versions` | データ取得 |
| P1 | `appc iap` | データ取得 |
| P1 | `appc profile` (マルチアカウント) | DX |
| P1 | `appc doctor` (診断) | DX |
| P1 | `--output` フラグ | 出力 |
| P1 | キャッシュ機構 | 運用 |
| P2 | `appc pricing` | データ取得 |
| P2 | `appc watch` (定期ポーリング) | 運用 |
| P2 | `--verbose`/`--quiet` | DX |
| P2 | `--format tsv` | 出力 |
| P2 | レビューセンチメント分類 | AI 連携 |
