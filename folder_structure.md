# 📁 クリーンアーキテクチャベースの Web アプリケーション構成

## フォルダ構成図

```
app/
├── api/
│   ├── middlewares/
│   │   ├── auth.go
│   │   ├── logging.go
│   │   └── tracing.go
│   └── router.go
├── apperrors/
│   ├── error.go
│   ├── handler.go
│   └── codes.go
├── common/
│   └── utils.go
├── controllers/
│   ├── base_controller.go
│   └── handlers/
├── models/
│   └── entities.go
├── repositories/
│   ├── base_repository.go
│   └── impls/
├── services/
│   ├── base_service.go
│   └── impls/
└── main.go
```

## 各フォルダの説明

### 📂 api/

HTTP ルーティングとミドルウェアを配置するのだ！

- router.go
  - ルーティング設定
  - ミドルウェアの適用
  - URL パスとハンドラの紐付け

#### 📂 middlewares/

- auth.go
  - 認証処理
  - トークンの検証
- logging.go
  - リクエスト/レスポンスのログ出力
- tracing.go
  - 分散トレーシングの ID 生成と管理

### 📂 apperrors/

エラーハンドリングの共通処理を配置するのだ！

- error.go
  - カスタムエラー型の定義
- handler.go
  - エラーレスポンスの生成ロジック
- codes.go
  - エラーコード定数

### 📂 common/

共通ユーティリティを配置するのだ！

- utils.go
  - 汎用的なヘルパー関数
  - 定数定義
  - 共通インターフェース

### 📂 controllers/

HTTP リクエストの制御を行うのだ！

- base_controller.go
  - 入力値検証
  - レスポンス整形
  - サービス呼び出し

### 📂 models/

データ構造を定義するのだ！

- entities.go
  - データモデルの構造体定義
  - バリデーションルール

### 📂 repositories/

データアクセスを担当するのだ！

- base_repository.go
  - DB 操作の実装
  - トランザクション制御
  - クエリ実行

### 📂 services/

ビジネスロジックを実装するのだ！

- base_service.go
  - ユースケース実装
  - トランザクション制御
  - リポジトリの利用

## ✨ 特徴

1. **レイヤードアーキテクチャ**

   - 各層の役割を明確に分離
   - 依存関係を一方向に制御

2. **インターフェース指向**

   - 抽象化による疎結合な設計
   - テスト容易性の向上

3. **エラーハンドリング**

   - 統一的なエラー処理
   - エラー情報の適切な伝播

4. **横断的関心事の分離**
   - ログ出力
   - 認証・認可
   - トレーシング
