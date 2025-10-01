# 技術スタック確認

## バックエンド
- 言語: Go 1.21 予定（Ginフレームワーク利用）。
- Webフレームワーク: Gin + middleware（ログ、CORS、認証）。
- API形式: RESTを基本、将来的なGraphQL拡張も検討。
- テスト: Go標準testing + testify、integrationテストはdocker-composeを活用。

## データベース
- メイン: PostgreSQL 15。
- 拡張: PostGIS（地域情報検索用）。
- マイグレーション: golang-migrate。
- ORマッパー/クエリ: sqlc or GORM（パフォーマンス要件次第で選択）。

## キャッシュ・セッション
- Redis（検索結果キャッシュ、レート制限、セッション管理）。

## 認証基盤
- Auth0等のIDaaSと連携、JWT発行。
- 開発環境ではスタブ実装／シークレット管理をVault/Secrets Managerで統一。

## インフラ
- コンテナ: Docker + docker-compose（ローカル）。
- デプロイ: AWS ECS Fargate または GCP Cloud Run（コスト比較中）。
- IaC: Terraform or Pulumiで環境定義。

## 運用基盤
- ロギング: CloudWatch Logs / Stackdriver。
- モニタリング: Prometheus + Grafana or Datadog。
- アラート: PagerDuty / Opsgenie連携。

## フロントエンド（MVP）
- React + TypeScript、UIフレームワークにChakra UI。
- APIクライアント: React Query。
- 静的ホスティング: Vercel or CloudFront + S3。

## 開発支援ツール
- Lint: golangci-lint, ESLint。
- フォーマッタ: gofmt, Prettier。
- コード品質: SonarCloud検討。
- ドキュメント: OpenAPI + Storybook。

## 決定待ち項目
- ORマッパー選定、クラウドベンダー最終決定、監視ツールのコスト比較。
