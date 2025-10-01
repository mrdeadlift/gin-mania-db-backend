# CI/CD雛形構築

## 目的
バグ検出とデプロイを自動化し、リリースサイクルを短縮する。

## CIワークフロー案（GitHub Actions）
- トリガー: PR作成・更新、mainブランチ push。
- ジョブ構成:
  1. **lint**: `golangci-lint`, `eslint` を並列実行。
  2. **test-backend**: Go単体テスト、integrationテストはdocker-composeサービス起動。
  3. **test-frontend**: `yarn test --watch=false`。
  4. **build**: バックエンドコンテナイメージ、フロントエンドビルド成果物生成。
- 成果物: Docker image をGitHub Container Registryへpush、フロントはアーティファクト保存。

## CDパイプライン案
- mainブランチマージ時に自動デプロイ。
- ステージング: ECS Fargate / Cloud Run へデプロイ、Smokeテスト実行。
- 本番: ステージング承認後に手動トリガー。
- インフラ変更はTerraform Cloudのplan/applyをIntegrate。

## セキュリティ
- Dependabotで依存更新通知。
- SAST（GoSec, Trivy）を週次実行。
- シークレットはGitHub OIDC + AWS/GCP IAMロールで取得。

## ロールバック戦略
- コンテナイメージのバージョン管理（タグ付け）。
- マイグレーションは`migrate`のdownスクリプト整備。
- 監視アラート発火時に直前の安定版へ切り戻し手順書。

## 今後の強化候補
- パフォーマンステスト自動実行（k6）。
- リリースノート自動生成（Release Drafter）。
- カナリアリリース/Blue-Green Deployの検証。
