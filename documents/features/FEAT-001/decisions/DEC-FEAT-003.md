# DEC-FEAT-003 — Backendを独立Go module・検証単位とする

状態: **承認済み（2026-08-15、L3）**

## 課題

利用者は`backend/`と`frontend/`の分離、およびBackendのClean Architecture準拠を指定した。Goの依存管理・検証単位を親リポジトリに残すか、`backend/`を独立Go moduleにするかは、`go.mod`/`go.sum`の配置、CIと開発者の検証コマンド、既存初期Go Serverの移行単位に影響する。DEC-ARCH-003は分離と依存方向を採用したが、この運用判断は未確定である。

## 選択肢

### A. `backend/`を独立Go module・独立検証単位にする（推奨）

`backend/`にGo moduleの正本を置き、Go依存の取得、整形、静的検査、unit/integration test、buildをBackendの作業ディレクトリから実行する。親リポジトリはFrontendとBackendをまとめて起動するTask/CI入口だけを持てる。

**影響:** Backendの依存グラフと検証範囲がFrontend・Planning成果物から明確に分離される。Backend担当者は`backend/`で完結して検証できる。既存の親直下Go moduleと初期Serverは移行対象となり、CI・README・開発手順を更新する必要がある。複数moduleを横断するGo testは行わず、共有するのはHTTP wire契約とE2Eだけとする。

### B. 親リポジトリ直下のGo moduleを維持する

親直下にGo moduleを置き、`backend/`をソース上の配置境界だけにする。Goの検証は親直下から実行する。

**影響:** 初期ServerのGo module移動を避けられるが、Backendの依存・検証単位がFrontend、documents、リポジトリ運用と混在しやすい。BackendとFrontendの分離をCI・オンボーディングで明確にする追加規約が必要になる。

## 推奨と承認

**選択肢Aを推奨し、利用者が承認した。** 利用者指定のBackend/Frontend分離と、Backendを独立した信頼境界・検証単位として扱う方針に最も整合するためである。HTTP契約とE2Eを横断確認に残すことで、独立module化による分断は防げる。

## 承認後の影響

- `backend/`をGo依存とBackend検証の唯一の正本とし、rootの初期Go module・Serverを移行または置換する。Task/CIはBackend検証を`backend/`で、Frontend検証を`frontend/`で実行する。
- BackendのGo依存は`backend/go.mod`と`backend/go.sum`で固定する。root Go moduleを残してBackend依存を管理してはならない。
- 複数moduleを横断するGo testは行わない。Frontend/Backend間の互換性は承認済みHTTP wire契約とE2Eで検証する。
- `handler`/`usecase`/`repository`/`domain`の依存方向、Ginの`handler`限定、HTTP/DB/OAuth/CSRF/秘密情報の承認済み契約は変更しない。

## 実装開始条件

本Decisionに関するL3承認は完了した。Task Breakdownは、Design ReviewとImplementation Readiness Reviewが`pass`になった後に開始できる。
