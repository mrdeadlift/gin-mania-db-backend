# 開発環境整備

## ローカル環境要件
- OS: macOS / Linux / Windows WSL2。
- Go 1.21、Node.js 20、Yarn or pnpm。
- Docker Desktop（またはPodman）インストール。
- make, git, golangci-lint。

## セットアップ手順（案）
1. リポジトリクローン後、`make setup`で以下を自動実行。
   - Go modules取得 (`go mod download`)
   - フロントエンド依存インストール (`yarn install`)
   - pre-commit hooks設定（lint、fmt）。
2. `docker-compose up`でDB・Redis起動。
3. サーバー起動: `make run`（ホットリロードはair導入）。
4. フロント起動: `yarn dev`。

## 環境変数とシークレット
- `.env.example`を配布し、`GIN_DB_URL`、`REDIS_URL`、`AUTH0_*`設定。
- シークレットはローカルではdirenv/1password-cli、クラウドはSecrets Managerで管理。

## テスト実行
- 単体テスト: `make test`
- 統合テスト: docker-compose上のサービスを利用し`make test-integration`
- フロントE2E: Playwrightで`yarn test:e2e`

## 品質ゲート
- PR作成時にGitHub Actionsでlint/testを自動実行。
- mainブランチへのマージはレビュー1名以上必須。

## 開発フロー
- GitHub Flowを基本（feature branch -> PR -> review -> merge）。
- Issueテンプレートでタスク粒度を統一。

## ドキュメント整備
- /doc配下に設計・運用ドキュメントを集約。
- OpenAPI生成は`make docs`で更新し、API仕様を同期。
