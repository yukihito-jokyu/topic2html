# FEAT-002 実行時設定

## 設定契約

`cmd`は次のServer限定設定を起動時に読み、妥当性を検証してCodex app-server adapterへ注入する。

| setting | 必須 | 規則 | Browserへの露出 |
| --- | --- | --- | --- |
| `TOPIC2HTML_CODEX_APP_SERVER_EXECUTABLE` | はい | 絶対パスの実行可能ファイル。利用者入力を連結せず、固定argv `app-server --stdio`だけで起動する。 | 禁止 |
| `TOPIC2HTML_CODEX_APP_SERVER_WORKDIR` | はい | 専用の空作業directory。app source、DB data、秘密設定を置かず、専用service accountだけが書込み可能。 | 禁止 |

設定値が欠落・形式不正・実行不能ならServerは起動を失敗させる。リクエスト単位のfallbackや、Browserからの接続先・引数指定は提供しない。

## 資格情報と実行隔離

Codex認証情報はapp-serverプロセスだけのサービス実行環境で管理し、`cmd`、HTTP、DB、Browser、ログへ値をコピーしない。app-serverには専用service account、最小filesystem権限、空のworkdirを使う。アプリのsource tree、PostgreSQL接続情報、管理session store、CSRF秘密値はworkdir・process environment・promptに渡さない。

アダプタはstdioで起動し、v2 JSON-RPCの最小wire、`read-only` sandbox、approval拒否、HTML出力選択、process cleanupを[Codex app-server adapter契約](codex-app-server-adapter.md)どおりに実施する。引数、URI、credential、thread/turn ID、モデル固有設定、外部エラー本文は安全な失敗分類へ変換する前に破棄し、永続化・監査ログ・HTTP応答へ出さない。

## 運用確認

本番相当環境では、秘密を隔離したsmoke testで起動・最小生成・安全な失敗変換を確認する。通常のunit、repository、HTTP、UIテストは決定的なapp-server test doubleを使い、実アカウント・認証情報をfixtureやCIログへ持ち込まない。
