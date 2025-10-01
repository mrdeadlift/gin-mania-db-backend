# 技術スタック確認

## バックエンド

- 言語: Go 1.21 予定（Gin フレームワーク利用）。
- Web フレームワーク: Gin + middleware（ログ、CORS、認証）。
- API 形式: REST を基本、将来的な GraphQL 拡張も検討。
- テスト: Go 標準 testing + testify、integration テストは docker-compose を活用。

## データベース

- メイン: PostgreSQL 15。
- 拡張: PostGIS（地域情報検索用）。
- マイグレーション: golang-migrate。
- OR マッパー/クエリ: GORM。

## キャッシュ・セッション

- Redis（検索結果キャッシュ、レート制限、セッション管理）。

## 認証基盤

- Auth0 等の IDaaS と連携、JWT 発行。
- 開発環境ではスタブ実装／シークレット管理を Vault/Secrets Manager で統一。

## インフラ

- コンテナ: Docker + docker-compose（ローカル）。
- デプロイ: AWS ECS Fargate または GCP Cloud Run（コスト比較中）。
- IaC: Terraform or Pulumi で環境定義。

## 運用基盤

- ロギング: CloudWatch Logs / Stackdriver。
- モニタリング: Prometheus + Grafana or Datadog。
- アラート: PagerDuty / Opsgenie 連携。

## フロントエンド（MVP）

- React + TypeScript、UI フレームワークに Chakra UI。
- API クライアント: React Query。
- 静的ホスティング: Vercel or CloudFront + S3。

## 開発支援ツール

- Lint: golangci-lint, ESLint。
- フォーマッタ: gofmt, Prettier。
- コード品質: SonarCloud 検討。
- ドキュメント: OpenAPI + Storybook。

## 決定待ち項目

- クラウドベンダー最終決定、監視ツールのコスト比較。
