# トピック解説HTML生成・管理Webアプリ Task Map議論記録

- 状態: 全階層承認済み・横断最終監査完了
- 最終更新: 2026-08-11
- 原典: [topic-to-html-requirements.md](./topic-to-html-requirements.md)
- 要件議論: [topic-to-html-requirements-discussion-map.md](./topic-to-html-requirements-discussion-map.md)
- 確定Task Map: [task-map.md](./task-map.md)

## TM-D-001 初期リリース直下の大分類案

- 親タスク: トピック解説HTML生成・管理Webアプリの初期リリースを実現する
- 今回決める階層: 大分類
- 状態: 承認済み
- 既存の承認済みTask: なし
- 方針: 一段下の大分類だけを候補とし、中分類・小分類・実装Issueは後続の議論で分割する。

### T-100 横断設計・共有契約と最小実装基盤を確立する

分割した理由：  
実装構成が未定であり、各機能が共有する信頼境界、依存方向、ドメイン語彙、Schema、Interface、所有境界と、起動・テスト可能な最小骨格が必要であるため。DD-001〜DD-007の機能固有設計は抱えず、横断判断と共有契約だけを所有する。

完了条件：

- 技術構成、アプリ境界、信頼境界、依存方向が記録されている。
- HTML全体、版、生成試行、公開状態、タグ、配置先、利用者の共有契約が定義されている。
- 初期Schema・Migration方針、共通Interface、所有境界が確定している。
- 起動・テスト可能な最小骨格と初期共有Fixtureがある。

原典対応: 要件定義全体、BR-009、BR-010、BR-017、BR-020、BR-021、CON-001、DD-001〜DD-007の横断部分。

### T-200 単一管理者のGoogle認証と管理認可を実現する

分割した理由：  
事前登録された1つのGoogleメールアドレスだけへ管理権限を与える責務が、各管理機能や匿名閲覧とは独立した認証・認可境界を持つため。Google認証、管理者判定、セッション、管理API・UIの認可を所有し、表示面の隔離はT-600へ渡す。

完了条件：

- 対象メールアドレスで管理者としてログインできる。
- 対象外アカウント、未認証者、閲覧者へ管理権限を付与しない。
- 管理操作の認可とセッション管理が機能する。
- T-600へ渡す資格情報・認可境界の契約が定義されている。

原典対応: FR-007、BR-001、BR-011、BR-019、AC-014、AC-023、ASM-001、ASM-002、DD-002。

### T-300 解説HTMLを生成・形式検証・再試行し生成ログを記録する

分割した理由：  
自然言語入力からCodex app-serverでHTMLを生成し、形式検証、失敗記録、要約ログ、自動・手動再試行までを一つの生成ワークフローとして完結させるため。成功時は検証済み結果、失敗時は内部ログを含まない試行要約を返す契約を所有する。

完了条件：

- トピックを含む自然言語プロンプトからHTMLを生成できる。
- 生成結果がHTML形式であることを機械検証できる。
- 生成失敗と検証失敗へ最大3回の自動再試行を適用できる。
- 3回失敗後の最終エラーと管理者による手動再試行が機能する。
- 成否と失敗理由の要約ログを保持・表示し、詳細内部ログは表示しない。
- `GenerationRequest`、修正元版、修正指示、検証済み生成結果、失敗試行要約の受け渡し契約がある。

原典対応: FR-001、FR-004〜FR-006、BR-003〜BR-008、NFR-001、NFR-004、CON-001、AC-001、AC-003〜AC-008、AC-016、AC-025〜AC-027、DD-001、DD-004。

### T-400 検証済みHTMLを確認・修正し版履歴を管理する

分割した理由：  
検証合格HTMLの確認、修正指示による生成の呼び出し、初回・修正後の成功版の不変保存、履歴閲覧を一つのHTMLライフサイクル責務として独立させるため。生成実行はT-300、公開版ポインターと公開状態遷移はT-600が所有する。

完了条件：

- 管理者が検証済みHTMLと版履歴を確認できる。
- 自然言語の修正指示からT-300の生成処理を呼び出せる。
- 初回生成と修正生成の成功結果を同じ版Repositoryへ不変版として保存できる。
- タグと配置場所を版履歴へ混入させない。
- 版の保存・列挙契約をT-600へ提供する。

原典対応: FR-002、FR-009、FR-010の履歴部分、BR-015〜BR-017、BR-020、AC-002、AC-017、AC-018、DD-006の版保存・履歴部分。

### T-500 HTML全体のタグとサイト内配置を管理する

分割した理由：  
タグと配置先はHTML全体へ紐付き、版とは異なるライフサイクルを持つため。検証済みHTMLだけを対象に、タグ操作と任意名の配置先を一貫して管理する独立成果とする。

完了条件：

- タグの作成、名前変更、付与、解除ができる。
- 検証失敗記録へタグを付与できない。
- 検証済みHTMLへ任意名の配置先を設定・変更できる。
- 公開版を切り替えてもタグと配置場所が維持される。

原典対応: FR-003、BR-009、BR-010、BR-014、BR-020、AC-009〜AC-013、AC-028、DD-003。

### T-600 公開版状態と隔離された生成HTML閲覧を実現する

分割した理由：  
公開承認、公開版切り替え、非公開化、匿名閲覧と、管理者プレビューを含む生成HTMLの隔離表示が、一つの公開・信頼境界を形成するため。T-400の不変版を参照し、公開版ポインターと表示面の隔離を所有する。

完了条件：

- 検証済み版を公開承認し、未承認版は公開しない。
- 新版の承認までは従来の公開版を維持できる。
- 過去版への切り替えを公開承認として扱い、公開を取り消して非公開化できる。
- 未ログイン閲覧者は公開版だけを閲覧でき、管理操作は実行できない。
- 管理者プレビューと匿名閲覧の双方で、生成HTMLから認証情報・管理データへアクセスできない。
- 生成HTMLの外部通信は一律に遮断しない。

原典対応: FR-008、FR-010の公開状態部分、FR-011、BR-002、BR-011〜BR-013、BR-016〜BR-018、BR-021、NFR-002、NFR-003、AC-015、AC-019〜AC-024、AC-029、DD-005、DD-006の公開切り替え部分、DD-007。

### T-700 初期リリースを統合し受入テストを自動化する

分割した理由：  
個別機能のテストだけでは、認証、生成、履歴、タグ・配置、公開、隔離を横断する利用体験を証明できないため。単なる確認Gateではなく、統合済み起動構成、配線、E2E、共有Fixture、再現可能な受入手順を成果物として所有する。

完了条件：

- 管理者の入力から生成、確認、修正、管理、公開までが一連で動作する。
- 匿名閲覧、権限制御、隔離を含むAC-001〜AC-029の自動受入テストがある。
- composition root、全体Router、DI・Registry、統合Fixtureが完成している。
- 起動・設定・検証手順を再現でき、重大な未統合や不具合がない。

原典対応: 目的・期待成果、成功指標、スコープ、AC-001〜AC-029。

## 依存DAG案

### 着手条件

- T-100は依存なし。
- T-200〜T-600の機能詳細設計、契約テスト、Mock・Fixture作成は、G0で該当する共有契約が承認されればT-100完了前でも着手できる。
- T-200〜T-600の本番実装は、最小基盤のMergeと、各機能のG1通過後に着手する。
- T-700の統合骨格とE2E準備はT-100のMerge後に着手できる。

### 完了・Merge依存

```text
T-100 ─┬─> T-200 ───────────────┐
       ├─> T-300 ─> T-400 ─────┼─> T-600 ─> T-700
       ├─> T-500 ───────────────┘
       └────────────────────────────> T-700
```

- T-200、T-300、T-500は、共有契約を変えない範囲で相互に独立してMergeできる。
- T-400の統合MergeはT-300に依存する。実装自体は確定InterfaceとMockで並行できる。
- T-600の統合MergeはT-200、T-400、T-500に依存する。実装自体はStubで並行できる。
- T-700の完了はT-200〜T-600の統合Mergeに依存する。

### 後続利用者

- T-100の共有契約と基盤をT-200〜T-700が利用する。
- T-300の検証済み生成結果をT-400が保存する。
- T-400の不変版と版一覧をT-600が参照する。
- T-200の認可境界とT-500のタグ・配置情報をT-600が利用する。
- T-200〜T-600の成果をT-700が統合・検証する。

## Gate／Milestone案

- G0 横断設計Gate: 技術構成、信頼境界、依存方向、ドメインモデル、Schema・Migration方針、共有Interface、所有境界を承認する。不可逆な本番Schema・Migration・共通基盤実装の開始条件とする。
- G1 機能実装準備Gate: 認証、生成、版履歴、タグ・配置、公開・隔離ごとに、該当するDD、状態遷移、受け渡し契約、テスト観点を確認する。全機能一括Gateにはせず、通過した機能から本番実装できる。
- G2 初期リリース受入Milestone: 統合成果物に対してAC-001〜AC-029が合格し、重大な未統合がないことを確認する。

G0〜G2は独立成果物を持つTaskではなく、Task Map上のGate／Milestoneとして扱う。

## worktree実行規則案

- 共通起点は、G0通過後にT-100が作成・Mergeした、未Commit変更を含まない最小基盤Commitとする。
- T-200〜T-600は、各機能Directoryと機能固有テスト・Fixtureを所有する。
- 初期Schema・Migration、共有Interface、初期Manifest・Lockfile、コード生成規則、初期共有FixtureはT-100の単一Ownerとする。
- T-100完了後の共有Schema、Migration、Interface、ルートManifest、Lockfile、生成物の変更はT-700の共有変更laneで逐次取り込み、Migration番号・順序を中央管理する。
- composition root、全体Router、DI・Registry、統合Fixture、E2EはT-700の単一Ownerとし、機能worktreeから直接変更しない。
- 機能worktreeは共有変更laneのCommitへrebaseしてからMergeする。共有変更を最終統合まで溜めない。
- 統合Merge順は、T-100 → T-200・T-300・T-500（分離できる範囲で任意）→ T-400 → T-600 → T-700とする。
- 技術スタックとDirectory構成が未確定のため、具体的なFile所有表はT-100の後続分割で確定する。

## 対象外

- 学習効果の測定。
- HTML生成時間の上限・目標設定。
- 特定ブラウザ・端末の表示保証。
- 測定可能なアクセシビリティ要求。
- アプリ独自の対象外トピック分類。
- 詳細内部ログの利用者向け表示。
- 原典にない本番ホスティング・運用監視。

## 既存構造の修正候補

現在Task Mapおよび承認済みTaskが存在しないため、修正候補はない。

## 監査記録

- 分割担当: 大分類7件を提案。横断基盤の先行、G0〜G2、機能worktreeの並行化を提示した。
- 妥当性確認担当: 7件は境界修正を条件に維持可能と判定した。
- 統合時の修正:
  - 横断基盤を共有契約と最小骨格へ限定した。
  - 認証情報の管理をT-200、表示面の隔離をT-600へ一意化した。
  - 版保存をT-400、公開版ポインターと公開状態遷移をT-600へ一意化した。
  - 修正操作のオーケストレーションをT-400、生成・検証・再試行をT-300へ一意化した。
  - G0とG1の重複を除き、G1を機能別Gateにした。
  - 全機能のT-100完了待ちを撤廃し、共有契約確定後の設計・Mock作業を並行可能にした。
  - 共有ManifestとMigration順序を単一Owner対象へ追加した。

## 承認状況

- TM-D-001: 承認済み
- 承認日: 2026-08-11
- 承認内容: 大分類7件、責務境界、依存DAG、G0・機能別G1・G2、worktree実行規則
- テンプレート適用: 指定されたTask Map形式に合わせ、候補ID `T-100`〜`T-700`を確定ID `L1`〜`L7`へ対応付けた。
- 反映先: [task-map.md](./task-map.md)

| 議論時ID | 確定ID |
| --- | --- |
| T-100 | L1 |
| T-200 | L2 |
| T-300 | L3 |
| T-400 | L4 |
| T-500 | L5 |
| T-600 | L6 |
| T-700 | L7 |

---

# TM-D-002: L1の中分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L1 横断設計・共有契約と最小実装基盤を確立する`
- 提案日: 2026-08-11
- 承認状況: 2026-08-11にユーザー承認済み
- Task Map反映: L1-M1〜L1-M4を「承認済み」として反映済み

## 中分類案

### L1-M1 横断アーキテクチャ・技術構成・信頼境界を設計する

#### 分割した理由

全機能が従う技術的な外枠を、機能固有設計と混在させず確定するため。

#### 主成果物・完了条件

- 技術選定ADR、実行トポロジー、信頼境界図、依存方向、Module・Directory境界を作成する。
- 管理者認証情報、Google認証設定、Codex app-server設定、公開可能設定を分類する。
- G0で技術構成、信頼境界、依存方向、所有境界を判定できる状態にする。

#### 境界

- 生成HTML隔離を判断できる粒度まで横断境界を扱う。
- Codex連携、Google認証、生成HTML隔離などDD-001〜DD-007の機能固有方式はL2〜L6へ残す。
- 原典外の本番Hosting・運用監視は対象外とする。

### L1-M2 共有ドメインモデル・状態語彙・機能間契約を設計する

#### 分割した理由

技術構成とは別に、L2〜L6の受け渡しと責務境界を技術非依存で固定するため。

#### 主成果物・完了条件

- HTML全体、版、生成試行、タグ、配置、公開状態、利用者の論理モデル、識別子規則、状態Ownerを定義する。
- `GenerationRequest`、修正元版、検証済み結果、失敗試行要約、不変版参照、公開版参照を含む機能間Interface仕様とMock例を定義する。
- 論理SchemaとMigration方針を定義し、L2〜L6が独立して詳細設計できる状態にする。

#### 境界

- 公開状態の具体的遷移はL6、再試行状態遷移はL3、認証状態はL2へ残す。
- 物理Schema、Repository実装、各機能UIは対象外とする。

### L1-M3 起動可能な最小Webアプリ・開発テスト基盤を構築する

#### 分割した理由

共有契約やMigrationを載せる前に、全機能が利用できる再現可能な実装土台を単一Ownerで作るため。

#### 主成果物・完了条件

- Project骨格、初期Manifest・Lockfile、Build・起動・テストコマンド、秘密値を含まない設定雛形を作成する。
- 最小Bootstrap、Smoke確認、コード生成規則、初期共有Fixture基盤を整備する。
- クリーン環境でBuild、起動、基礎テスト、利用する場合のコード再生成を再現できる状態にする。

#### 境界

- Unit・Contract testは実行基盤だけを所有し、機能固有テストは所有しない。
- 機能Router、最終composition root、DI・Registry、横断E2EはL7へ残す。

### L1-M4 共有契約とSchema・Migration実行基盤を実装する

#### 分割した理由

G0で承認した設計を共有コードとMigration基盤へ変換し、機能固有Schemaへの先回りを防ぎながらL2〜L6の共通起点を完成させるため。

#### 主成果物・完了条件

- 共有識別子、共有型、Port・Interfaceを実装する。
- Migration実行・検証機構、必要な場合は空または基盤用の初期Migration、Schema Fixtureを実装する。
- コンパイル、シリアライズ、契約互換性、Migration実行をテストし、L2〜L6から参照可能にする。

#### 境界

- L2の認証、L3の生成試行、L4の版、L5のタグ・配置、L6の公開状態に関する業務Table・Repository・状態遷移は実装しない。
- 機能固有Migrationの実ファイル更新は、L1完了後にL7の共有変更laneで直列化する。

## 依存DAG・Gate

```text
L1-M1 ─┐
       ├─→ G0 ─→ L1-M3 ─→ L1-M4 ─→ L1完了
L1-M2 ─┘
```

- L1-M1とL1-M2は別Path・別Ownerとして並行し、相互レビューとG0で最終整合する。
- G0はModule依存方向と機能間Interfaceの整合も確認し、独立Taskにはしない。
- G0前でも技術Spike、雛形案、Mock、契約テスト案は作成できるが、共通起点への本実装MergeはG0後とする。
- L1-M3はG0通過後にMergeし、L1-M4はG0通過とL1-M3 Mergeに依存する。
- L2〜L6の詳細設計・Mock作成はG0後に並行できる。本番実装はL1完了Commitと各機能G1に依存する。

## worktree所有境界・Merge順

| 対象 | 単一Owner |
| --- | --- |
| ADR、実行トポロジー、信頼境界、Module・Directory設計文書 | L1-M1 |
| ドメイン語彙、論理Schema、Interface仕様、Mock例 | L1-M2 |
| Root Manifest・Lockfile、Build・Test設定、Bootstrap、設定雛形、Fixture基盤、コード生成規則 | L1-M3 |
| Shared Source、共有型・Port、Migration実行基盤、基盤用Migration、Schema Fixture、契約互換テスト | L1-M4 |

- L1内Merge順は `L1-M1・L1-M2 → G0 → L1-M3 → L1-M4` とする。
- L1-M4で依存追加が必要な場合、Manifest・Lockfileを直接変更せず、L1-M3 Ownerの共有変更Commitを先にMergeしてrebaseする。
- L2〜L6の共通起点は、L1-M4までMergeされ未Commit変更がないL1完了Commitとする。
- L1完了時に具体的なPath・Glob所有表と引き渡し記録を作成し、共有Schema、Migration、Interface、Root Manifest、Lockfile、生成物のOwnerをL7共有変更laneへ移管する。

## 監査結果

- 分割担当: 4件は独立成果物を持ち、中分類として同程度の粒度と判定した。
- 妥当性確認担当: 4件を維持可能と判定した。
- 統合時の修正:
  - L1-M4の名称から「初期Schema」を外し、機能固有業務Tableを先行実装しないことを明確化した。
  - L1-M3のテスト責務を実行基盤に限定し、機能固有テストとの重複を除いた。
  - Manifest・Lockfileの単一Ownerと依存追加手順を明記した。
  - L1完了Commitを後続worktreeの共通起点とし、L7へのOwner移管条件を追加した。
- 既存構造への影響: L1の親子構造とL2〜L7の責務は変更不要。承認時にL1のSchema表現、共通起点、Owner移管規則だけをTask Mapへ同期する。

## 承認結果

- L1-M1〜L1-M4の計画状態を「議論中」から「承認済み」へ変更した。
- `L1`の次の議論をL1-M1の小分類分割へ進めた。
- L1のSchema表現、worktree共通起点、L7共有変更laneへの引き渡し条件を更新した。

---

# TM-D-003: L1-M1の小分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L1-M1 横断アーキテクチャ・技術構成・信頼境界を設計する`
- 今回決める階層: L1-M1直下の小分類のみ
- 提案日: 2026-08-11
- 承認状況: 2026-08-11にユーザー承認済み
- Task Map反映: L1-M1-S1〜S4を「承認済み」として反映済み

## 小分類案

### L1-M1-S1 論理実行トポロジー・横断信頼境界・主要データフローを設計する

#### 分割した理由

技術を選ぶ前に、管理者・匿名閲覧者・Webアプリ・生成HTML・Google・Codex app-server・外部サービス間で守るべき境界を、技術非依存の評価制約として固定するため。

#### 主成果物・完了条件

- 論理実行トポロジー図、横断信頼境界図、主要データフロー、生成HTMLの到達可能・到達禁止領域表を作成する。
- 管理者ブラウザ、匿名閲覧者、Webアプリ、永続化先、Google、Codex app-server、生成HTML実行領域、外部サービスを論理主体として識別する。
- 生成HTMLから認証情報・管理データへ到達させず、外部通信は一律禁止しない制約を表現する。
- L2・L3・L6が機能固有方式を設計する際の入力境界と移管先を明記する。

#### 境界・依存・所有

- Process、Container、Origin、Hosting製品、iframe・CSP等の隔離方式、OAuth方式、Codex連携方式、本番Hosting・監視は対象外とする。
- 直接依存はなく、G0前に完了する。S2の入力となる。
- 論理Pathは `docs/architecture/topology/**`、`docs/architecture/trust-boundaries/**`、`docs/architecture/data-flows/**` とする。

### L1-M1-S2 横断技術構成を選定しADRを作成する

#### 分割した理由

L1-M3とL1-M4が実装基盤を構築できるよう、S1の境界とCodex app-server固定制約を満たす横断技術スタックを、比較理由と見直し条件を含む決定記録として固定するため。

#### 主成果物・完了条件

- Web実行方式、Frontend・Backend、永続化、Migration、Build・Test基盤のADRを作成する。
- 各ADRに対象課題、評価基準、候補、採否理由、制約、Version固定方針、見直し条件を記録する。
- S1のどの信頼境界を満たすかを対応付け、全境界を選定技術で実現可能であることを逆照合する。
- 未決定事項をL2〜L6の機能固有DDへ明示的に移管する。

#### 境界・依存・所有

- CodexのAPI・Process起動・Prompt・失敗処理、Google認証Flow、HTML形式検証、版保存、隔離方式、機能UI、実装コード、依存Package追加は対象外とする。
- 直接依存はS1。S1 Merge後のclean commitから開始し、G0前に完了する。
- 論理Pathは `docs/architecture/decisions/**`、`docs/architecture/technology-stack/**` とする。

### L1-M1-S3 Module・Directory境界と依存方向を設計する

#### 分割した理由

選定技術をL1-M3以降が競合なく実装できるよう、機能責務を配置する構造、許可・禁止する依存方向、共有Fileの単一OwnerをProject骨格の実装と分離して固定するため。

#### 主成果物・完了条件

- Module・Package・Directory構成図と許可・禁止依存Matrixを作成する。
- Port Owner、Adapter・Stub・Mockの配置規則、L1-M2〜L7の論理Path・Glob所有表を定義する。
- Root Manifest、Lockfile、共有Interface、Migration、Registry、DI、Fixture、生成物のOwnerと、L1完了後のL7への引き渡し規則を一意にする。
- L1-M2の各契約Owner・提供側・利用側を依存Matrixへ配置できる状態にする。

#### 境界・依存・所有

- Interfaceの項目・型・状態語彙、Directory・Fileの実作成、Manifest・Lockfile更新、最終composition root・Router・DI・Registry実装は対象外とする。
- 直接依存はS2。S2後にS4と並行作業し、共有Indexと所有表の単一OwnerとしてS4より先にMergeする。
- 論理Pathは `docs/architecture/modules/**`、`docs/architecture/dependencies/**`、`docs/architecture/ownership/**` とする。

### L1-M1-S4 設定・秘密情報の分類と露出境界を設計する

#### 分割した理由

管理者認証情報、Google認証設定、Codex app-server設定、永続化資格情報、Session秘密、公開可能設定では露出可能範囲が異なるため、実装前に分類台帳と受け渡し制約を固定するため。

#### 主成果物・完了条件

- 設定・秘密情報台帳を作り、機密区分、供給元、利用主体、許可された露出面、記録可否を定義する。
- 管理者認証情報、Google Client設定・秘密値、Codex app-server設定、永続化資格情報、Session秘密、公開可能設定、Test用代替値を分類する。
- Server、管理画面、公開画面、生成HTML、Test Fixtureへの受け渡し可否を明示する。
- 文書・Log・Fixtureへ秘密値を残さず、公開Client設定と秘密値を区別する。

#### 境界・依存・所有

- 秘密値の発行・保存・注入・Rotation、OAuth設定手順、Codex接続実装、生成HTML隔離実装は対象外とする。生成HTMLの外部通信を一律禁止する規則は追加しない。
- 直接依存はS2とし、S1は推移依存とする。S2後にS3と並行作業する。
- 論理Pathは `docs/architecture/configuration/**`、`docs/architecture/secrets/**` とする。

## 依存DAG・Gate

```text
L1-M1-S1 → L1-M1-S2 → L1-M1-S3 ─┐
                         └→ L1-M1-S4 ┴→ L1-M1完了

L1-M1完了 + L1-M2完了 → G0
```

- S1は技術非依存の制約モデル、S2は制約を満たす技術、S3はコード配置・依存規則、S4は設定情報の露出規則を所有する。
- S2がS1の前提を満たせない場合は、S1を再検討してからS2をMergeし、暗黙の循環を残さない。
- S3とS4はS2後の同一clean commitから別worktreeで並行作業する。
- G0はTask化せず、S1〜S4とL1-M2が完了・Merge済みであること、S1・S2の実現可能性、S3・L1-M2の契約Ownerと利用方向、S4・S1・S2の網羅性、DD-001〜DD-007の残置、所有Pathの非重複、ADR採用版の追跡可能性を判定する。

## worktree所有境界・Merge順

| 小分類 | 単一Owner |
| --- | --- |
| L1-M1-S1 | `docs/architecture/topology/**`、`docs/architecture/trust-boundaries/**`、`docs/architecture/data-flows/**` |
| L1-M1-S2 | `docs/architecture/decisions/**`、`docs/architecture/technology-stack/**` |
| L1-M1-S3 | `docs/architecture/modules/**`、`docs/architecture/dependencies/**`、`docs/architecture/ownership/**`、共有Index |
| L1-M1-S4 | `docs/architecture/configuration/**`、`docs/architecture/secrets/**` |
| 計画管理Owner | `docs/task-map.md`、Task Map議論記録、G0承認記録 |

- Merge順は `S1 → S2 → S3 → S4` とする。S3・S4の作業自体は並行可能とする。
- S1・S2・S4は共有Indexを直接変更せず、S3へ登録を依頼する。
- S3がS4のPathを所有表へ登録し、S4をS3 Merge後へrebaseしてからMergeする。
- 各小分類worktreeはTask Map・議論記録・G0記録を変更しない。

## 監査結果

- 分割担当: 4件は独立成果物を持ち、小分類として同程度の粒度と判定した。
- 妥当性確認担当: 4件とも維持可能であり、統合・分割・移動・削除は不要と判定した。
- 統合時の修正:
  - S1の名称と成果物を技術非依存の論理トポロジーへ限定し、原典外の配備設計への膨張を防いだ。
  - S2にS1との実現可能性逆照合、ADRの見直し条件、機能固有DDへの移管を追加した。
  - S4の直接依存をS2だけに整理し、S1は推移依存とした。
  - S3・S4は並行作業し、共有Indexと所有表の一意性のためS3を先にMergeすることとした。
  - G0の具体的な整合判定条件を追加した。
- 既存構造への影響: L1-M1、L1-M2〜M4、L2〜L7の承認済み責務と依存は変更不要。

## 承認結果

- L1-M1-S1〜S4の計画状態を「議論中」から「承認済み」へ変更した。
- L1-M1内部DAG、論理Path・Glob Owner、G0具体条件を確定した。
- 次の議論を`L1-M2`の小分類分割へ進めた。
- ユーザーから全階層を最後までタスク化する指示を受けたため、以降は一段ずつ分割・独立監査し、肥大化を避けて結合可能な候補を統合したうえでTask Mapへ逐次確定反映する。

---

# TM-D-004: L1-M2の小分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L1-M2 共有ドメインモデル・状態語彙・機能間契約を設計する`
- 今回決めた階層: L1-M2直下の小分類のみ
- 決定日: 2026-08-11
- 承認状況: 全階層タスク化のユーザー指示に基づき、分割・独立監査後に確定
- Task Map反映: L1-M2-S1〜S3を「承認済み」として反映済み

## 小分類

### L1-M2-S1 共有ドメイン語彙・論理モデル・識別子・状態Ownerを定義する

#### 分割した理由

L2〜L6が共通に参照する概念、識別子、不変条件、状態・データOwnerを、InterfaceとSchemaより先に一意化するため。用語・論理モデル・Ownerは同じ正本を参照するため分離しない。

#### 主成果物・完了条件

- HTML全体、版、生成試行、タグ、配置、公開版参照、利用者の用語集・論理関連図を作成する。
- 識別子・参照規則、不変条件、状態・データOwner Matrixを定義する。
- 失敗生成試行と検証済み不変版、修正元版と生成結果の系譜を区別する。
- タグ・配置はHTML全体、公開版参照は検証済み版へ紐付き、未公開と旧版維持を表現できるようにする。
- 具体的な認証・再試行・公開状態遷移、物理Schema、Repository、UIは対象外とする。
- 直接依存はなく、`docs/domain/glossary/**`、`model/**`、`ownership/**`を所有する。

### L1-M2-S2 機能間Interface仕様とMock例を定義する

#### 分割した理由

L2〜L6が相手機能の実装を待たず契約とMockで詳細設計できるようにし、MockをInterfaceと同じOwner・変更単位で管理するため。

#### 主成果物・完了条件

- GenerationRequest、修正元版参照・修正指示、検証済み結果、失敗試行要約、不変版参照、公開版参照、管理者Principal、匿名公開Projection、タグ・配置参照を定義する。
- 提供側・利用側、必須・任意項目、論理エラー分類、互換性規則を定義する。
- 初回生成、修正生成、手動再試行、失敗、未承認、旧版維持、非公開、権限なしのCanonical Mock例を作成する。
- 内部詳細ログを利用者向け契約へ含めず、HTTP・RPC・serialization・UI DTO・実行Fixtureは対象外とする。
- 直接依存はS1。`docs/contracts/interfaces/**`、`docs/contracts/examples/**`を所有する。

### L1-M2-S3 論理SchemaとMigration統治方針を定義する

#### 分割した理由

論理データ構造と変更統治は同じデータ設計Owner・後続利用者を持ち、分離すると同じ不変条件と順序規則の調整だけが増えるため、一つに結合する。

#### 主成果物・完了条件

- 論理Schema、Cardinality、必須性、一意性、不変条件を定義する。
- S1の概念・状態Owner、S2の契約項目について永続化対象・一時データの対応表を作る。
- Migration変更Owner、番号予約、Merge順、命名、互換性、適用失敗時の扱い、L7共有変更laneへの移管を定義する。
- SQL型、Table・Index名、Repository、物理Migration、無停止更新・downgrade保証は対象外とする。
- 直接依存はS1。S2と並行作業し、S2 Merge後へrebaseして最終化する。
- `docs/data/logical-schema/**`、`docs/data/migration-policy/**`、`docs/domain/README.md`を所有する。

## 依存DAG・worktree

```text
L1-M2-S1 ─┬─→ L1-M2-S2 ──────────┐
           └─→ L1-M2-S3（並行）───┴─→ L1-M2完了
```

- 作業はS2・S3を並行し、Mergeは `S1 → S2 → S3` とする。
- S2を契約形状、S3を永続・一時対応表とMigration統治の正本とする。
- `docs/domain/README.md`はS3の単一Ownerとし、M1-S3の共有IndexはArchitecture側Indexに限定する。
- Task Map、議論記録、G0記録は計画管理Ownerだけが変更する。

## 監査結果

- 分割担当・妥当性確認担当とも3件を維持し、追加分割・統合・移動・削除は不要と判定した。
- MockをInterfaceへ、Migration統治を論理Schemaへ結合し、細分化による調整コストを避けた。
- 要件から語彙・不変条件、Interface、論理Schemaまでの追跡可能性をG0条件へ追加した。
- 既存L1-M1、L1-M3・M4、L2〜L7の責務・依存変更は不要と判定した。

---

# TM-D-005: L1-M3の小分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L1-M3 起動可能な最小Webアプリ・開発テスト基盤を構築する`
- 決定日: 2026-08-11
- 承認状況: 全階層タスク化のユーザー指示に基づき、分割・独立監査後に確定

### L1-M3-S1 最小Webアプリ骨格・共通Toolchain・安全な設定雛形を構築する

#### 分割した理由

Project構造、依存解決、共通Command、最小BootstrapはRoot Manifest・Lockfileと相互依存するため、同じ単一Ownerへ結合する。

#### 主成果物・完了条件

- G0で確定した技術・Directoryに従うProject骨格、Root Manifest・Lockfile、Tool Versionを実装する。
- Build・起動・Test executor・採用時のLint/Format、最小Bootstrapとreadinessを実装する。
- 秘密値なし設定雛形、安全な設定不足エラー、採用時のCodegen共通入口を整備する。
- clean checkoutから依存解決、Build、起動・停止、基礎Test commandを再現する。
- 機能Router、業務Port配線、最終composition root・DI・Registry、本番Hostingは対象外とする。

### L1-M3-S2 基盤Smoke・初期共有Fixture・コード生成再現性を整備する

#### 分割した理由

S1の土台を再現可能な実行証跡で検証し、Smoke・Fixture・Codegen検証は同じTest harnessと後続利用者を持つため一つに結合する。

#### 主成果物・完了条件

- Bootstrapの起動・終了・readiness・基本設定読込を確認するSmokeを実装する。
- 共通初期化、Process起動補助、一時領域、Fixture loader/builderを持つTest harnessを整備する。
- 技術用の決定論的最小Fixtureだけを用意し、機能固有状態・業務Schema・秘密値を含めない。
- Codegen採用時だけ入力正本、出力先、再生成Command、差分検出、生成物Ownerを登録し再現性を確認する。
- Browser/E2E、受入基準、契約互換・Migration検証、機能固有Testは対象外とする。

## 依存・Owner

```text
G0 → L1-M3-S1 → L1-M3-S2 → L1-M3完了 → L1-M4
```

- S1はRoot共有FileとBootstrap、S2はSmoke・Harness・Fixture・Codegen検証を所有する。
- S2で追加依存やBootstrap修正が必要なら、S1 Owner Commitを先にMergeしS2をrebaseする。
- L1-M3内は直列Mergeとし、無理なworktree並行化を行わない。

## 監査結果

- 2候補とも維持し、追加分割・統合・移動・削除は不要と判定した。
- Build・起動・Test・LintをS1、Smoke・Fixture・Codegen再現性をS2へ結合し、Root共有Fileの併有を防いだ。
- L1-M4、L7および他分類の既存責務・依存変更は不要と判定した。

---

# TM-D-006: L1-M4の小分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L1-M4 共有契約とSchema・Migration実行基盤を実装する`
- 決定日: 2026-08-11
- 承認状況: 全階層タスク化のユーザー指示に基づき、分割・独立監査後に確定

### L1-M4-S1 共有識別子・契約型・Portを実装し互換性を検証する

#### 分割した理由

共有型、Port、Serialization、Canonical Mock互換Testは同じ契約変更とSource Pathで同時に変わるため、一つに結合する。

#### 主成果物・完了条件

- L1-M2の識別子・Interfaceを共有識別子、契約型、Portへ実装する。
- Serialization roundtrip、Compile適合、Canonical Mock互換Test、Test用Port実装を整備する。
- 公開向け契約へ認証情報・管理データ・内部Logを混入させない。
- HTTP/RPC Handler、Repository、業務Service、UI DTO、機能固有状態遷移は対象外とする。
- 契約Codegen採用時は契約固有入力と生成物だけを所有し、共通入口・再現性規則はM3へ残す。

### L1-M4-S2 Schema・Migration実行検証基盤を実装する

#### 分割した理由

Migration runner、履歴・失敗処理、初期Migration、Schema Fixture、実行Testは同じMigrationライフサイクル・番号・Pathを所有するため、一つに結合する。

#### 主成果物・完了条件

- Migration適用、適用済み判定、失敗報告、隔離Test環境を実装する。
- 必要な場合だけMigration管理用の空または基盤Migrationと技術Schema Fixtureを作る。
- 初回適用、適用済みSkip、順序違反、失敗状態を検証する。
- 業務Table、個別Migration、Repository、Rollback・無停止更新・Downgrade保証は対象外とする。
- L1期間中の初期Migration番号を単独管理し、L1完了後にL7共有変更laneへ移管する。

## 依存・Owner

```text
G0 + L1-M3完了・依存固定Commit
       ├─→ L1-M4-S1 ────────────┐
       └─→ L1-M4-S2（並行）─────┴─→ L1-M4完了
```

- S1・S2は並行作業し、Mergeは `S1 → S2` とする。
- Root Manifest・LockfileはM3-S1、共通Codegen規則はM3-S2、契約生成物はM4-S1、Schema生成物はM4-S2が所有する。
- M3の技術Fixture、M4-S1の契約Test vector、M4-S2のSchema Fixture、機能Fixture、L7統合Fixtureを分離する。

## 監査結果

- 2候補を維持し、統合・追加分割・移動・削除は不要と判定した。
- S2のS1待ちは着手依存ではなくMerge依存とし、安全な並行作業を維持した。
- L1-M2を意味・論理設計の正本、L1-M4を実行可能実装として境界化した。
- L1-M3、L2〜L7の承認済み構造修正は不要と判定した。

---

# TM-D-007: L2の中分類分割

## 議論状態

- 状態: 承認済み
- 対象: `L2 単一管理者のGoogle認証と管理認可を実現する`
- 決定日: 2026-08-11
- 承認状況: 全階層タスク化のユーザー指示に基づき、分割・独立監査後に確定

### L2-M1 Google認証とSessionライフサイクルを実現する

#### 分割した理由

Google本人確認、Callback、Session発行・更新・失効、Login UIは同じ認証設定・状態・失敗処理を共有するため一つに結合する。

- 出力はVerifiedGoogleIdentityとし、Google本人確認だけで管理者権限を付与しない。
- Google Flow、Claim・Callback防御、Cookie・Session・Logout・CSRF、Login・期限切れUI、認証固有Testを所有する。
- 対象外アカウントへ永続利用者・管理者Sessionを作らず、管理者SessionはM2の許可判定後だけ発行する。
- 管理者email照合、管理API認可、生成HTML隔離は対象外とする。

### L2-M2 単一管理者認可と管理面アクセス境界を実現する

#### 分割した理由

事前登録email照合、Backend強制認可、Frontend表示補助、対象外アカウント拒否は同じ認可Matrixを共有するため一つに結合する。

- 検証済みemailとServer専用の単一設定値を照合し、AuthorizationDecision・AdministratorPrincipalを返す。
- Backend Guardを強制境界、Frontend Guardを表示・遷移補助とする。
- 設定未指定・不正・複数値はfail closedとし、対象外アカウントへ管理API・UIを許可しない。
- 期限切れUI、Google Protocol、Session機構、生成HTML隔離は対象外とする。

## 依存・G1・Owner

- G0後に両者の詳細設計を並行し、L2 G1で認証Flow、Session、認可Matrix、Schema要否、L6境界を判定する。
- L1完了とG1後にMockで並行実装し、Mergeは `M1 → M2` とする。
- Manifest・Lockfile、共通設定、Migration番号・共有Path、全体Router・DIはL7共有変更laneが所有する。
- L2はPrincipalと認可判断をL6へ渡し、公開Route・匿名閲覧・生成HTML表示隔離はL6へ残す。

## 監査結果

- 2候補を維持し、工程別の追加分割や統合は不要と判定した。
- Session期限切れUIをM1、非管理者拒否UIをM2へ一意化した。
- L7共有変更laneはL1完了直後から統合役として利用できることを確認した。

---

# TM-D-008: L2-M1の小分類分割

- 状態: 承認済み
- 対象: `L2-M1 Google認証とSessionライフサイクルを実現する`
- 決定日: 2026-08-11

### L2-M1-S1 Google認証・Session・M2連携を詳細設計する

分割した理由：  
Login開始からCallback、M2認可、Session発行・LogoutまでのSequenceと安全性判断を一つのG1入力として固定し、S2・S3・M2の循環を防ぐため。

- 認証・Session状態遷移、共有契約、Callback・Cookie・CSRF・Session fixation対策、UI状態Owner、Schema要否、Test Matrixを成果物とする。
- M2のemail照合・Guard内部は所有せず、合同G1はTask化しない。

### L2-M1-S2 Google Identity認証Flowを実装する

分割した理由：  
Google Protocol・認証Transaction・Callback検証を、Session保存から独立したVerifiedGoogleIdentity提供能力として完成できるため。

- Login開始・Callback、Provider Adapter、Claim検証、Transaction一度限り消費、Provider Stub・固有Testを所有する。
- email管理者照合、管理者Session、非管理者拒否UIは実装しない。

### L2-M1-S3 M2許可後の管理者Sessionライフサイクルを実装する

分割した理由：  
M2認可後のSession発行・更新・失効、Cookie・CSRF・Logout・期限切れUIが同じSecurity境界を形成するため。

- M2のallow時だけSessionを作成し、deny・error・timeoutではRepository書込みもCookie発行もしない。
- M2 Fakeで独立完了可能とし、実M2統合はL2全体の完了条件へ置く。

## DAG・Owner・監査結果

- G0後にS1を作りM2設計と合同G1、L1完了+G1後にS2/S3を並行、Mergeは `S2 → S3` とする。
- S2は認証Transaction、S3はSessionのSchema意味Owner、Migration共有FileはL7 laneが所有する。
- 3候補を維持し、Test・Schemaだけの追加Taskは作らないと判定した。

---

# TM-D-009: L2-M2の小分類分割

- 状態: 承認済み
- 対象: `L2-M2 単一管理者認可と管理面アクセス境界を実現する`
- 決定日: 2026-08-11

### L2-M2-S1 認可Matrix・設定・強制境界を詳細設計する

分割した理由：  
M1の認証・Session、管理者判定、管理面Guard、L6 Previewを実装前に整合させる認可設計Packageが合同G1の独立入力になるため。

- 認証認可Matrix、管理Route分類、email設定・照合、Decision/Principal条件、401/403/UI遷移、Schema要否、Test Matrixを定義する。
- M1のGoogle Protocol・Session機構、L6の隔離方式は再設計しない。

### L2-M2-S2 VerifiedGoogleIdentityを単一管理者に照合しAuthorizationDecisionを提供する

分割した理由：  
単一管理者設定とverified identityから安全な判断を返すPolicyは、SessionやRouteから独立して再利用できるため。

- 設定不備、未検証・不一致identityはdenyし、allow時だけ最小AdministratorPrincipalを返す。
- 原則statelessとし、利用者台帳・Decision・Role状態を保存しない。

### L2-M2-S3 管理API・UI Guardと非管理者拒否体験を実装する

分割した理由：  
Policy判断を管理面で強制するBackend Guardと表示補助Frontend Guardは、管理者判定とは別のSecurity成果であるため。

- Backendを唯一の強制境界とし、Frontend迂回でも拒否する。
- 未認証・期限切れはM1、認証済み非管理者はS3の拒否UIへ振り分ける。
- Guardと保護Route契約を提供し、全体Router mountはL7へ残す。

## DAG・Owner・監査結果

- S2/S3はMockで並行し、最終Mergeは `L2-M1 → S2 → S3` とする。
- S2はemail Policy、S3はGuard、L6は匿名Route・Preview隔離、L7はRoot共有Fileと全体配線を所有する。
- 3候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-010: L3の中分類分割

- 状態: 承認済み
- 対象: `L3 解説HTMLを生成・形式検証・再試行し生成ログを記録する`
- 決定日: 2026-08-11

### L3-M1 Codex app-serverで単一生成試行を実行する

分割した理由：  
Providerの1呼出しを再試行状態から切り離し、GenerationRequestから候補HTMLまたは安全なProvider失敗を返す独立Adapter境界にするため。

### L3-M2 生成結果のHTML形式を機械検証する

分割した理由：  
Codex接続に依存せず、決定的な判定条件とTest corpusで単独検証できるため。

### L3-M3 Generation Run・Attemptを統括し記録・管理操作を提供する

分割した理由：  
再試行、Attempt履歴、要約Log、手動再試行、生成UIは同じRun状態Ownerを共有し、分離すると回数・履歴の正本が重複するため中分類では結合する。

## 決定事項・監査結果

- 初回1回＋自動再試行最大3回、1 Run最大4 Attemptとする。
- 手動再試行は旧Run参照を持つ新Runとする。
- M1/M2/M3はG0後に詳細設計し合同G1、G1後にMockで並行実装、M1/M2後にM3を統合する。
- L3は生成・検証・試行履歴、L4は修正Request編成・不変版保存・版履歴を所有する。
- 3候補を維持し、追加統合・分割・移動・削除は不要と判定した。

---

# TM-D-011: L3-M1の小分類分割

- 状態: 承認済み
- 対象: `L3-M1 Codex app-serverで単一生成試行を実行する`
- 決定日: 2026-08-11

### L3-M1-S1 Codex単一試行契約・接続を詳細設計する

分割した理由：  
Provider Port、Lifecycle、Request/Response写像、失敗分類をG1前に固定し、M3のMock実装を先行可能にする独立設計Packageだから。

### L3-M1-S2 Adapter・Provider Stub・Contract Testを実装する

分割した理由：  
Adapter、Stub、Mapping、Lifecycle component、Contract Testは同じProvider Path・依存・完了条件を共有するため一つに結合する。

## 決定事項

- 固定生成timeoutは設けず、技術的deadlineはG1で根拠を審査する。
- M1はCodex固有写像だけを持ち、修正Request編成はL4、RetryはM3、HTML判定はM2へ残す。
- S2はprotocol fakeによる決定的Testを基本とし、実Codex Smoke/E2EはL7へ渡す。
- 2候補を維持し、追加分割・統合は不要と判定した。

---

# TM-D-012: L3-M2の小分類分割

- 状態: 承認済み
- 対象: `L3-M2 生成結果のHTML形式を機械検証する`
- 決定日: 2026-08-11

### L3-M2-S1 HTML形式判定規則・非改変原則・安全な失敗契約を詳細設計する

分割した理由：  
DD-004の受理・拒否規則、結果不変条件、失敗契約はL3合同G1へ提出する独立設計成果であり、M3のMock並行実装を成立させるため。

### L3-M2-S2 決定的HTML Validator・境界Corpus・契約Testを実装する

分割した理由：  
Parser Wrapper、型変換、Corpus、Table/Contract Test、Canonical Fakeは同じ判定成果を完成・立証し、同じ機能Pathを所有するため一つに結合する。

## 決定事項・監査修正

- 文書・fragment、DOCTYPE、最低構造、空・BOM・空白・comment/DOCTYPEのみ・平文・破損・truncated・fence・説明文混入・複数文書・文字列やscript内だけのHTML要素・Parser回復境界をS1の決定表に含める。
- Validatorは候補HTMLを補修・抽出・再serializeせず、合格結果には原則として元の`CandidateHtml`を保持する。
- Parserの生Errorを安定Codeへ変換し、内部診断・Stack・候補全文を安全な失敗契約へ露出しない。
- S2がCanonical Fakeと機能内Corpusを所有し、M3は並行中だけ自機能Pathの最小Fakeを使う。
- 内容品質・Sanitize・外部通信隔離はL6、Retry・履歴・生成要約はM3、版保存はL4、Root依存・全体配線はL7へ残す。
- 2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

# TM-D-013: L3-M3の小分類分割

- 状態: 承認済み
- 対象: `L3-M3 Generation Run・Attemptを統括し記録・管理操作を提供する`
- 決定日: 2026-08-11

### L3-M3-S1 Run・Attempt状態、再試行・冪等性・L4受渡しを詳細設計する

分割した理由：  
状態遷移、再試行予算、永続化不変条件、L4受渡し、冪等性はBackendとUIが共有するL3合同G1成果であるため。

### L3-M3-S2 Run・Attempt状態機械、Repository、生成Orchestratorを実装する

分割した理由：  
状態機械、永続化、再試行Schedule、M1/M2呼出しは同じTransactional State Ownerであり、分けるとAttempt番号・再試行予算の正本が重複するため一体化する。

### L3-M3-S3 管理者向け生成・手動再試行・試行ログAPI/UIを実装する

分割した理由：  
Backend状態正本を持たずCommand/Query Port経由で並行開発できる一方、APIとUIは生成開始から失敗確認・手動再試行まで一つの管理者体験を完成させるため結合する。

## 決定事項・監査修正

- 初回生成は単一の自然言語Promptを受け、topic・追加指示の固定項目へ分解しない。空白入力だけを拒否し、内容分類は追加しない。
- 各Provider/Validation失敗を永続記録し、安全な要約と自動再試行中状態を参照可能にする。UI確認を待たず次Attemptへ進み、最終失敗後だけ手動再試行を許す。
- 利用者向け取消は追加せず、System shutdown・接続断・Worker crash時の整合だけを扱う。
- 冪等Command、一意Attempt、ClaimのLease/FencingまたはCASで重複を抑止するが、Provider側保証がない外部呼出しのExactly-onceは約束しない。
- Repository障害は生成失敗と区別し再試行予算を消費しない。呼出し後・結果保存前のCrashでは同一Attemptを盲目的に再実行しない。
- L4受渡しは同じ検証済み結果の再取得・冪等再受渡しを保証し、Push/OutboxまたはPull方式をG1で一意に選ぶ。
- S2は業務Job・Retry/Fencing、L7はMigration番号・共有File・Worker Host・Scheduler起動・Root DIを所有する。
- 3候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-014: L4の中分類分割

- 状態: 承認済み
- 対象: `L4 検証済みHTMLを確認・修正し版履歴を管理する`
- 決定日: 2026-08-11

### L4-M1 検証済みHTMLを不変Versionとして保存し履歴正本を提供する

分割した理由：  
検証済み結果の冪等受入、HTML全体・不変Version・履歴の正本、L6向けRead Portは同じVersion状態Ownerを共有するため一つに結合する。

### L4-M2 管理者がHTMLを確認し修正生成を依頼できるようにする

分割した理由：  
Version選択・確認、隔離Preview導線、自然言語修正指示、L3 Run開始はVersion正本を持たず、同じ管理者ユースケースを完成させるため一つに結合する。

## 決定事項・監査修正

- 初案の設計専用中分類は廃止し、設計内容をM1/M2それぞれの小分類へ移してL2・L3と同じ成果別階層にする。
- L3は検証済み結果と再受渡し境界、M1はHandoff消費・重複排除・Version作成Transactionを所有する。Push/Pullは合同G1で一意に選ぶ。
- 任意履歴Versionは閲覧できるが、修正元は最小範囲として最新作成Versionに限定し、過去版からの分岐を必須化しない。
- L4は確認画面ShellとPreview導線、L6は隔離描画Surfaceを所有し、L4管理DOMへRaw HTMLを直接挿入しない。
- 公開状態・公開Version参照はL6、タグ・配置はL5、Migration番号・共有File・Root配線はL7へ残す。
- 3中分類案を2中分類へ統合し、中分類は成果責任、小分類は設計と実装の一貫した粒度に修正した。

---

# TM-D-015: L4-M1の小分類分割

- 状態: 承認済み
- 対象: `L4-M1 検証済みHTMLを不変Versionとして保存し履歴正本を提供する`
- 決定日: 2026-08-11

### L4-M1-S1 不変Version・生成結果の冪等受入・履歴／読取契約を詳細設計する

分割した理由：  
Version不変条件、Handoff Transaction、同時受入、履歴順序、Repository契約、L6読取境界は実装とMock並行の前にG1で固定する独立成果であるため。

### L4-M1-S2 Version Aggregate・Repository・Handoff Consumer・履歴Queryを実装する

分割した理由：  
重複排除、Version作成・採番、Commit、Repository、履歴Queryは同じTransactional State Ownerであり、分けると冪等性と順序の正本が分散するため結合する。

## 決定事項・監査修正

- Push/OutboxまたはPullの一方式をS1完了時に選び、配送Retry Ownerを一つにする。
- 初回Versionは修正元なし、修正Versionは実際の修正元IDを保持し、不正参照をGeneration失敗・Retryへ混入させない。
- 「現在確認中」はM1へ永続化せず、M2のURL/画面状態とする。公開中VersionはL6が所有する。
- Handoff受入とVersion作成を同一Transactionにし、重複・同時受入は一意制約とConcurrency制御で同じVersionRefへ収束させる。
- Repository失敗・Ack消失時は同一Handoffを再処理し、Codexを再実行しない。
- 管理用Metadata QueryはRaw HTMLを返さず、L6 Server-side隔離Surface専用Content Portだけが不変本文を取得する。
- Version履歴は生成履歴に限定し、公開承認・公開切替履歴はL6へ残す。
- 2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-016: L4-M2の小分類分割

- 状態: 承認済み
- 対象: `L4-M2 管理者がHTMLを確認し修正生成を依頼できるようにする`
- 決定日: 2026-08-11

### L4-M2-S1 Version選択・修正元適格性・修正Command・隔離Preview連携を詳細設計する

分割した理由：  
画面選択、修正元規則、L3 Command、L6 Preview Consumer、認可、冪等性はAPI/UI実装が共有するL4 G1成果であるため。

### L4-M2-S2 管理者向けVersion確認・修正生成・履歴・Preview導線API/UIを実装する

分割した理由：  
Version確認、履歴表示、Preview導線、自然言語修正、L3 Run開始は一つの管理者体験とDTOを共有し、分けても状態Ownerを持たないため結合する。

## 決定事項・監査修正

- 任意履歴Versionは`selected_version_id`として閲覧できるが、修正元は最新作成Versionに限定する。過去版からの分岐生成は要件拡張として再承認する。
- 選択状態はURL/画面状態を正本とし、DBへCurrent Pointerを保存せず、新Version作成時にも勝手に切り替えない。
- 修正送信時はBackendでVersionの存在・所属・最新適格性を再確認し、単一自然言語指示と冪等Command IDをL3へ渡す。
- L4はL3 Run/Attempt/Retry/失敗ログを保存せず、Read ProjectionまたはL3画面への導線を使う。
- L4はVersionRefを渡すPreview Consumer契約だけを持ち、L6具体設計に先行依存しない。Raw HTMLへFallbackしない。
- L4はPreview Stubで機能内完了し、実HTML確認・隔離統合はL6、入力から修正までのE2EはL7/G2で完了する。
- API/UIを結合した2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-017: L5の中分類分割

- 状態: 承認済み
- 対象: `L5 HTML全体のタグとサイト内配置を管理する`
- 決定日: 2026-08-11

### L5-M1 HTML全体に紐付くタグ・配置Metadataの正本を提供する

分割した理由：  
TagとPlacementは同じHtmlId・検証済み適格性・Version非依存ライフサイクル・Read Projectionを共有し、別中分類にすると状態境界が重複するため結合する。

### L5-M2 管理者がタグ・配置を管理できるAPI/UIを提供する

分割した理由：  
Metadata正本を持たずM1 Portで並行開発でき、Tag/Placement操作は同じHTML管理画面・認可・受入体験を共有するためAPI/UIを結合する。

## 決定事項・監査修正

- Tagは多対多で作成・改名・付与・解除を含み、削除・統合・階層化は追加しない。
- Placementは任意名作成/初回設定・選択・置換、`HTML 0..1 → Placement`に限定し、改名・削除・階層・並び順・複数配置は追加しない。
- 適格性はL4の検証済みHTML Aggregate Portで確認し、失敗AttemptやClient申告を使わずL5へ検証状態を複製しない。
- 名前の空白・重複・正規化はL5 G1で最低限を決め、黙って利用者入力を狭めない。
- L6 Projectionは現在Metadataだけとし、公開状態・匿名Route・NavigationはL6、共有Migration/Root配線はL7へ残す。
- タグ/配置を状態正本と管理体験の2中分類へまとめ、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-018: L5-M1の小分類分割

- 状態: 承認済み
- 対象: `L5-M1 HTML全体に紐付くタグ・配置Metadataの正本を提供する`
- 決定日: 2026-08-11

### L5-M1-S1 Tag・Placement Domain、HTML関連Commandの適格性、Port、Schemaを詳細設計する

分割した理由：  
Cardinality、Command別適格性、名前規則、Version非依存性、L4/L6 Portは実装とMock並行前に合同G1で固定する独立成果であるため。

### L5-M1-S2 Tag・Placement MetadataのDomain・Repository・Application・最小Read Projectionを実装する

分割した理由：  
TagとPlacementは同じHtmlId・Metadata Root・適格性不変条件・Schema領域・Projectionを共有し、別Taskにすると境界が重複するため結合する。

## 決定事項・監査修正

- L4適格性確認はTag初回付与・Placement初回設定など新しいHTML関連を成立させる時だけ行い、Tag/Placement作成、Tag改名・解除には要求しない。
- 適格性確認後にMetadata Rootを作り、関連が空になっても保持して以後の適格性不変条件とする。
- Placementは未設定から初回設定、その後は別Placementへの置換だけとし、解除は追加しない。
- 名前制約はG1でBackend/UI/Schema間を統一するが、任意名の文字種を過度に狭めない。
- L6 Projectionは現在Metadataだけに限定し、公開絞込・Navigation・並び順を追加しない。
- 並行性を保つためL4 Tableへの物理FKを設けず、共有HtmlId・適格性PortとL7 Adapter統合で論理整合する。
- 2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-019: L5-M2の小分類分割

- 状態: 承認済み
- 対象: `L5-M2 管理者がタグ・配置を管理できるAPI/UIを提供する`
- 決定日: 2026-08-11

### L5-M2-S1 Tag・Placement管理API、Feature-local管理Panel、認可・冪等性・エラー契約を詳細設計する

分割した理由：  
DD-003、HtmlId単位のExtension、Backend認可、冪等性、競合、エラー表示は実装前にL5合同G1で固定する独立成果であるため。

### L5-M2-S2 Tag・Placement管理API、Feature-local管理Panel、固有Testを実装する

分割した理由：  
APIとUIは同じ管理者操作・DTO・Guard・エラー契約を共同で完成させ、分けると操作結果の受入責任が分散するため結合する。

## 決定事項・監査修正

- Panel入力はHtmlIdと管理文脈に限定し、VersionId/Raw HTMLを受けず、L4 Page本体を直接編集しない。
- 強いID型を混同せず、不正/未知/不適格HtmlIdとClient申告FlagをBackendで拒否する。
- M1の名前Policyを参照し、作成冪等Key、rename競合、関連操作の再送、Placement置換を整合させる。
- Backend Guardを全管理Command/Queryの強制境界とし、Frontend Guardは補助に限定する。
- S2はFake Slot Hostで完了し、実Guard・L4 Slot Mount・Global Router/DIはL7へ引き渡す。
- 匿名公開画面・公開状態との合成・公開Navigationの有無は対象外とし、L6へ未確定のNavigation実装を要求しない。
- 2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-020: L6の中分類分割

- 状態: 承認済み
- 対象: `L6 公開版状態と隔離された生成HTML閲覧を実現する`
- 決定日: 2026-08-11

### L6-M1 公開Version Pointerの正本と管理者向け公開操作を提供する

分割した理由：  
公開承認・切替・非公開化は同じPublication Pointer・Transaction・管理操作を共有し、隔離描画方式とは独立した業務状態であるため。

### L6-M2 管理Previewと匿名公開をResolver分離・Renderer共通の隔離HTML Surfaceで提供する

分割した理由：  
Preview認可と匿名公開判定は別Resolverとしつつ、危険な生成HTMLに対する資格情報なしRenderer・Header・Security Testを共用して欠陥差を防ぐため。

## 決定事項・監査修正

- PublicationはHtml単位のPointerなし/1件だけを正本とし、新Version生成で従来Pointerを変えない。
- Public Projectionは現在公開中の参照だけを返し、Raw HTML・Version履歴・過去承認状態を持たない。
- 匿名RouteはPublic Handleから現在Pointerだけを解決し、任意Version迂回を許さず、非公開系Responseを同じ外形にする。
- 管理Preview認可はRenderer外で完了し、対象固定の短命Capabilityだけを隔離Originへ渡して管理資格情報を渡さない。
- 外部通信を一律禁止せず、Cookie非共有Origin、CORS/CSRF/Fetch Metadata、内部宛遮断により管理面を隔離する。
- M2は安全条件とFail-closed Config契約、L7は実Host/Proxy値と全体配線を所有する。
- L5 Tagは直接依存にせず、AC-010に必要なPlacement最小Projectionだけを必要時に接続し、Navigation等を追加しない。
- 2候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-021: L6-M1の小分類分割

- 状態: 承認済み
- 対象: `L6-M1 公開Version Pointerの正本と管理者向け公開操作を提供する`
- 決定日: 2026-08-11

### L6-M1-S1 Publication Pointer・状態遷移・L4適格性・冪等性・Projection／管理Slot契約を詳細設計する

分割した理由：  
Publication状態、L4適格性、Command冪等性/競合、Public Projection、Panel Slotは状態正本と管理体験が共有する合同G1成果であるため。

### L6-M1-S2 Publication Aggregate・Repository・Application・Revision付きPublic Projectionを実装する

分割した理由：  
Pointer、Repository、Command処理、Revisionは同じ状態正本とLocal Transactionを共有するため一体化する。

### L6-M1-S3 管理者向け公開API・Feature-local Publication Panel・固有Testを実装する

分割した理由：  
管理API/Panelは同じ公開操作体験を形成する一方、Backend正本とは別Path・Mock境界を持つためS2と分離する。

## 決定事項・監査修正

- L4適格性確認後、PublicationだけのLocal TransactionでCAS・Pointer・Dedupeを保存し、L4を分散Transactionへ参加させない。
- CommandIdとExpected Revisionを別目的で使い、同一Pointerの再設定はNo-opとしてRevisionを増やさない。
- 公開正本は現在Pointerだけとし、過去承認履歴・Version別Approved Flagを追加しない。
- Public Projection RevisionはS2、Renderer Cache/再検証はM2、実Cache/Event配線はL7が所有する。
- 任意履歴Versionの公開と、最新Versionだけを修正元にするL4 Correctionを型・文言・Commandで分離する。
- S3はFeature-local PanelとFake Hostで完了し、L4 Page実MountはL7へ残す。
- 3候補を維持し、追加分割・統合・移動・削除は不要と判定した。

---

# TM-D-022: L6-M2の小分類分割

- 状態: 承認済み
- 対象: `L6-M2 管理Previewと匿名公開をResolver分離・Renderer共通の隔離HTML Surfaceで提供する`
- 決定日: 2026-08-11

### L6-M2-S1 Preview／Public解決・Capability交換・隔離Surface・Cache／Origin境界を詳細設計する

分割した理由：  
Resolver、Capability、Renderer、Origin、Cacheを別々に設計すると迂回経路が残るため、一つのThreat Model・契約PackageとしてL6合同G1で確定する。

### L6-M2-S2 Capabilityを生成HTMLへ露出せず消費するPreview Resolverと現在Pointerだけを解決するPublic Resolverを実装する

分割した理由：  
PreviewとPublicは別Trust Contextを維持しつつ、Rendererへ安全な`ResolvedRenderTarget`を渡す解決境界と迂回防止Testを共有するため結合する。

### L6-M2-S3 ResolvedRenderTargetだけを受け取るCredentialなしRenderer・隔離Surface・Version本文Cache・Multi-origin Security Harnessを実装する

分割した理由：  
Rendererの安全性はHeader・Origin・Cacheだけで完了せず、外部通信成功と管理面到達失敗をMulti-origin Harnessで同時検証して一つの受入成果になるため結合する。

## 決定事項・監査修正

- Capabilityは隔離OriginのPreview Resolverだけが受け取り、生成HTML実行前に原子的に消費する。最終URL、DOM、Referer、JS可視Cookie、Rendererへ残さない。
- RendererはS3が所有する`ResolvedRenderTarget`だけを受け、Capability、Principal、Session/Cookie、Google Token、CSRF、Public Handle、元Request Headerを受けない。
- Public ResolverはHandleからM1 Projectionを毎回解決し、Pointer独自Cacheを持たない。匿名失敗は同じ404相当本文と`no-store`へ収束する。
- M1-S2はPointer/Revision、M2-S2はPointer解決とCapability状態、M2-S3はVersion本文Cache/Header、L7は実Cache backend/失効配線を所有する。
- 外部HTTPSを一律禁止せず、Credential非送信と管理API側防御で管理データ読取・状態変更を防ぐ。Server-side Proxy利用時の内部宛Network遮断と実Host/Proxy値はL7が所有する。
- S3は安全条件、Feature設定、Route要件、Test Server/Harnessまでを持つ。実Host/Port/TLS/Proxy、Global Router/DI、横断E2EはL7へ残す。
- S3が`ResolvedRenderTarget`の最小Contract Seedを先行Mergeし、S2/S3を別worktreeで並行可能にする。
- 3候補を維持し、S3の追加分割、S2/S3の統合、Preview/Public Resolverの別Task化は不要と判定した。

---

# TM-D-023: L7の中分類分割

- 状態: 承認済み
- 対象: `L7 初期リリースを統合し受入テストを自動化する`
- 決定日: 2026-08-11

### L7-M1 共有変更を逐次統合し再現可能なRelease Baselineを確定する

分割した理由：  
Root共有Fileの単一Owner laneを、未反映要求がなくClean Checkoutから再現可能な最終Baselineという完了成果へ収束させるため。

### L7-M2 全機能Runtimeを構成し実行環境・隔離境界を成立させる

分割した理由：  
Composition、Router/DI、実Adapter、Worker、Codex lifecycle、実Origin/Cacheは一つの起動可能Runtimeへ収束し、機能固有ロジックとは別の統合成果であるため。

### L7-M3 AC-001〜AC-029の自動受入とG2判定用証跡を作成する

分割した理由：  
Trace Matrix、統合Fixture、横断E2E、Release Evidenceは全受入基準を判定可能にする一つの成果であり、M1/M2の実装責任から分離できるため。

## 決定事項・監査修正

- M1は機能開発中の長寿命laneだが、未反映要求なし、共有File整合、Codegen差分なし、Clean Build/Test/Migration成功、最終Baseline Commit記録を完了条件にする。後続変更時は再度開く。
- 共有契約の業務意味はL1/該当機能、機能固有生成物は各機能、M1は物理統合・生成物台帳・最終再現性だけを所有する。
- M2はDescriptor/PortをMountして機能ロジックを再実装せず、実Network PolicyはReference/CI topologyまでとする。
- 通常Blocking TestはCodexのProtocol-faithful Test Doubleを使う。実Codex Smokeは管理環境・Versionを固定してG2証跡とし、環境不足Skipを合格にしない。
- AC-010はPlacement Repositoryの保持だけでなく、統合Site Slotへの反映と公開Version切替後の維持までを横断E2Eで証明する。
- 全29件を生成/Retry/ログ、Preview、Metadata、認証、公開状態、隔離、外部通信、Version横断整合のScenario群へ割り当て、Skip/QuarantineなしでTest ID・Fixture・環境・結果を追跡する。
- L7完了条件からG2通過を外し、`L7成果完成 → G2判定 → 初期リリース完了`の非循環DAGへ修正した。
- 3候補を維持し、詳細設計専用中分類や機能固有Testの再実装は追加しない。

---

# TM-D-024: L7-M1の小分類分割

- 状態: 承認済み
- 対象: `L7-M1 共有変更を逐次統合し再現可能なRelease Baselineを確定する`
- 決定日: 2026-08-11

### L7-M1-S1 共有変更laneの統治契約を確定しL1成果を引き受ける

分割した理由：  
L1からの実引受、変更要求/承認Matrix、Baseline状態、台帳初期化とDry-runは、各機能の並行統合前に一度完了できる独立成果であるため。

### L7-M1-S2 共有変更を逐次統合しRelease Baselineを確定する

分割した理由：  
Manifest、Migration、共有契約物理表現、CodegenとClean検証は同じ単一Writer/共有Commit/Baselineを持ち、分けるとOwnerが再分散するため結合する。

## 決定事項・監査修正

- S1は台帳形式・初期状態を所有し、完了時に中央Queue、Migration予約、Baseline記録の運用WriterをS2へ移管する。
- S1は文書作成だけでなく、L1 Path/Glob/Commit/Owner/予約/Codegen登録の棚卸し、台帳初期化、架空要求Dry-runまで完了する。
- S2は`Open / Stable / Final / Invalidated`を運用し、Stable CommitをTask完了扱いにせず、Final Baselineだけを完了成果とする。
- Final後の要求や不整合でS2を再開し、旧Baseline IDに基づくM2/M3証跡を失効させる。
- S2は物理統合Ownerであり、共有契約の意味を独自変更しない。影響度に応じてConsumer承認、機能G1、G0を再確認する。
- 共通生成物/台帳はS2、Feature-local生成物と業務的正しさは各機能、S2は全登録生成物の最終再現性を所有する。
- S2のBuild/Lockfile/Codegen/Migration結果を一次成果とし、M3はBaseline IDを参照してG2証跡へ集約する。
- 2候補を維持し、逐次統合と最終検証の分離、Manifest/Migration/Codegen別Taskへの追加分割は不要と判定した。

---

# TM-D-025: L7-M2の小分類分割

- 状態: 承認済み
- 対象: `L7-M2 全機能Runtimeを構成し実行環境・隔離境界を成立させる`
- 決定日: 2026-08-11

### L7-M2-S1 全Feature Descriptor・Port・SlotをRuntime Compositionへ統合する

分割した理由：  
Composition、Global Router、DI/Registry、Slot、横断Glueは同じRoot File群を編集する単一統合成果であり、機能別に分けると競合するため結合する。

### L7-M2-S2 Codex app-serverの実Process lifecycle・設定・Smoke実行物を統合する

分割した理由：  
外部Processの発見・Version・起動停止・Transport・SmokeはL3の生成業務処理とは別の障害/設定境界を持つため分離する。

### L7-M2-S3 L6隔離Surfaceの実Host・Origin・Network・Cache Infrastructureを統合する

分割した理由：  
実Origin/Proxy/Header/Cache backendはL6の安全条件を実行環境へ写す一つのInfrastructure成果であり、Codex ProcessやComposition Fileとは独立するため。

## 決定事項・監査修正

- S1はFeature Descriptor/PortをMountし、機能ロジックや業務Policyを複製しない。Root Configは集約/登録だけを行い、共有物理FileはM1-S2へ要求する。
- Feature固有Adapterは各機能、Codex Process/TransportはS2、Host/Proxy/Cache InfrastructureはS3、単純GlueとDI登録はS1が所有する。
- Placement Projectionを統合Site Slotへ必ずMountし、設定名と対象HTMLの関連を確認可能にする。Navigation等は追加しない。
- S2はL3 AdapterへTransport Clientを渡すHostであり、Prompt写像、Candidate抽出、Typed Failure、Retry/Validationを再実装しない。
- S3はL6のCache/Revision/失効意味を再定義せず、Concrete backendと実Origin/Proxy/Headerへ配線する。
- S2/S3は並行可能とし、S1最終Mount後にM1 Final Baselineを確定する。Runtime Smokeで共有変更が判明した場合はBaselineを再度Openする。
- 実Host/TLSはReference/CI topologyまで、本番Hosting/DNS/証明書/CDN/監視は対象外とする。
- 3候補を維持し、Worker Hostの別Task化、S2/S3の結合、機能別Composition分割は不要と判定した。

---

# TM-D-026: L7-M3の小分類分割

- 状態: 承認済み
- 対象: `L7-M3 AC-001〜AC-029の自動受入とG2判定用証跡を作成する`
- 決定日: 2026-08-11

### L7-M3-S1 AC Trace Matrix・Test Architecture・Integration Fixture Harnessを確立する

分割した理由：  
Manifest/Evidence Schema、決定的Harness、Fixture合成、Skip検出は全Test Packが共有する先行成果であるため。

### L7-M3-S2 生成・検証・再試行・確認修正の横断E2Eを自動化する

分割した理由：  
L3/L4/L6 Previewを跨ぐ生成系AC群は同じRun/Attempt/Version FixtureとBrowser/API/Worker Flowを共有するため結合する。

### L7-M3-S3 タグ・配置・Version非依存性の横断E2Eを自動化する

分割した理由：  
L5 Metadata、Site Slot、L6公開切替を跨ぐAC群は同じHtmlId/Version非依存Fixtureを共有し、生成系や認証系と独立して実行できるため。

### L7-M3-S4 Google認証・管理認可・公開状態の横断E2Eを自動化する

分割した理由：  
L2/L4/L6を跨ぐ認証・Guard・Publication AC群はGoogle互換IdPと管理/匿名Contextを共有し、別worktreeで完結できるため。

### L7-M3-S5 Actual-origin隔離・実Codex Smoke・Release証跡を確定する

分割した理由：  
AC-019/029、実Codex Smoke、Evidence集約は同じRelease Profile/Baselineで実行するため結合し、Gate判断だけの別Taskを作らない。

## 決定事項・監査修正

- 初案の27件単一E2E Taskは依存・Fixture・画面領域が分散し肥大化するため、生成、Metadata、認証/公開の3 Test Packへ分割した。
- Acceptance Manifestを唯一の対応正本とし、各ACにPrimary Scenarioをちょうど1件割り当て、Report/Trace Matrix/Evidenceを機械生成する。
- Google認証はVerified Identity直接注入ではなくGoogle互換IdPから実Callback/State/Nonce/Session経路を通す。
- Validation失敗とProvider失敗は別Scenarioとし、最大4 Attempt、各失敗表示後のRetry開始、最終失敗、手動新Runを確認する。
- AC-002/017は隔離Preview経由、AC-010はSite Slot反映と切替後維持、AC-015/023はBackend Guard、AC-019はPreview/Public双方、AC-029は外部HTTPS成功と管理面失敗を証明する。
- Evidence PackageをBaseline ID/Source/Config/Tool/Provider/結果/Skip/一次Artifact参照へ固定し、Baseline再Open時に無効化して全Profileを再実行する。
- 通常CIはManifest整合とS2〜S4、Release ProfileはM2 Runtime SmokeとS5、G2前Release CheckはM1/M2/S2〜S5/実Codexをすべて必須とする。
- 5候補を維持し、S5のEvidence専用Task化や各AC単位への過分割は不要と判定した。

---

# TM-D-027: 全階層の横断最終監査

- 状態: 承認済み
- 対象: `L1`〜`L7`の全階層
- 決定日: 2026-08-11

## 監査結論

- 大分類7件、中分類18件、小分類48件を原典へ再照合し、FR-001〜011、BR-001〜021、NFR-001〜006、CON-001、DD-001〜007、AC-001〜029の配置に実質的な欠落がないことを確認した。
- 全大分類に中分類、全中分類に最終小分類があり、親成果物は子成果物の和で包含されている。Taskのmerge/split/move/deleteは不要と判定した。
- 状態Owner、Migration/Manifest/Codegen、Composition、Cache、受入証跡の単一Ownerを再確認し、worktree別Pathと共有変更laneの競合回避が成立する。
- G2をL7完了条件から除き、`L7成果完成 → G2判定 → 初期リリース完了`へ統一した。
- 機能本実装の起点をL1完了Commitへ統一し、G0後の設計着手と区別した。
- L6のL5直接依存を解除した。L5 Placement ProjectionはL7 Site Slotが直接消費し、L7 E2EがL6公開切替との横断整合を証明する。
- L7を最後に初めてMergeする表現を廃止し、L1完了後から共有lane、Composition骨格、Test骨格を稼働させる逐次統合DAGへ同期した。
- L7-M2-S2/S3は骨格・実Adapter・最終Smokeの段階依存を区別し、Root共有FileはM1、Config Descriptorは各Task、集約はM2-S1のOwnerとした。
- Acceptance中央IndexはM3-S1のSchema/Tool/出力Pathを正本とし、各PackはFragment、M3-S5はTool実行とEvidence収録だけを行う。
- 最終DAGは非循環で、横断監査指摘を反映後はG0に向けてL1-M1/L1-M2を並行着手可能と判定した。
