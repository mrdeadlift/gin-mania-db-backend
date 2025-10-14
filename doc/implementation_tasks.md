# 実装タスクリスト

## フェーズ0: 基盤整備
- [x] `internal/config`で環境変数ロードとバリデーションを行う設定モジュールを実装する（DB/Redis/Auth0/ログ出力設定）。
- [x] zapベースの構造化ロギングラッパーを追加し、HTTPリクエストIDとコンテキスト情報を出力できるようにする。
- [x] `internal/http/router`を作成し、ルート登録と共通ミドルウェア（リカバリ、CORS、ロギング、リクエストID）を集中管理する。
- [x] `cmd/server/main.go`をDIエントリポイント化し、将来のサービス追加に備えて初期化処理をモジュール化する。

## フェーズ1: データモデルとマイグレーション
- [ ] `users`/`gins`/`botanicals`/`gin_botanicals`/`flavor_tags`/`gin_flavor_tags`/`recommended_serves`/`tasting_logs`/`moderation_events`/`csv_import_jobs`のスキーマをマイグレーションで定義する（enum相当はCHECK制約で実装）。
- [ ] 主要インデックス（名称全文検索、地域、タグJOIN、承認待ち一覧）を追加し、検索とモデレーションのパフォーマンス要件を満たす。
- [ ] 初期データ投入用のシードスクリプト（管理者ユーザ、代表的なボタニカル/味覚タグ）を準備する。

## フェーズ2: 認証・認可
- [ ] Auth0 JWKSフェッチ・キャッシュ機構と署名検証ロジックを`internal/auth`に実装する。
- [ ] JWTクレームからロールを読み出し`context`へ格納するミドルウェアを追加し、管理者・会員ロールを判定する。
- [ ] 初回アクセス時の`users`テーブル自動作成とロール同期処理をリポジトリ層で実装する。

## フェーズ3: ドメインサービス
- [ ] `internal/ginbrand`サービスを作成し、銘柄CRUD・検索・公開ステータス更新を提供する。
- [ ] `internal/tasting`サービスで試飲記録の登録・更新・ユーザ別取得を実装する（承認待ちはpending固定）。
- [ ] `internal/moderation`サービスでレビュー承認フローと監査ログ生成を行う。
- [ ] それぞれに対応するGORMリポジトリを`internal/repository`に実装し、トランザクション操作をサポートする。

## フェーズ4: APIハンドラ（公開エンドポイント）
- [ ] `/api/v1/healthz`を移設し、バージョン付きルーティングの土台を整える。
- [ ] `/api/v1/gins`検索で名称・地域・ボタニカル・味覚タグフィルタ、ページング、キャッシュ利用状況を返却する。
- [ ] `/api/v1/gins/:id`詳細APIで推奨サーブ・タグ・公開状態を返すDTOを定義する。
- [ ] `/api/v1/meta/botanicals`と`/api/v1/meta/flavor-tags`でフロントのフィルタ用マスタを提供する。

## フェーズ5: APIハンドラ（会員／管理者向け）
- [ ] `/api/v1/tastings`のGET/POST/PATCHを実装し、本人確認とステータス制約を適用する。
- [ ] `/api/v1/admin/gins`のCRUDと公開ステータス切り替え、CSVからの一括登録処理を公開する。
- [ ] `/api/v1/admin/reviews`系エンドポイントで承認待ち一覧、個別承認／却下、監査ログレスポンスを返す。
- [ ] DTO整形を`pkg/api`に集約し、OpenAPIスキーマと同期させる。

## フェーズ6: 検索・キャッシュ・レート制限
- [ ] Redisクライアントを`internal/cache`へ追加し、接続設定とヘルスチェックを用意する。
- [ ] 検索結果キャッシュをハッシュキー（`gin:search:{hash(params)}`）で実装し、更新時に失効する仕組みを組み込む。
- [ ] IP+ユーザID単位のトークンバケットによるレート制限ミドルウェアを追加する。

## フェーズ7: CSVインポート
- [ ] `/api/v1/admin/gins/import`でmultipartアップロード、UTF-8検証、行数上限チェックを実装する。
- [ ] 行単位バリデーション・エラー集計とDBトランザクション処理を`internal/importer`モジュールで行う。
- [ ] `csv_import_jobs.summary`に成功／失敗件数をJSON保存し、APIレスポンスに含める。

## フェーズ8: 観測・運用
- [ ] Prometheusメトリクス（HTTPレイテンシ、DBクエリ時間、承認件数）を`/metrics`で公開する。
- [ ] 検索APIの95パーセンタイル遅延、エラー率、レート制限判定をダッシュボード化できるようメトリクス名を整備する。
- [ ] 構造化ログにトレースID・ユーザID・エンドポイント情報を含め、JSON形式で出力する。

## フェーズ9: テストと品質保証
- [ ] ドメインサービスの単体テストをtable driven + gomockで実装し、成功・失敗パスを網羅する。
- [ ] ハンドラのHTTPテストを`httptest`で追加し、ロール別のアクセス制御を検証する。
- [ ] docker-compose上でPostgres/Redisを用いた統合テストスイートを`make test-integration`に組み込む。
- [ ] CSVバリデーションの境界ケース（最大行数・文字コード・重複）をテストデータで検証する。

## フェーズ10: 開発体験とドキュメント
- [ ] Makefileを整備し（`setup`/`run`/`test`/`docs`）、オンボーディングを容易にする。
- [ ] OpenAPI定義ファイルを`pkg/api/openapi.yaml`として管理し、Swagger UI/Redocの配信ルートを追加する。
- [ ] `doc/`配下の`dev_environment.md`や`system_design.md`を実装結果に合わせて更新する。
- [ ] READMEをMVP機能・起動手順・主要エンドポイント例で刷新する。

## トラッキングとマイルストーン
- [ ] Phase 0〜4をMVP APIリリースのブロッカーとし、完了時点でβ向け通しテストを開始する。
- [ ] Phase 5〜7をβ公開前の必須条件としてアラート設定とCSV運用フローを確定する。
- [ ] Phase 8〜10をβ運用開始までに順次完了し、残タスクはIssue化してバックログ管理する。
