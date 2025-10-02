# Gin Mania Backend MVP 設計

## 目的と前提
- `doc/mvp_scope.md` で定義した最小機能（認証、銘柄管理、検索、詳細表示、試飲記録、管理者承認）に対応するバックエンド設計をまとめる。
- 既存の Gin ベース API を拡張し、将来の機能追加に備えた疎結合なレイヤ構成を採用する。
- インフラはローカル開発を docker-compose、ステージング／本番はコンテナ実行基盤（ECS Fargate or Cloud Run）を想定する。
- 銘柄検索・詳細閲覧系の API は非会員でも利用できる公開エンドポイントとし、会員向け機能との境界を明示する。

## システム全体像
```
[React Client] --HTTPS--> [Gin API Gateway] --RPC--> [Domain Services]
                    |                         |---> PostgreSQL 15 (永続化)
                    |                         |---> Redis (検索キャッシュ/レート制限)
                    |                         '---> Auth0 JWKS (署名検証)
                    '--> OpenAPI schema / Admin Console
```
- フロントは React + React Query で REST API を利用。管理向け UI も同一フロントのロール別画面とする。
- Gin サーバは API ルーティング層とユースケースサービス層を分離し、永続化／外部サービスアクセスをインフラ層に隔離する。
- 画像は当面 URL 文字列で保持し、アップロード機能はスコープ外。

## アプリケーション構成
```
cmd/
  server/main.go         ... HTTP サーバ起動、DI
internal/
  config                 ... 環境変数・設定ロード
  http
    middleware           ... 認証・リクエストID・ロギング
    router               ... ルート定義（v1）
  auth                   ... Auth0 JWKS 取得 + JWT 検証、ロール管理
  ginbrand               ... 銘柄ユースケース（CRUD, 検索, CSVインポート）
  tasting                ... 試飲記録ユースケース
  moderation             ... レビュー承認ユースケース
  user                   ... ユーザロール解決
  repository             ... データアクセス (GORM)
  cache                  ... Redis アクセサ
  importer               ... CSV バリデーション/取込（管理者のみ）
  logging/metrics        ... 共通ロガー・Prometheus 計測
pkg/
  api                    ... OpenAPI スキーマ由来 DTO, エラーレスポンス整形
migrations/              ... golang-migrate 用 SQL スクリプト
```
- サービス層 (`ginbrand.Service` など) は複数リポジトリ・キャッシュをオーケストレーションしトランザクションを定義する。
- リポジトリ層は GORM を利用し、複雑な検索は SQL Builder (gorm.Expr) で実装。
- 依存注入は `fx` 等は採用せず、最小構成として wire も使用せずハンドメイド DI。

## データモデル（PostgreSQL）
| テーブル | 主キー | 主要カラム | 説明 |
| --- | --- | --- | --- |
| `users` | `id UUID` | `auth0_sub (unique)`, `email`, `display_name`, `role (enum: member, admin)`, `created_at` | Auth0 から得たサブIDとロールを保持。管理者判定に使用。 |
| `gins` | `id UUID` | `name`, `distillery`, `region`, `abv DECIMAL(4,1)`, `tasting_notes TEXT`, `image_url`, `status (enum: draft, published)`, `created_by`, `updated_at` | 銘柄の基本情報。公開状態を管理者が制御。 |
| `botanicals` | `id UUID` | `name (unique)` | ボタニカルマスタ。 |
| `gin_botanicals` | `gin_id`, `botanical_id` | `PRIMARY KEY(gin_id, botanical_id)` | 多対多関連。 |
| `flavor_tags` | `id UUID` | `code (unique)`, `label` | 味覚タグマスタ（例: citrus, spice）。 |
| `gin_flavor_tags` | `gin_id`, `flavor_tag_id` | `PRIMARY KEY(gin_id, flavor_tag_id)` | 味覚タグ付け。検索フィルタに利用。 |
| `recommended_serves` | `id UUID` | `gin_id (unique)`, `description TEXT` | 推奨サーブ文言。最小要件として 1 レコード。 |
| `tasting_logs` | `id UUID` | `user_id`, `gin_id`, `rating SMALLINT`, `memo TEXT`, `tasted_at DATE`, `status (enum: pending, published, rejected)`, `created_at`, `updated_at` | ユーザー試飲記録。承認ステータスを保持。 |
| `moderation_events` | `id UUID` | `tasting_log_id`, `reviewer_id`, `action (enum: approve,reject)`, `note`, `created_at` | 管理者の承認操作を記録し監査証跡とする。 |
| `csv_import_jobs` | `id UUID` | `uploaded_by`, `file_name`, `status (enum: queued, processing, completed, failed)`, `summary JSONB`, `created_at` | 管理者による CSV 取込の状態。MVP では同期処理だが履歴管理のみ保持。 |

- 検索要件に対応するため `gins.name`, `gins.region` に B-Tree、`gin_flavor_tags.flavor_tag_id` に複合インデックスを付与。
- Botanic/Flavor タグは将来のカスタム入力に備えマスタテーブル化。
- `status` 列で公開フローを表現し、API レイヤで権限制御する。

## API 設計（v1）
| メソッド | パス | ロール | 用途 | 備考 |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/healthz` | public | ヘルスチェック | 既存実装流用 |
| `GET` | `/api/v1/gins` | public | 検索・一覧 | クエリ: `q`, `region`, `botanical`, `flavor_tag`; Redis キャッシュを適用 |
| `GET` | `/api/v1/gins/:id` | public | 詳細取得 | 推奨サーブ・タグ含む |
| `POST` | `/api/v1/admin/gins` | admin | 銘柄登録 | CSV 取込も内部的にこのサービスを利用 |
| `PUT` | `/api/v1/admin/gins/:id` | admin | 銘柄更新 | 状態遷移（draft→published）含む |
| `DELETE` | `/api/v1/admin/gins/:id` | admin | 論理削除 | `status=archived`（将来拡張用）を検討 |
| `POST` | `/api/v1/admin/gins/import` | admin | CSV アップロード → 同期取込 | ファイルは multipart、最大500行までをバリデート |
| `GET` | `/api/v1/tastings` | member | 自身の試飲記録一覧 | ページング、`status=published` のみ公開タブで閲覧可 |
| `POST` | `/api/v1/tastings` | member | 試飲記録登録 | 作成時 `status=pending`、管理者承認待ち |
| `PATCH` | `/api/v1/tastings/:id` | member | 自身の記録更新 | 承認後は `memo` のみ編集可 |
| `GET` | `/api/v1/admin/reviews` | admin | 承認待ち一覧 | クエリ: `status=pending`, `user_id` |
| `PATCH` | `/api/v1/admin/reviews/:id` | admin | レビュー承認/却下 | `action` フィールドで制御し `moderation_events` へ記録 |
| `GET` | `/api/v1/meta/botanicals` | public | ボタニカル候補取得 | フロントのフィルタ表示用 |
| `GET` | `/api/v1/meta/flavor-tags` | public | 味覚タグ候補取得 | 同上 |

- OpenAPI (Swagger) を `pkg/api/openapi.yaml` に定義し、gin-swagger 等で配信する。
- 入出力 DTO は `pkg/api` に集約し、ドメインエンティティと分離。

## 認証・認可
- Auth0 で発行された JWT (RS256) を `Authorization` ヘッダで受け取り、JWKS を定期フェッチし署名検証。公開エンドポイントではトークンがなくても処理を継続するが、利用者がログイン済みの場合は同じミドルウェアで認証情報を解決する。
- ミドルウェアでクレーム `sub`, `email`, `https://ginmania.app/roles` を抽出し `request.Context` に格納。
- `users` テーブルになければ初回アクセス時にレコード作成。管理者ロールはシードデータで登録し、Auth0 側のロールも同期。
- ハンドラレベルで `member` / `admin` などのロールチェックを実施し、公開 API ではロールチェックをスキップする。

## 検索とキャッシュ
- 通常検索は Postgres で `ILIKE` + JOIN を行い、結果（ページング単位）を Redis に 5 分間キャッシュ。
- キャッシュキー: `gin:search:{hash(params)}`。更新／削除時は関連キーを失効させる。
- レート制限は IP + ユーザー ID に対して Redis トークンバケットを適用（30 req/min）。

## CSV インポートフロー
1. 管理者がテンプレート CSV（UTF-8, ヘッダ付き）をダウンロード。
2. `POST /admin/gins/import` でアップロード。
3. サーバ側でバリデーション → トランザクション内で upsert。
4. 成功・失敗件数を `csv_import_jobs.summary` に JSON で保存しレスポンス。
- 行数が閾値を超える場合は失敗させ、将来的に非同期処理（キュー）へ移行可能にする。

## ログ・モニタリング (MVP)
- ログ: zap ベース JSON ログ。構造化フィールド（`trace_id`, `user_id`, `endpoint`）。
- メトリクス: Prometheus exporter（HTTP レイテンシ、DB クエリ時間、承認件数）。
- アラート閾値（例）: 検索 API 95% タイル > 500ms, エラー率 > 2%.

## テスト戦略
- 単体: サービス層のビジネスロジックを testify + gomock でカバー。
- ハンドラ: httptest + モックサービス。
- 統合: docker-compose で Postgres/Redis を起動し、主要ユースケースを回す。
- CSV バリデーションはテーブル駆動テストで境界値を確認。

## 今後の拡張余地（MVP完了後）
- ソーシャルログイン追加時は `users` テーブルに `provider` 列を追加。
- レコメンド機能は別マイクロサービスとして切り出し可能なよう現在の API をステートレスに保つ。
- 多言語化のため `gin_translations` テーブルを追加し、レスポンスをロケール対応させる予定。
