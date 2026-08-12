# 論理Path・Globの単一Ownerと共有変更引き渡し

- 対象作業: `L1-M1-S3`
- 状態: 設計済み
- 確認日: 2026-08-12

## 用語と目的

論理Pathは、実装開始前に「何の責務を置く場所か」を表す予定のRepository上の場所である。Globは`**`のような記号で複数Fileをまとめて示すPathパターンである。単一Ownerは、そのPath内の変更を最終判断して統合を依頼できる一つのTaskまたは担当である。単一Ownerが必要なのは、二つのTaskが同じ共有Fileを別々に変えると、依存、履歴、責任が矛盾するためである。

この表は、現在存在しない実装Directoryの所有を予告するものである。具体File、Package、Schema、API、依存追加を許可するものではない。実装開始時は各後続Taskがここで定めた範囲を受け取り、実際のFile名と実装内容はそのTaskの承認範囲で決める。

## Owner表

| 論理Path・Glob | 単一Owner | Ownerが判断すること | 他Taskが直接変更してはいけない理由 |
| --- | --- | --- | --- |
| `docs/architecture/modules/**` | `L1-M1-S3` | Module責務、境界、禁止事項。 | 後続実装Taskが設計文書を都合よく変えると、依存方向の根拠が失われるため。 |
| `docs/architecture/dependencies/**` | `L1-M1-S3` | 許可・禁止依存、循環禁止、依存変更の判断入口。 | 個別機能が例外を直接追加すると、全体の到達不能条件を確認できなくなるため。 |
| `docs/architecture/ownership/**` | `L1-M1-S3` | 論理Path/Glob、単一Owner、共有変更の引き渡し規則。 | Ownerが重複すると、共有変更の承認者を決められないため。 |
| `frontend/**` | `L1-M1-S3`。L1完了時に`L7-M1-S1`へ移管するまでの、唯一の変更要求受付Owner。 | 後続Featureごとの一意な配下Pathを割り当てること、共有変更かを判定すること。 | まだFeature別Pathを分けていないため、どのFeatureも直接変更しない。複数Taskが同じFrontend範囲を編集すると、Browser公開情報の責任が混ざるため。 |
| `backend/http/**` | `L1-M1-S3`。L1完了時に`L7-M1-S1`へ移管するまでの、唯一の変更要求受付Owner。 | Route/API Featureごとの一意な配下Pathを割り当てること、共有変更かを判定すること。 | まだFeature別Pathを分けていないため、どのFeatureも直接変更しない。共通HTTP規則や他FeatureのRouteを一Taskが横断変更しないため。 |
| `backend/features/<feature>/**` | 対応するL2〜L6のFeature Task。ただし`<feature>`の具体名とPathは、割当て前は`L1-M1-S3`、L1完了後は`L7-M1-S1`だけが決める。 | 業務意味、状態遷移、用途別Port、Feature-local Fixture。 | 他Featureが状態の意味を直接変更しないため。未割当の`<feature>` Pathは直接変更禁止である。 |
| `contracts/**` | L1-M4の契約物理化Task。 | 言語非依存Field定義、Canonical Fixture、Contract Test、Code生成採否の実装。 | TypeScript型またはGo型の片方だけを正本にしてはならないため。 |
| `backend/adapters/codex/**` | `L3`の後続Task。 | Codex app-serverへの生成用Port実装。 | Frontend/Rendererへ接続情報を渡さず、生成処理の具体方式をL3に閉じるため。 |
| `backend/adapters/postgresql/**` | `L1-M1-S3`。L1完了時に`L7-M1-S1`へ移管するまでの、唯一の変更要求受付Owner。 | Featureごとの一意なAdapter配下Pathを割り当てること、共有接続・Pool変更をL7へ回すこと。 | まだFeature別Adapter Pathを分けていないため、どのFeatureも直接変更しない。Feature TaskがDriverや資格情報を直接扱わず、保存技術の変更を局所化するため。 |
| `backend/migrations/**` | `L1-M4-S2`。 | Runner実装、SQL実行、Lock/Timeout順序、失敗停止、接続解放の試験。 | 通常Backend起動やFeatureがSchema変更を直接起こさないため。Migration統治上の意味は`L1-M2-S3`が別に所有する。 |
| `backend/resolver/**`、`renderer/**` | `L6`の唯一のOwner。 | Preview/Publicの対象解決、最小表示契約、Renderer隔離方式。 | 保存・管理画面Taskを含む他Taskは直接変更しない。公開表示の安全境界を一つのOwnerだけが判断できるようにするため。 |
| `test/**` | 対象機能のTest Owner。ただし横断Mount/統合Fixture/E2Eは`L7`。 | Feature-local Test、Test用Fixture。 | 共有Test基盤や全体E2Eを個別Featureが変更すると、統合責任が重複するため。 |
| Root Manifest、Lockfile、Composition root、全体Router、DI/Registry、統合Fixture、E2E | `L7`共有変更lane。 | 共有依存、全体配線、共通基盤、最終統合。 | Feature worktreeから直接変更すると、他FeatureのBuild・起動・Testを壊し得るため。 |

## L7への共有変更引き渡し規則

L7共有変更laneは、複数Featureが同じ基盤を使うときに、変更を一つずつ受け取り全体へ統合する後続の責任範囲である。これはFeatureの業務意味をL7が決める仕組みではない。Feature Ownerが意味を決め、L7は共有Fileと全体配線を一意に管理する。

1. Feature Ownerは、自分のOwner Path外の変更が必要になった時点で、直接編集せず変更要求を作る。変更要求には、必要なPath/Glob、変更理由、依存Matrix上の利用方向、影響を受けるFeature、必要なTest、要求元Ownerを記録する。理由を残すのは、L7が意味変更を推測して採否しないためである。
2. Root Manifest・Lockfile・Composition root・全体Router・DI/Registry・統合Fixture・E2Eは、L7が受領して変更する。依存追加や共有配線を先に一つのOwnerが統合し、必要なFeature worktreeはその統合結果へ追随する。これにより、同じ共有Fileを複数worktreeで競合編集しない。
3. 契約の追加は、提供Ownerと利用者Ownerが互換性を確認してからL7へ渡す。意味変更、削除、状態Ownerの変更は、該当FeatureのGate再確認を必要とする。L7は単独で業務意味を変更しない。
4. Migration番号、共有Migration File、Root依存はL7が統合するが、Migrationの論理Schema・互換性・復旧方針は`L1-M2-S3`、Runner実装・Timeout/Lock/解放試験は`L1-M4-S2`の判断を必要とする。これは一つの共有Fileに複数の責務を混ぜないためである。
5. 引き渡し後も、元のFeature Ownerは自分の業務要件とFeature-local Testを所有する。L7は受領した共有変更の統合状態と全体のBuild・起動・横断Testを所有する。問題がFeature意味に戻る場合は、L7が直接修正せず元Ownerへ差し戻す。

## L1完了・G0と後続への渡し方

`docs/task-connections.md`（Task Issue同士の直接の前後関係と、引き渡す成果物を記録する接続台帳）は、現在まだ作成されていない。そのため、このTaskから直接続くIssue番号を確定できず、Task Mapだけから番号を推測してはならない。

今回の三文書は、まず`L1-M1`の完了確認と、L1-M1とL1-M2の成果物を合わせて確認する`G0`の整合確認へ渡す入力である。L1が完了した時点で、Path/Glob Ownerと共有変更の引き渡し記録は`L7-M1-S1`へ渡す。これはTask Mapが定める引き渡し先であり、現時点で未作成の接続台帳に代えて、直接後続Issueを断定するものではない。

## 境界変更の確認

Module、許可依存、Path/Glob Ownerの変更は相互に影響する。そのため変更を承認する前に、(1) このOwner表、(2)[Module・Directory境界](../modules/module-directory-boundaries.md)、(3)[Module依存Matrix](../dependencies/dependency-matrix.md)、(4) Task Mapの直接依存とGateを照合する。差異があれば、実装や共有変更を進めず、設計Ownerと後続利用者に判断を戻す。
