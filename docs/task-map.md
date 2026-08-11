# Task Map

## 目的

トピック解説HTML生成・管理Webアプリの初期リリースについて、現在のタスク構造、状態、成果物、依存関係、次の議論位置を一覧できるようにする。

この文書は現在のスナップショットだけを示す。提案の変遷、分割理由、監査結果、承認経緯は[Task Map議論記録](topic-to-html-task-map-discussion.md)に残す。

## Source of Truth

| 情報 | 参照先 |
| --- | --- |
| システム要件・受け入れ基準 | [要件定義](topic-to-html-requirements.md) |
| 要件の議論・決定経緯 | [要件議論マップ](topic-to-html-requirements-discussion-map.md) |
| 分割案・監査結果・承認記録 | [Task Map議論記録](topic-to-html-task-map-discussion.md) |
| 現在のタスク構造・進行位置 | 本文書 |

## ID規則

- `L<n>`: 大分類
- `L<n>-M<n>`: 中分類
- `L<n>-M<n>-S<n>`: 小分類
- 承認済みIDは原則として維持する。
- 議論中のタスクは承認時に名称・IDが変わる可能性がある。

## 状態

タスク構造の確定度と、実作業の進捗を分離して管理する。

### 計画状態

| 状態 | 意味 |
| --- | --- |
| 承認済み | タスクの存在、名称、親子関係が承認されている |
| 議論中 | 提案・監査済みだが、まだ承認されていない |
| 再検討中 | 承認済みだが、後続分割との矛盾に対する修正案があり再承認を待っている |
| 未分割 | 親タスクのみ承認済みで、子タスクをまだ議論していない |

### 実行状態

| 状態 | 意味 |
| --- | --- |
| 未着手 | 成果物の作成を開始していない |
| 進行中 | 成果物を作成中 |
| 完了 | 完了条件を満たした |
| ブロック | 依存関係や外部判断により進行できない |

## 現在位置

- 現在の議論: 全階層の承認・横断監査が完了し、実装着手位置を確定する
- 現在の状態: L1〜L7の全中分類・小分類が承認済み、横断最終監査の指摘反映済み
- 直前の完了: 全要件・全29 AC、粒度、単一Owner、非循環DAG、Gate、worktreeの横断再監査
- 次の議論: G0に向けて`L1-M1`と`L1-M2`を並行着手する
- GitHub Issue対応表: 未作成

## 階層マップ

```text
L1 横断設計・共有契約と最小実装基盤を確立する [承認済み]
├─ L1-M1 横断アーキテクチャ・技術構成・信頼境界を設計する [承認済み]
│  ├─ L1-M1-S1 論理実行トポロジー・横断信頼境界・主要データフローを設計する [承認済み]
│  ├─ L1-M1-S2 横断技術構成を選定しADRを作成する [承認済み]
│  ├─ L1-M1-S3 Module・Directory境界と依存方向を設計する [承認済み]
│  └─ L1-M1-S4 設定・秘密情報の分類と露出境界を設計する [承認済み]
├─ L1-M2 共有ドメインモデル・状態語彙・機能間契約を設計する [承認済み]
│  ├─ L1-M2-S1 共有ドメイン語彙・論理モデル・識別子・状態Ownerを定義する [承認済み]
│  ├─ L1-M2-S2 機能間Interface仕様とMock例を定義する [承認済み]
│  └─ L1-M2-S3 論理SchemaとMigration統治方針を定義する [承認済み]
├─ L1-M3 起動可能な最小Webアプリ・開発テスト基盤を構築する [承認済み]
│  ├─ L1-M3-S1 最小Webアプリ骨格・共通Toolchain・安全な設定雛形を構築する [承認済み]
│  └─ L1-M3-S2 基盤Smoke・初期共有Fixture・コード生成再現性を整備する [承認済み]
└─ L1-M4 共有契約とSchema・Migration実行基盤を実装する [承認済み]
   ├─ L1-M4-S1 共有識別子・契約型・Portを実装し互換性を検証する [承認済み]
   └─ L1-M4-S2 Schema・Migration実行検証基盤を実装する [承認済み]

L2 単一管理者のGoogle認証と管理認可を実現する [承認済み]
├─ L2-M1 Google認証とSessionライフサイクルを実現する [承認済み]
│  ├─ L2-M1-S1 Google認証・Session・M2連携を詳細設計する [承認済み]
│  ├─ L2-M1-S2 Google Identity認証Flowを実装する [承認済み]
│  └─ L2-M1-S3 M2許可後の管理者Sessionライフサイクルを実装する [承認済み]
└─ L2-M2 単一管理者認可と管理面アクセス境界を実現する [承認済み]
   ├─ L2-M2-S1 認可Matrix・設定・強制境界を詳細設計する [承認済み]
   ├─ L2-M2-S2 VerifiedGoogleIdentityを単一管理者に照合しAuthorizationDecisionを提供する [承認済み]
   └─ L2-M2-S3 管理API・UI Guardと非管理者拒否体験を実装する [承認済み]

L3 解説HTMLを生成・形式検証・再試行し生成ログを記録する [承認済み]
├─ L3-M1 Codex app-serverで単一生成試行を実行する [承認済み]
│  ├─ L3-M1-S1 Codex単一試行契約・接続を詳細設計する [承認済み]
│  └─ L3-M1-S2 Adapter・Provider Stub・Contract Testを実装する [承認済み]
├─ L3-M2 生成結果のHTML形式を機械検証する [承認済み]
│  ├─ L3-M2-S1 HTML形式判定規則・非改変原則・安全な失敗契約を詳細設計する [承認済み]
│  └─ L3-M2-S2 決定的HTML Validator・境界Corpus・契約Testを実装する [承認済み]
└─ L3-M3 Generation Run・Attemptを統括し記録・管理操作を提供する [承認済み]
   ├─ L3-M3-S1 Run・Attempt状態、再試行・冪等性・L4受渡しを詳細設計する [承認済み]
   ├─ L3-M3-S2 Run・Attempt状態機械、Repository、生成Orchestratorを実装する [承認済み]
   └─ L3-M3-S3 管理者向け生成・手動再試行・試行ログAPI/UIを実装する [承認済み]

L4 検証済みHTMLを確認・修正し版履歴を管理する [承認済み]
├─ L4-M1 検証済みHTMLを不変Versionとして保存し履歴正本を提供する [承認済み]
│  ├─ L4-M1-S1 不変Version・生成結果の冪等受入・履歴／読取契約を詳細設計する [承認済み]
│  └─ L4-M1-S2 Version Aggregate・Repository・Handoff Consumer・履歴Queryを実装する [承認済み]
└─ L4-M2 管理者がHTMLを確認し修正生成を依頼できるようにする [承認済み]
   ├─ L4-M2-S1 Version選択・修正元適格性・修正Command・隔離Preview連携を詳細設計する [承認済み]
   └─ L4-M2-S2 管理者向けVersion確認・修正生成・履歴・Preview導線API/UIを実装する [承認済み]

L5 HTML全体のタグとサイト内配置を管理する [承認済み]
├─ L5-M1 HTML全体に紐付くタグ・配置Metadataの正本を提供する [承認済み]
│  ├─ L5-M1-S1 Tag・Placement Domain、HTML関連Commandの適格性、Port、Schemaを詳細設計する [承認済み]
│  └─ L5-M1-S2 Tag・Placement MetadataのDomain・Repository・Application・最小Read Projectionを実装する [承認済み]
└─ L5-M2 管理者がタグ・配置を管理できるAPI/UIを提供する [承認済み]
   ├─ L5-M2-S1 Tag・Placement管理API、Feature-local管理Panel、認可・冪等性・エラー契約を詳細設計する [承認済み]
   └─ L5-M2-S2 Tag・Placement管理API、Feature-local管理Panel、固有Testを実装する [承認済み]

L6 公開版状態と隔離された生成HTML閲覧を実現する [承認済み]
├─ L6-M1 公開Version Pointerの正本と管理者向け公開操作を提供する [承認済み]
│  ├─ L6-M1-S1 Publication Pointer・状態遷移・L4適格性・冪等性・Projection／管理Slot契約を詳細設計する [承認済み]
│  ├─ L6-M1-S2 Publication Aggregate・Repository・Application・Revision付きPublic Projectionを実装する [承認済み]
│  └─ L6-M1-S3 管理者向け公開API・Feature-local Publication Panel・固有Testを実装する [承認済み]
└─ L6-M2 管理Previewと匿名公開をResolver分離・Renderer共通の隔離HTML Surfaceで提供する [承認済み]
   ├─ L6-M2-S1 Preview／Public解決・Capability交換・隔離Surface・Cache／Origin境界を詳細設計する [承認済み]
   ├─ L6-M2-S2 Capabilityを生成HTMLへ露出せず消費するPreview Resolverと現在Pointerだけを解決するPublic Resolverを実装する [承認済み]
   └─ L6-M2-S3 ResolvedRenderTargetだけを受け取るCredentialなしRenderer・隔離Surface・Version本文Cache・Multi-origin Security Harnessを実装する [承認済み]

L7 初期リリースを統合し受入テストを自動化する [承認済み]
├─ L7-M1 共有変更を逐次統合し再現可能なRelease Baselineを確定する [承認済み]
│  ├─ L7-M1-S1 共有変更laneの統治契約を確定しL1成果を引き受ける [承認済み]
│  └─ L7-M1-S2 共有変更を逐次統合しRelease Baselineを確定する [承認済み]
├─ L7-M2 全機能Runtimeを構成し実行環境・隔離境界を成立させる [承認済み]
│  ├─ L7-M2-S1 全Feature Descriptor・Port・SlotをRuntime Compositionへ統合する [承認済み]
│  ├─ L7-M2-S2 Codex app-serverの実Process lifecycle・設定・Smoke実行物を統合する [承認済み]
│  └─ L7-M2-S3 L6隔離Surfaceの実Host・Origin・Network・Cache Infrastructureを統合する [承認済み]
└─ L7-M3 AC-001〜AC-029の自動受入とG2判定用証跡を作成する [承認済み]
   ├─ L7-M3-S1 AC Trace Matrix・Test Architecture・Integration Fixture Harnessを確立する [承認済み]
   ├─ L7-M3-S2 生成・検証・再試行・確認修正の横断E2Eを自動化する [承認済み]
   ├─ L7-M3-S3 タグ・配置・Version非依存性の横断E2Eを自動化する [承認済み]
   ├─ L7-M3-S4 Google認証・管理認可・公開状態の横断E2Eを自動化する [承認済み]
   └─ L7-M3-S5 Actual-origin隔離・実Codex Smoke・Release証跡を確定する [承認済み]
```

## 確定した横断境界

| 対象 | 確定した境界 | 主な影響 |
| --- | --- | --- |
| L1 | 横断判断、共有契約、論理Schema・Migration方針、最小骨格に限定 | DD-001〜DD-007の機能固有設計との重複を防ぐ |
| L2・L6 | L2は資格情報・認証・管理認可、L6は生成HTML表示面の隔離を所有 | 認証情報隔離の二重所有を防ぐ |
| L3・L4 | L3は生成・検証・再試行、L4は修正操作の編成と成功版保存を所有 | 初回生成と修正生成で同じ受け渡し契約を使う |
| L4・L6 | L4は不変版と履歴、L6は公開版参照と公開状態遷移を所有 | DD-006とFR-010の責務を一意化する |
| L4・L5 | L4は版履歴、L5はHTML全体に紐付くタグ・配置を所有 | 公開版切り替えでタグ・配置を巻き戻さない |
| L7・各機能 | 各機能は固有テスト、L7は統合配線・横断E2E・受入証跡を所有 | テストの二重所有を防ぐ |

## タスク台帳

| ID | タスク名 | 計画状態 | 実行状態 | 主成果物・到達状態 | 直接依存 |
| --- | --- | --- | --- | --- | --- |
| L1 | 横断設計・共有契約と最小実装基盤を確立する | 承認済み | 未着手 | 技術構成・信頼境界・共有契約・論理Schema方針・起動可能な最小骨格 | なし |
| L1-M1 | 横断アーキテクチャ・技術構成・信頼境界を設計する | 承認済み | 未着手 | 技術選定ADR、実行トポロジー、信頼境界、依存方向、Module・Directory境界 | なし |
| L1-M1-S1 | 論理実行トポロジー・横断信頼境界・主要データフローを設計する | 承認済み | 未着手 | 技術非依存の論理トポロジー、信頼境界、主要データフロー、生成HTMLの到達可否境界 | なし |
| L1-M1-S2 | 横断技術構成を選定しADRを作成する | 承認済み | 未着手 | Web・Frontend・Backend・永続化・Migration・Build・TestのADR、評価基準、見直し条件 | L1-M1-S1 |
| L1-M1-S3 | Module・Directory境界と依存方向を設計する | 承認済み | 未着手 | 構成図、依存Matrix、論理Path・Glob Owner、共有変更引き渡し規則 | L1-M1-S2 |
| L1-M1-S4 | 設定・秘密情報の分類と露出境界を設計する | 承認済み | 未着手 | 設定・秘密情報台帳、露出可否、秘密値非記録規則 | L1-M1-S2 |
| L1-M2 | 共有ドメインモデル・状態語彙・機能間契約を設計する | 承認済み | 未着手 | 論理モデル、状態Owner、識別子、機能間Interface、論理Schema、Mock例 | なし。L1-M1と並行しG0で整合 |
| L1-M2-S1 | 共有ドメイン語彙・論理モデル・識別子・状態Ownerを定義する | 承認済み | 未着手 | 用語集、論理関連、識別子・参照規則、不変条件、状態・データOwner Matrix | なし |
| L1-M2-S2 | 機能間Interface仕様とMock例を定義する | 承認済み | 未着手 | 技術非依存Interface、提供・利用・エラー・互換性規則、Canonical Mock例 | L1-M2-S1 |
| L1-M2-S3 | 論理SchemaとMigration統治方針を定義する | 承認済み | 未着手 | 論理Schema、契約の永続・一時対応、Migration Owner・順序・互換性・移管規則 | L1-M2-S1。MergeはL1-M2-S2後 |
| L1-M3 | 起動可能な最小Webアプリ・開発テスト基盤を構築する | 承認済み | 未着手 | Project骨格、Manifest・Lockfile、Build・Test設定、Bootstrap、Fixture基盤 | G0 |
| L1-M3-S1 | 最小Webアプリ骨格・共通Toolchain・安全な設定雛形を構築する | 承認済み | 未着手 | Project骨格、Root Manifest・Lockfile、共通Command、最小Bootstrap、設定雛形 | G0 |
| L1-M3-S2 | 基盤Smoke・初期共有Fixture・コード生成再現性を整備する | 承認済み | 未着手 | Smoke、Test harness、技術用Fixture、Codegen規則・差分検出、clean検証 | L1-M3-S1 |
| L1-M4 | 共有契約とSchema・Migration実行基盤を実装する | 承認済み | 未着手 | 共有型・Port、Migration実行基盤、基盤用Migration、契約互換テスト | G0、L1-M3 |
| L1-M4-S1 | 共有識別子・契約型・Portを実装し互換性を検証する | 承認済み | 未着手 | 共有契約Code、Serialization、Canonical Mock互換Test、契約由来生成物 | G0、L1-M3 |
| L1-M4-S2 | Schema・Migration実行検証基盤を実装する | 承認済み | 未着手 | Migration runner、履歴・失敗処理、基盤Migration、Schema Fixture、実行Test | G0、L1-M3。MergeはL1-M4-S1後 |
| L2 | 単一管理者のGoogle認証と管理認可を実現する | 承認済み | 未着手 | Google認証、管理者判定、セッション、管理API・UIの認可 | 設計: G0の認証・信頼境界契約。実装: L1完了Commit＋L2 G1 |
| L2-M1 | Google認証とSessionライフサイクルを実現する | 承認済み | 未着手 | VerifiedGoogleIdentity、Google Flow・Callback、Session発行・更新・失効、Login UI | 設計: G0。実装: L1完了＋L2 G1 |
| L2-M1-S1 | Google認証・Session・M2連携を詳細設計する | 承認済み | 未着手 | Sequence、状態遷移、Callback/Cookie/CSRF対策、UI Owner、Schema要否、Test Matrix | G0 |
| L2-M1-S2 | Google Identity認証Flowを実装する | 承認済み | 未着手 | Login・Callback、Google Adapter、認証Transaction、VerifiedGoogleIdentity、Provider Test | L1完了、L2 G1、L2-M1-S1 |
| L2-M1-S3 | M2許可後の管理者Sessionライフサイクルを実装する | 承認済み | 未着手 | Authorization Port連携、Session・Cookie・CSRF・Logout・期限切れUI、Session Test | L1完了、L2 G1、L2-M1-S1。Merge: S2 |
| L2-M2 | 単一管理者認可と管理面アクセス境界を実現する | 承認済み | 未着手 | 管理者照合、AuthorizationDecision、Backend・Frontend Guard、拒否UI | 設計: G0。実装: L1完了＋L2 G1。Merge: L2-M1 |
| L2-M2-S1 | 認可Matrix・設定・強制境界を詳細設計する | 承認済み | 未着手 | 認証認可Matrix、管理者設定・照合、Guard境界、L6/L7引渡し、Test Matrix | G0、L2-M1-S1との合同G1 |
| L2-M2-S2 | VerifiedGoogleIdentityを単一管理者に照合しAuthorizationDecisionを提供する | 承認済み | 未着手 | Server設定、verified email照合、Policy、AuthorizationDecision・AdministratorPrincipal | L1完了、L2 G1、L2-M2-S1 |
| L2-M2-S3 | 管理API・UI Guardと非管理者拒否体験を実装する | 承認済み | 未着手 | Backend Guard、Frontend補助Guard、拒否UI、保護Route契約、固有Test | L1完了、L2 G1、L2-M2-S1。Merge: L2-M1、S2 |
| L3 | 解説HTMLを生成・形式検証・再試行し生成ログを記録する | 承認済み | 未着手 | Codex app-server生成、HTML形式検証、再試行、要約ログ、生成結果契約 | 設計: G0の生成・版受け渡し契約。実装: L1完了Commit＋L3 G1 |
| L3-M1 | Codex app-serverで単一生成試行を実行する | 承認済み | 未着手 | GenerationRequest写像、Codex Adapter、候補HTML・型付き失敗、Provider Contract Test | 設計: G0。実装: L1完了＋L3 G1 |
| L3-M1-S1 | Codex単一試行契約・接続を詳細設計する | 承認済み | 未着手 | 接続Sequence、Lifecycle、Request/Response写像、Provider失敗、設定、Test Matrix | G0 |
| L3-M1-S2 | Adapter・Provider Stub・Contract Testを実装する | 承認済み | 未着手 | Codex Client、Lifecycle component、Stub、Failure mapping、決定的Contract Test | L1完了、L3 G1、S1 |
| L3-M2 | 生成結果のHTML形式を機械検証する | 承認済み | 未着手 | 決定的Validator、ValidatedGenerationResult、安全な失敗理由、境界Corpus | 設計: G0。実装: L1完了＋L3 G1 |
| L3-M2-S1 | HTML形式判定規則・非改変原則・安全な失敗契約を詳細設計する | 承認済み | 未着手 | 判定表、結果不変条件、安定失敗Code、安全要約、Contract Matrix、Corpus仕様 | G0 |
| L3-M2-S2 | 決定的HTML Validator・境界Corpus・契約Testを実装する | 承認済み | 未着手 | Parser Wrapper、安定変換、機能内Corpus、Table/Contract Test、Canonical Fake | L1完了、L3 G1、S1 |
| L3-M3 | Generation Run・Attemptを統括し記録・管理操作を提供する | 承認済み | 未着手 | Run/Attempt状態、最大4試行、手動Run、履歴・要約、生成UI、L4受渡し | 設計: G0。実装: L1完了＋L3 G1。Merge: M1、M2 |
| L3-M3-S1 | Run・Attempt状態、再試行・冪等性・L4受渡しを詳細設計する | 承認済み | 未着手 | 状態遷移、Retry予算、Command/Query、Schema意味、Safe Log、L4 Handoff、Test Matrix | G0、M1-S1・M2-S1と合同G1 |
| L3-M3-S2 | Run・Attempt状態機械、Repository、生成Orchestratorを実装する | 承認済み | 未着手 | Aggregate、Repository、Job Handler、最大4 Attempt、冪等Command、Fencing、Handoff | L1完了、L3 G1、S1。実統合: M1-S2、M2-S2 |
| L3-M3-S3 | 管理者向け生成・手動再試行・試行ログAPI/UIを実装する | 承認済み | 未着手 | 単一Prompt入力、管理API/UI、各失敗要約・Retry状態、最終失敗、手動新Run | L1完了、L3 G1、S1。実統合: S2、L2認可契約 |
| L4 | 検証済みHTMLを確認・修正し版履歴を管理する | 承認済み | 未着手 | 管理者確認、修正生成の編成、不変版Repository、版履歴 | 着手: G0のHTML・版契約。完了・Merge: L3 |
| L4-M1 | 検証済みHTMLを不変Versionとして保存し履歴正本を提供する | 承認済み | 未着手 | HTML/Version Aggregate、冪等Handoff受入、不変Repository、履歴Query、L6 Read Port | 設計: G0。実装: L1完了＋L4 G1。実統合: L3 |
| L4-M1-S1 | 不変Version・生成結果の冪等受入・履歴／読取契約を詳細設計する | 承認済み | 未着手 | 不変条件、修正元、Handoff方式/Transaction、同時受入、Schema、Metadata/Content Port | G0、L3-M3-S1。L4 G1 |
| L4-M1-S2 | Version Aggregate・Repository・Handoff Consumer・履歴Queryを実装する | 承認済み | 未着手 | 検証済み結果の冪等受入、不変保存、一意採番、履歴、Metadata Query、L6 Content Port、固有Test | L1完了、L4 G1、S1。実統合: L3 |
| L4-M2 | 管理者がHTMLを確認し修正生成を依頼できるようにする | 承認済み | 未着手 | Version選択・確認、隔離Preview導線、自然言語修正、L3 Run開始、管理API/UI | 設計: G0。実装: L1完了＋L4 G1。実統合: M1、L3、L6 Port |
| L4-M2-S1 | Version選択・修正元適格性・修正Command・隔離Preview連携を詳細設計する | 承認済み | 未着手 | 選択状態、最新Version適格性、修正Command/冪等性、L3 Query、Preview Consumer契約、Test Matrix | G0、L4-M1-S1・L3-M3-S1とL4 G1 |
| L4-M2-S2 | 管理者向けVersion確認・修正生成・履歴・Preview導線API/UIを実装する | 承認済み | 未着手 | Metadata/履歴API/UI、隔離Preview導線、単一修正Prompt、L3 Run参照、Backend Guard、固有Test | L1完了、L4 G1、S1。実統合: M1-S2、L3、L2。Preview実統合: L6 |
| L5 | HTML全体のタグとサイト内配置を管理する | 承認済み | 未着手 | タグ作成・変更・付与・解除、任意名の配置先、版非依存の管理状態 | 設計: G0のHTML全体・版識別契約。実装: L1完了Commit＋L5 G1 |
| L5-M1 | HTML全体に紐付くタグ・配置Metadataの正本を提供する | 承認済み | 未着手 | Tag/Placement Domain、Repository、適格性Port、Metadata Query、統合Site Slot向け最小Read Projection | 設計: G0。実装: L1完了＋L5 G1。実統合: L4適格性Port、L7 Site Slot |
| L5-M1-S1 | Tag・Placement Domain、HTML関連Commandの適格性、Port、Schemaを詳細設計する | 承認済み | 未着手 | Cardinality、Command適格性表、名前規則、Transaction、Schema、L4/Site Slot Port、Test Matrix | G0、L4-M1-S1。L5 G1 |
| L5-M1-S2 | Tag・Placement MetadataのDomain・Repository・Application・最小Read Projectionを実装する | 承認済み | 未着手 | Metadata Root、Tag/Placement操作、Repository、L4適格性Adapter、現在Metadata Projection、固有Test | L1完了、L5 G1、S1。実Adapter: L4完了後L7統合 |
| L5-M2 | 管理者がタグ・配置を管理できるAPI/UIを提供する | 承認済み | 未着手 | Tag作成/改名/付与/解除、Placement作成/選択/置換、管理API/UI、Backend Guard | 設計: G0。実装: L1完了＋L5 G1。実統合: M1、L2 |
| L5-M2-S1 | Tag・Placement管理API、Feature-local管理Panel、認可・冪等性・エラー契約を詳細設計する | 承認済み | 未着手 | HtmlId拡張契約、API/Panel、Backend Guard、冪等/競合/Error、Test Matrix | G0、L5-M1-S1。L5 G1 |
| L5-M2-S2 | Tag・Placement管理API、Feature-local管理Panel、固有Testを実装する | 承認済み | 未着手 | Feature API/DTO、管理Panel、Route Descriptor、M1/L2/L4 Fake、API/Component Test | L1完了、L5 G1、S1。実Mount: L7 |
| L6 | 公開版状態と隔離された生成HTML閲覧を実現する | 承認済み | 未着手 | 公開承認・切り替え・非公開化、匿名閲覧、管理者プレビューと公開表示の隔離 | 設計: G0の版・公開・隔離契約。実装: L1完了Commit＋L6 G1。完了・Merge: L2、L4 |
| L6-M1 | 公開Version Pointerの正本と管理者向け公開操作を提供する | 承認済み | 未着手 | Publication Aggregate/Repository、Approve/Switch/Unpublish、最小Public Projection、管理API/Panel | 設計: G0。実装: L1完了＋L6 G1。実統合: L4、L2 |
| L6-M1-S1 | Publication Pointer・状態遷移・L4適格性・冪等性・Projection／管理Slot契約を詳細設計する | 承認済み | 未着手 | 0..1 Pointer、Command/CAS、Schema、Revision/Cache、Public Projection、Panel Slot、Test Matrix | G0、L4/L2契約、M2設計とL6 G1 |
| L6-M1-S2 | Publication Aggregate・Repository・Application・Revision付きPublic Projectionを実装する | 承認済み | 未着手 | Pointer正本、Local Transaction、Dedupe、CAS、Projection/Tombstone、固有Test | L1完了、L6 G1、S1。実統合: L4 |
| L6-M1-S3 | 管理者向け公開API・Feature-local Publication Panel・固有Testを実装する | 承認済み | 未着手 | Approve/Switch/Unpublish API/Panel、競合表示、Route/Action Descriptor、Fake Host | L1完了、L6 G1、S1。実統合: S2、L2、L4 SlotはL7 |
| L6-M2 | 管理Previewと匿名公開をResolver分離・Renderer共通の隔離HTML Surfaceで提供する | 承認済み | 未着手 | Preview/Public Resolver、CredentialなしRenderer、匿名Route、Header/Config契約、Multi-origin Security Test | 設計: G0。実装: L1完了＋L6 G1。実統合: M1、L4、L2 |
| L6-M2-S1 | Preview／Public解決・Capability交換・隔離Surface・Cache／Origin境界を詳細設計する | 承認済み | 未着手 | 別Trust Flow/Context、Capability交換、ResolvedRenderTarget、Origin/Header/Cache/Fail-closed契約、Test Matrix | G0、M1 Projection、L4 Content Port、L2 Guard/Cookie契約。L6 G1 |
| L6-M2-S2 | Capabilityを生成HTMLへ露出せず消費するPreview Resolverと現在Pointerだけを解決するPublic Resolverを実装する | 承認済み | 未着手 | Capability発行・原子的消費、Exact Version Preview、毎回Pointer解決するPublic Resolver、Generic Failure、固有Test | L1完了、L6 G1、S1。実統合: M1/L4/L2 |
| L6-M2-S3 | ResolvedRenderTargetだけを受け取るCredentialなしRenderer・隔離Surface・Version本文Cache・Multi-origin Security Harnessを実装する | 承認済み | 未着手 | 共通Renderer、隔離Response/Header、Version本文Cache、Fail-closed Feature設定、Multi-origin Harness | L1完了、L6 G1、S1、L4 Content Port。S2 Contract Seed後に並行可 |
| L7 | 初期リリースを統合し受入テストを自動化する | 承認済み | 未着手 | Release Baseline、統合済みRuntime、全29件の自動受入・G2判定用証跡 | 着手: L1完了Commit。完了: M1〜M3完了、全必須Profile合格、重大な未統合なし |
| L7-M1 | 共有変更を逐次統合し再現可能なRelease Baselineを確定する | 承認済み | 未着手 | 共有変更lane、Manifest/Lockfile、Migration台帳/共有File、共有契約物理表現、Codegen台帳、最終Baseline | L1完了・Owner移管。各機能G1後の変更要求を逐次受領 |
| L7-M1-S1 | 共有変更laneの統治契約を確定しL1成果を引き受ける | 承認済み | 未着手 | Path/Glob引渡し、変更要求/承認Matrix、Baseline状態、台帳初期化、Dry-run、Writer移管 | L1完了・Owner引渡し可能 |
| L7-M1-S2 | 共有変更を逐次統合しRelease Baselineを確定する | 承認済み | 未着手 | 単一Writer lane、Stable Commit、共有物理統合、Clean検証、Final Baseline/失効記録 | S1。個別変更は該当G1、最終完了はL2〜L6 Release対象Merge |
| L7-M2 | 全機能Runtimeを構成し実行環境・隔離境界を成立させる | 承認済み | 未着手 | Composition/Router/DI、実Adapter/Slot/Worker、Codex lifecycle、L6実Origin/Proxy/Cache/Fail-closed | L1後に骨格着手。完了: L2〜L6実装、M1最終Baseline |
| L7-M2-S1 | 全Feature Descriptor・Port・SlotをRuntime Compositionへ統合する | 承認済み | 未着手 | Composition/Router/DI/Registry、必要なWorker、Glue、管理/匿名/Render Route Group、L4/L5/L6 Slot、Wiring Smoke | L1、M1-S1後に骨格。最終Mount: L2〜L6、S2/S3。後にM1 Final |
| L7-M2-S2 | Codex app-serverの実Process lifecycle・設定・Smoke実行物を統合する | 承認済み | 未着手 | Executable/Version検証、Process/Transport Host、L3 Adapter配線、Protocol Double Smoke、実接続Command | 骨格: L1完了・L3 G1 Provider契約。実配線: L3-M1-S2。最終Smoke: M1 Final |
| L7-M2-S3 | L6隔離Surfaceの実Host・Origin・Network・Cache Infrastructureを統合する | 承認済み | 未着手 | 実Origin/Port、CI TLS/Proxy/Header反映、Cache Store/失効Adapter、Fail-closed、Multi-origin topology | 骨格: L1完了・L6 G1。実配線: L6実装・L2 Cookie/Guard・L4 Content。最終証明: M1 Final/M2 Runtime |
| L7-M3 | AC-001〜AC-029の自動受入とG2判定用証跡を作成する | 承認済み | 未着手 | Trace Matrix、統合Fixture、横断E2E、Actual-origin Security E2E、Release Check・証跡 | L1後に骨格着手。全実行: M2、L2〜L6。完了後G2判定 |
| L7-M3-S1 | AC Trace Matrix・Test Architecture・Integration Fixture Harnessを確立する | 承認済み | 未着手 | Acceptance Manifest/Evidence Schema、Harness、決定Clock/ID、DB Reset、Worker Probe、IdP/Protocol Scenario入力 | L1後に骨格。確定: L2/L3 G1 Contract Seed |
| L7-M3-S2 | 生成・検証・再試行・確認修正の横断E2Eを自動化する | 承認済み | 未着手 | AC-001〜008・016〜018・025〜027の決定的Browser/API/Worker Test Pack | S1、L3/L4/L6 Preview。最終実行: M1 Final/M2 Runtime |
| L7-M3-S3 | タグ・配置・Version非依存性の横断E2Eを自動化する | 承認済み | 未着手 | AC-009〜013・028、特にAC-010 Placement Site Slot維持の決定的Test Pack | S1、L5/L4/L6/Site Slot。最終実行: M1 Final/M2 Runtime |
| L7-M3-S4 | Google認証・管理認可・公開状態の横断E2Eを自動化する | 承認済み | 未着手 | AC-014・015・020〜024のGoogle互換IdP、Backend Guard、Publication Test Pack | S1、L2/L4/L6。最終実行: M1 Final/M2 Runtime |
| L7-M3-S5 | Actual-origin隔離・実Codex Smoke・Release証跡を確定する | 承認済み | 未着手 | AC-019/029、外部HTTPS/Admin Canary、実Codex Smoke実行、Release Check、不変Evidence Package | S1、S2〜S4結果、M1 Final、M2 Runtime/Topology/Smoke実行物 |

## Gate前設計成果物のPath・単一Owner

技術スタックとDirectory構成が未確定であるため、具体的なPath・GlobはL1の中分類分割とG0で確定する。現時点では次の論理Ownerを固定する。

| 対象 | 単一Owner | Merge前提 |
| --- | --- | --- |
| 技術選定ADR、実行トポロジー、信頼境界、Module・Directory境界 | L1-M1 | なし |
| 共有ドメイン契約、論理Schema・Migration方針、共有Interface | L1-M2 | G0 |
| 初期Manifest・Lockfile、コード生成規則、初期共有Fixture | L1-M3 | G0 |
| 共有型・Port、Migration実行基盤、基盤用Migration、Schema Fixture | L1-M4 | G0、L1-M3 |
| Google認証・管理認可の機能固有設計 | L2 | G0の認証・信頼境界契約 |
| 生成・検証・再試行・生成ログの機能固有設計 | L3 | G0の生成・版受け渡し契約 |
| 不変版Repository・版履歴の機能固有設計 | L4 | G0のHTML・版契約 |
| タグ・配置の機能固有設計 | L5 | G0のHTML全体・版識別契約 |
| 公開状態・隔離閲覧の機能固有設計 | L6 | G0の版・公開・隔離契約 |
| composition root、全体Router、DI・Registry、統合Fixture、E2E | L7 | L1完了Commit |

## L7内部（承認済み）

```text
L1完了
 ├─ M1共有変更laneを開く
 ├─ M2統合骨格を開始
 └─ M3 Trace Matrix・Fixture骨格を開始

各機能ごと:
 M1共有Commit → 機能worktreeをrebase・Merge → M2逐次配線 → M3対応E2E追加

L3 → L4、L2 + L4 → L6、L5は独立統合
 → M1最終Release Baseline → M2最終Runtime適合
 → M3全29件合格・証跡確定 → L7完了 → G2判定 → 初期リリース完了
```

- M1は長寿命の単一共有変更laneであると同時に、未反映要求なし、Migration/Manifest/Lockfile/共有契約物理表現の整合、Codegen再生成差分なし、Clean CheckoutからのBuild/Test/Migration初回・順序・再適用成功、最終Baseline Commit記録を完了条件とする。後続共有変更時は再度開き、M2/M3の古い証跡を無効化する。
- 共有契約の業務意味はL1・該当機能、機能固有生成物は各機能、M1は物理統合・生成物台帳・最終再現性を所有する。Migration Runnerを再実装しない。
- M2は各機能Descriptor/PortをMountし、機能ロジックを複製しない。Worker/Schedulerは採用Architectureで必要なものだけを配線する。
- Codexの通常Blocking TestはProtocol-faithful Test Doubleを実Adapter/Lifecycle経由で実行する。G2用実接続Smokeは管理環境・対応Version・前提条件を固定し、Handshake/起動停止/Request-Response/HTML候補取得を確認する。環境不足のSkipはG2合格にしない。
- M2のTLS/Proxy/Network PolicyはReference/CI実行トポロジーまでとし、本番Hosting、DNS、証明書発行、監視を含めない。Server-side Proxyを採用した場合だけPrivate/Internal Address拒否を実配線する。
- AC-010は任意名Placementを設定し、L5 Projectionを利用する統合Site Slotへの反映、公開Version切替後の同一HtmlIdのPlacement維持までをE2Eで証明する。Navigation階層・並び順・検索は追加しない。
- M3はAC-004/025の失敗表示後自動Retry、AC-016/027のValidation/Provider失敗別の手動新Run、AC-019のPreview/Public両Surface、AC-029の外部HTTPS成功と管理API読取・変更失敗を明示的に覆う。
- M3完了時、AC-001〜AC-029にSkip/Quarantineがなく、各ACをTest ID・Fixture・環境・結果へ追跡可能にする。M2はSmoke実行物、M3はRelease環境での実行と証跡を所有する。

### L7-M1内部（承認済み）

```text
L1 Owner引渡し可能 → M1-S1統治契約・引受・台帳初期化 → M1-S2 Open

各Feature:
 Feature G1で意味承認 → 共有変更要求 → S2 Stable Commit
   → Feature rebase・Merge → M2逐次配線 → M3逐次E2E

全Release対象Feature Merge → S2 Queue空・Clean検証 → Final Baseline
 → M2最終Runtime → M3全受入・証跡

後続共有変更/不整合:
 Final → Invalidated → Open → 新Final → M2/M3再検証
```

- S1はL1から引き継ぐPath/Glob・Commit・Owner・Migration予約・Codegen登録を棚卸しし、二重Ownerなしの受領記録を作る。変更要求Schema、承認Matrix、`Open / Stable / Final / Invalidated`、Baseline ID/失効規則、rebase/Merge手順を定義する。
- S1はMigration予約台帳・Codegen台帳・共有変更Queueを初期化し、代表的な架空要求のDry-runでOwner・Gate・状態遷移を検証する。実共有変更や業務意味の採否は行わない。
- S1は台帳形式と初期状態を所有し、完了時に中央Queue・Migration予約台帳・Baseline記録の単一WriterをS2へ移管する。Featureは要求を提出できるが中央Index/番号/状態を直接編集しない。
- S2は最初の要求からRelease Cutまで進行中を維持し、途中のStable CommitをTask完了扱いにしない。`Final`は全Release対象Merge、Queue/承認待ち0、最終Clean検証、Baseline Commit/Input Hash/Tool Version/結果記録でのみ成立する。
- Lockfile更新・承認済み依存・決定済みCodegenはS2が統合できる。互換な共有型/Schema追加は提供OwnerとConsumer承認、意味変更/削除/状態Owner変更は機能G1再確認、G0信頼境界等の変更はG0再確認を要求する。S2が独自に意味を変更しない。
- S2はRoot Codegen設定・共通入口・共有契約由来成果物・生成物台帳を所有する。Feature-local入力/生成物と業務的正しさは各機能が所有し、S2はRelease Cut時に全登録生成物の再現性だけを検証する。
- S2の一次成果はBaseline Manifest、Build/Lockfile/Codegen/Migration検証結果である。M3は再実装せずBaseline IDと成果をG2証跡から参照する。

### L7-M2内部（承認済み）

```text
L1完了 + M1-S1
 ├─ M2-S1 Composition骨格（Mock/Descriptor Seed）
 ├─ M2-S2 Codex Process Host骨格
 └─ M2-S3 Isolation Infrastructure骨格

必要なRoot依存: M1-S2 Stable Commit → S1/S2/S3 rebase
L3 Merge → S2実Transport配線
L6 Merge → S3実Port/Host/Cache backend配線

L2〜L6 Merge + S2/S3 Merge → S1最終Mount・Composition Merge
 → M1-S2 Final Baseline → Baseline上でM2全Runtime Smoke → M2完了 → M3全受入
```

- S1はComposition root、Global Routerと管理/匿名/Render Route Group、DI/Registry、必要なWorker Host登録、Feature Descriptor Mount、L4 Page Slot、Placement Site Slot、単純Glue、Readiness/Wiring Smokeを単一Ownerとして持つ。
- Feature固有Protocol/Repository Adapterは各機能、Codex Process/Transport HostはS2、Host/Proxy/Cache Infrastructure AdapterはS3、変換なしのPort接続Glueと全AdapterのDI登録はS1が所有する。業務Policyを含む変換をS1で発明しない。
- S1のRoot Config責務は型付きConfig Descriptorの集約/登録に限定し、Root Manifest/Lockfile・共有Root Config Indexの物理変更はM1-S2へ要求する。
- S1はL5 Placement Projectionを必須の統合Site SlotへMountし、設定名と対象HTMLの関連をSite Shellで確認可能にする。階層Navigation、並び順、検索、複数配置を追加しない。
- S2はExecutable発見/Version検証、Process起動停止/異常終了、Transport Channel、L3 AdapterへのClient注入を所有し、Prompt写像、Candidate抽出、Typed Failure、Retry/Validationを再実装しない。Protocol DoubleはL3契約Fixtureを参照する。
- S2の実接続Commandは前提設定、Executable/Version、Handshake、Request/Response、HTML候補、正常終了だけを確認し、生成内容の完全一致・学習効果・時間目標を判定しない。インストールは対象外で、未導入/不一致を前提条件Errorにする。
- S3は実Host/Port/Origin、Reference/CI TLS/Proxy、L6 Header/Origin Policy反映、Concrete Cache Storeと失効Port/Event Adapter、Multi-origin topology、危険設定のFail-closedを所有する。
- Version本文Cache key/可否、Handle対応非Cache、Preview/Negative非共有Cache、Publication Revision/Tombstone/失効意味はL6に残す。S3はConcrete backendへ配線し、外部Cache製品を必須化しない。
- Server-side Proxy採用時だけPrivate/Internal IP、DNS Rebinding、Redirect再解決を拒否する。本番Hosting/DNS/証明書発行/CDN/監視は対象外とする。
- S3は実設定Response/Header/Origin分離/Cache Adapter Contract Smokeまでを完了し、BrowserでのCredential/CORS/外部HTTPS/旧版防止はM3のActual-origin E2Eで証明する。

### L7-M3内部（承認済み）

```text
L1完了 → M3-S1 Harness骨格
L2/L3 G1 Contract Seed → S1 Identity/Protocol Contract確定 → S1完了

S1完了後
 ├─→ S2 生成・検証・Retry・確認修正Pack
 ├─→ S3 Tag・Placement・Version横断Pack
 ├─→ S4 認証・認可・公開状態Pack
 └─→ S5 Actual-origin/Evidence骨格

M1 Final Baseline → M2最終Runtime Smoke
 ├─→ S2/S3/S4最終実行
 └─→ S5 Security・実Codex Smoke実行
      → S5 Evidence集約 → M3/L7完了 → G2判定
```

- S1はVersion管理されたAcceptance ManifestをSource of Truthとし、`AC ID → Primary Scenario ID → Profile → Fixture → 必須前提`を定義する。各ACのPrimaryはちょうど1件、Supportingは複数可とし、Trace MatrixはManifest/Reportから生成して手編集しない。
- S1は決定Clock/ID、DB Reset、状態Query/Job Probe/Eventによる固定SleepなしのWorker待機、Feature Fixture Builder合成、Google互換IdP Harness、Codex Protocol DoubleのScenario入力、Required TestのSkip/Quarantine検出を所有する。Double実行物はM2-S2、Feature Fixture意味は各機能に残す。
- S2はAC-001〜008、016〜018、025〜027を所有する。1 Runは初回Attempt 1回＋自動Retry最大3回、手動Retryは新Runで予算Resetとし、Validation/Provider失敗それぞれで各失敗表示後の次Attempt開始を観測する。
- S2のAC-002/017はL6 Preview Resolver/Renderer経由の実HTML内容を確認し、Raw HTMLを管理DOMへ直接挿入しない。AC-001は視覚要素付き決定Fixtureを使い、主観的学習効果を判定しない。
- S3はAC-009〜013、028を所有し、特にAC-010についてPlacement作成・設定、統合Site Slot反映、公開Version切替後の維持までを確認する。
- S4はAC-014、015、020〜024を所有する。Google互換IdPから実OAuth Client/Callback/State/Nonce/管理者照合/Session経路を通し、Verified Identityを直接注入しない。AC-015では匿名閲覧に加えGeneration/Tag/Publication管理API Canary拒否、AC-023では未登録/未検証emailのBackend拒否を確認する。
- S5はAC-019をPreview/Public双方、AC-029を管理対象外部HTTPS成功と管理API応答読取/状態変更失敗の同一Scenarioで確認し、外部Server受信Header/Cookieにもアプリ資格情報がないことを確認する。
- S5はM2-S2の実Codex Smoke Commandを固定Version/前提の管理Release環境で必須実行する。環境不足・Version不一致・SkipはRelease Check失敗とし、G2判断自体は行わない。
- Evidence Packageは内容Hash、Baseline ID、Source Commit、Runtime Config Digest、Tool/Adapter Version、Provider Mode、全Scenario結果、Skip/Quarantine、M1/M2一次Artifact参照、Sanitized Log参照を持つ不変Artifactとする。
- M1 FinalがInvalidatedとなった場合、そのBaseline IDのM2/M3証跡を無効化し、新BaselineでM2 Smoke・S2〜S5を再実行する。旧Evidenceを新Packageへ部分コピーしない。
- S1はAcceptance Manifest Schema、生成Tool、中央Index出力Pathの単一Ownerである。S2〜S5は別Directory・Manifest Fragment・Fixture Namespaceを所有し、S5はS1 Toolを実行して生成Index/ReportをEvidence Packageへ収録するが手編集しない。本番Sourceへ直接Test Hookを追加せず、必要時はFeature/M2 Ownerへ要求する。
- Browser E2Eの実行環境は受入証跡取得用の固定Test Profileであり、生成HTMLについて特定ブラウザ・端末の表示保証（NFR-005）を新設しない。

## 依存関係

### L6-M1内部（承認済み）

```text
G0 → M1-S1 ─ L4 Version/L2 Guard/M2 Projection Revision整合 → L6合同G1

L1完了 + L6 G1
 ├─→ M1-S2
 └─→ M1-S3（S2/L2/L4 Fake）

S2 Merge → S3 rebase・Merge → L4/L2 Adapter → M2統合 → L7共有統合
```

- Set PointerはL4 PortでVersion存在・HtmlId所属・検証済みを確認した後、L6 PublicationだけのLocal Transactionで`expected_revision`比較、Pointer更新、Command Dedupeを保存する。L4を分散Transactionへ参加させない。
- L4 Versionは不変・削除なしを前提とし、適格性をL6へ複製しない。UnpublishはL4確認不要でPublicationだけをCAS更新する。
- `CommandId`は同一Payload再送を同じ結果へ収束させ、異なるPayload再利用を拒否する。`expected_revision`は別Tab等の古い状態を検出する。
- Approve/SwitchはDomain上同じPointer設定、UnpublishはClearとする。同じPointer再設定は成功No-opとしてRevisionを増やさず、公開履歴やVersion別Approved Flagは持たない。
- S2はPointerとPublication Revisionの原子的更新、Public Projection/Tombstone、必要時の失効Event意味を所有する。M2は本文Cache/Header/Revision再検証、L7はCache/Event実配線を所有する。
- 公開操作は任意の検証済み履歴Versionを対象にできるが、L4修正元は最新Versionだけである。公開TargetとCorrection Sourceを型・文言・Commandで分離する。
- S3は`HtmlId + selectedVersionId`のFeature-local Panel/Slot DescriptorとFake Hostを所有し、L4 Pageを直接編集しない。実MountはL7が所有する。
- S1はPublication/Revision/Command DedupeのSchema意味、S2は物理化要求、L7はMigration番号・共有File・Root DIを所有する。

### L6内部（承認済み）

```text
G0
 ├─→ M1詳細設計
 └─→ M2詳細設計 ─ L2 Guard・L4 Content Port整合

M1/M2設計整合 → L6合同G1

L1完了 + L6 G1
 ├─→ M1実装（L4/L2 Fake）
 └─→ M2実装（M1/L4/L2 Fake、Multi-origin Harness）

M1 Merge → M2 rebase・Merge → L2/L4 Adapter統合
       → L7 Host/Router/DI/E2E統合
```

- M1はHTMLごとに公開Version Pointerなし/1件だけを正本とし、Approve/SwitchはPointer設定、UnpublishはPointer除去、新Version生成では変更しない。
- 対象Versionの存在・HtmlId所属・検証済みをL4 Portで確認しClient申告を信用しない。Public Projectionは現在公開中のHtmlId/VersionRefだけを返しRaw HTMLを含めない。
- M1はFeature-local公開操作Panel/Action Descriptorを提供し、L4履歴Page本体を直接編集しない。実Slot MountはL7が所有する。
- 管理Previewは外側でL2 GuardとExact Version確認後、短命・対象固定・用途限定Capabilityを発行する。隔離OriginのPreview Resolverが生成HTML実行前にこれを原子的に消費し、Tokenなしの内部`ResolvedRenderTarget`だけをRendererへ渡す。最終URL、DOM、Referer、JS可視Cookie、生成HTML ContextへCapability、Principal、Session/Cookie、Google Token、CSRF、管理URL/データを残さない。
- 匿名RouteはPublic Handleだけを受け、M1 Projectionから現在Pointerを解決する。任意Version/過去公開版/Preview URLによる迂回を許さず、非存在・未公開・非公開・無効Handleを同じ外形へ収束させる。
- 共通Rendererは`ResolvedRenderTarget`だけを受け取り、資格情報・Capability・Public Handleを受け取らない。Preview/PublicのResolver・Request Contextは分離する。Public HandleからPointerを毎回解決した後にVersion本文Cacheを参照し、切替・非公開時に旧版へ迂回させない。
- 管理面と生成HTMLをCookie非共有HostまたはOpaque Originへ分け、Host-only Cookie、CORS拒否、CSRF/Origin/Fetch Metadata、Storage/Opener/Service Worker分離をG1で確定する。
- 外部HTTPS通信を一律禁止しない。Cookie非送信、管理APIの認証・CSRF・Origin/Fetch Metadata強制、CORS拒否、GET状態変更禁止により管理データ読取・状態変更を防ぎ、Server-side Proxyを使う場合だけ管理Host・Loopback・Private Address・DNS RebindingをNetwork層で遮断する。外部通信成功と管理API到達失敗を同じMulti-origin Testで証明する。
- M1-S2はPointer/Revision、M2-S2は毎回のPointer解決・Capability状態、M2-S3はVersion本文Cache/Header、L7は実Cache backend・失効Consumerを所有する。匿名失敗は同じ404相当本文と`no-store`へ収束し、成功URLもPointer再検証を阻害するImmutable Cacheにしない。
- M2は安全条件・Header/Config Schema・Fail-closed検証・Route Descriptor・Harness、L7は実Host/Port/TLS/Proxy値・Network Policyと全体配線を所有する。
- L5はL6の直接依存にしない。Placementのサイト内反映に必要な最小ProjectionはL7-M2-S1のSite Slotが直接消費し、L7-M3-S3がL6公開切替との横断整合を証明する。Navigation・並び順・Tag検索は追加しない。

### L6-M2内部（承認済み）

```text
G0 → M2-S1 ─ M1 Projection・L2 Guard/Cookie・L4 Content Port整合 → L6合同G1

L1完了 + L6 G1 → S3所有のResolvedRenderTarget Contract Seedを先行Merge
 ├─→ M2-S2（M1/L4/L2 Fake）
 └─→ M2-S3（S2/L4 Fake）

M1-S2 Merge → M2-S2/S3 rebase・実統合 → L2/L4 Adapter統合
            → L7 Host/Proxy/Router/DI/Cache/E2E統合
```

- S1はCapabilityの発行・交換・消費・失効Sequence、論理Config、Header/Destination Policy、Fail-closed条件とThreat/Test Matrixを所有する。
- S2はPreview CapabilityのTTL・原子的単回消費、Exact Version Resolver、Public Handleからの毎回Projection解決、Generic Resolution Failureを所有し、HTML描画・Header・本文Cacheを含めない。
- S3はRendererが受け入れてよい情報を定義する`ResolvedRenderTarget` Code Ownerとなり、その最小Contract Seedを先行Mergeする。S2/S3の両worktreeで同じ型Fileを編集しない。
- S3の本文Cache keyは`VersionId + Content Digest + Renderer Policy Version`とし、Handle→Version対応を保存しない。Preview認可結果・Capability Response・匿名Negative Responseは共有Cacheしない。
- CapabilityはPreview Resolver専用Audience・Exact Version・用途・短いTTLへ固定し、同一Tokenの別対象利用と消費後再利用を拒否する。失敗詳細を生成HTML側へ返す認可外Channelを作らない。
- S2はResolver/Capability/Route Descriptor、S3はRender Contract/Renderer/Header/Cache/Surface Config/Harness、L7はGlobal Route/実Host・Proxy/Root Config/Cache backendを所有する。

### L5-M2内部（承認済み）

```text
G0
 ├─→ M1-S1
 └─→ M2-S1草案

M1-S1 → M2-S1再整合 → L5合同G1

L1完了 + L5 G1
 ├─→ M1-S2
 └─→ M2-S2（M1/L2/L4 Slot Fake）

M1-S2 Merge → M2-S2 rebase・Merge → L7でGuard/Slot/Router/DI統合
```

- S1/S2は`HtmlId`単位のFeature-local管理Panel・API・Mount Descriptorを所有し、`VersionId`やRaw HTMLを入力にしない。L4 Page本体とGlobal Routerは直接編集しない。
- 強い`GenerationAttemptId`と`HtmlId`を混同せず、Backendは不正/未知HtmlId、L4 Portが不適格を返すHtmlId、Client申告の検証済みFlagを安全に拒否する。
- M1の名前Policyを正本とし、作成Commandの冪等Key、rename競合、Tag付与/解除の再送、Placement置換TransactionをAPI/UIで一致させる。
- 全管理Command/QueryはBackend Guardを先に通し、未認証・非管理者・期限切れ時にM1を呼ばない。Frontend Guardは表示補助とする。
- S2はFake Slot Hostで独立完了でき、実L2 Guard・L4 Slot Mount・全体配線はL7が所有する。
- 匿名公開画面、公開状態との合成、公開Navigationの有無・形状は本Taskで扱わず、L6へ未確定Navigation Taskを自動追加しない。

### L5-M1内部（承認済み）

```text
G0 → M1-S1 → L5合同G1

L1完了 + L5 G1
 ├─→ M1-S2（L4適格性Fake）
 └─→ M2実装（M1 Fake）

M1-S2 Merge → M2 rebase・Merge → L4完了後にL7で適格性Adapter統合
```

- S1はCommand別適格性表、Tag多対多、Placement `0 → 1 → 別1へ置換`、Version非依存、名前制約、Concurrency、L4適格性PortとL7 Site Slot向けRead PortをG1へ提出する。
- `HtmlMetadata` RootはL4適格性確認後の最初のHTML関連成立時に作成し、最後のTag解除後も削除しない。公開/Version切替で不適格へ戻らない。
- Tag作成/改名・Placement作成はHTML非依存、Tag解除は既存関連解消のためL4確認不要。Tag付与・Placement初回設定はRoot未作成ならL4確認し、以後は不変条件を利用できる。
- Placement解除は要件に含めず、未設定から初回設定、以後の置換だけを扱う。
- 名前は安定IDを正本とし、空/空白・Trim・重複・大小文字/Unicode・合理的最大長をBackend/UI/Schemaで一致させる。文字種を過度に制限しない。
- 統合Site Slot向けRead Projectionは`HtmlId`から現在Tag/Placement、必要時だけ`PlacementId`から関連HtmlIdを返し、公開絞込・Navigation・並び順を持たない。
- L4 Tableへの物理FKは設けず共有`HtmlId`と適格性Portで論理整合する並行方針とし、実Adapterと全体DIはL7が統合する。

### L5内部（承認済み）

```text
G0 → M1詳細設計 → M2詳細設計の再整合 → L5合同G1

L1完了 + L5 G1
 ├─→ M1実装（L4適格性Fake）
 └─→ M2実装（M1/L2 Fake）

M1 Merge → M2 rebase・Merge → L5完了 → L7 Site Slot/横断E2E統合
```

- TagとPlacementは同じ`HtmlId`へのVersion非依存MetadataとしてM1に結合し、M2は同じ管理画面のAPI/UIを所有する。
- TagはHTMLと多対多で、改名してもTag ID・関連を維持する。削除・統合・階層化は追加しない。
- Placementの最小操作は任意名での作成または初回設定、既存選択、現在設定の置換とする。改名・削除・階層・並び順・複数配置は追加しない。
- Cardinalityは`HTML 0..1 → Placement`、`Placement 0..N ← HTML`とし、1 HTMLの複数配置は要件拡張として再承認する。
- 適格性は新しいHTML関連を成立させるTag初回付与・Placement初回設定時にL4の「検証済みVersionを持つHTML aggregate」Portで確認する。Tag作成/改名、Placement作成、Tag解除には要求せず、既存Metadata Rootが以後の適格性不変条件を証明する。
- 名前の空白・重複・大小文字・Unicode正規化はG1で最低限を決めるが、利用者入力を大きく制限する規則は黙って追加しない。
- Site Slot向けRead Projectionは公開状態を含まない現在Metadataだけとする。L7がL6公開状態と横断統合し、公開可否・匿名RouteはL6、Navigation UIは要件外に残す。
- M1はSchema意味・物理化要求、M2はFeature Route Descriptorを所有し、L7はMigration番号・共有File・Root依存・全体Router/DIを所有する。

### L4-M2内部（承認済み）

```text
G0
 ├─→ L4-M1-S1 ─┐
 └─→ L4-M2-S1 ─┴→ L4 G1

L1完了 + L4 G1
 ├─→ M1-S2（L3 Handoff Fake）
 └─→ M2-S2（M1/L3/L6 Port Fake）

L3完了 → M1-S2実統合・Merge → M2-S2実統合・Merge
       → L4機能内完了 → L6 Preview実統合 → L7 E2E
```

- `selected_version_id`はURL/画面状態で明示する任意履歴Versionで、DBへCurrent Pointerとして保存せず公開中も意味しない。表示後に新Versionが増えても勝手に切り替えない。
- `correction_source_version_id`は最小範囲として最新作成Versionだけを許可する。過去版閲覧中は修正Formを無効化するか最新Versionへの導線を示し、送信時にもBackendで適格性を再確認する。
- 修正指示は単一自然言語入力で空白だけを拒否する。意図的送信ごとのCommand IDをL3冪等Keyへ伝播し、API再送は同じRunRefへ収束させる。
- Backendは管理者Principal、Versionの存在・所属・修正適格性を強制検証する。Frontend Guardだけを認可境界にしない。
- L4はL3 Run QueryのRead Projectionまたはログ画面導線を使い、Retry回数・Attempt・失敗理由を保存せず、独自の手動再試行も作らない。
- L4はPreview対象VersionRefとConsumer契約だけを定義し、L6の具体Route・Sandbox・CSPを先決めしない。Raw HTMLを管理API/DOMへ返さず、Preview失敗時も直接挿入へFallbackしない。
- L4はPreview StubとのContract Testで機能内完了できる。実HTML確認のFR-002/AC-017と隔離AC-019はL6統合、入力から修正までのE2EはL7/G2で完了する。

### L4-M1内部（承認済み）

```text
G0
 ├─→ L4-M1-S1 ────────┐
 └─→ L4-M2詳細設計 ───┴→ L4 G1
       ↑
L3-M3-S1 Handoff設計と整合

L1完了 + L4 G1
 ├─→ M1-S2（L3 Handoff Fake）
 └─→ M2実装（M1/L3/L6 Port Fake）

L3完了 → M1-S2実Handoff統合・Merge → M2実統合
```

- S1完了時にPush/OutboxまたはPullの一方式へ決定し、配送Retry Ownerを一つに固定する。L3は結果/Handoff ID/再取得、M1はConsumer/重複排除/Version Transactionを所有する。
- 初回Versionは修正元を持たず、修正Versionは実際の`source_version_id`を保持する。不在または別HTMLの参照は統合Errorとし、Generation Retry予算を消費しない。
- 「現在確認中」はM1へ保存せずM2のURL/画面状態とする。M1は明示Version IDの読取と履歴だけを提供し、公開中VersionはL6が所有する。
- Handoff受入とVersion作成を同一Transactionにし、Handoff/ValidatedResult IDとHTML単位Version番号を一意にする。同時受入はLock/楽観Concurrency/一意制約のいずれかで直列化する。
- Version番号は作成TransactionのCommit順とし、配送到着順を生成開始順とみなさない。修正元参照は別に保持する。
- Repository失敗・Ack消失は同一Handoffを再処理し、L3 Provider/Validation Retryを開始しない。Commit後は既存VersionRefを返す。
- 管理画面用`VersionMetadataQuery`は本文を返さず、L6 Server-side隔離Surface専用`RenderableVersionContentPort`だけが指定Version本文を返す。
- S2はSchema意味・物理化要求をL7へ渡し、Migration番号・共有Migration File・Worker Host・Root DIはL7が所有する。

### L4内部（承認済み）

```text
G0
 ├─→ L4-M1詳細設計 ─┐
 └─→ L4-M2詳細設計 ─┴→ L4 G1

L1完了 + L4 G1
 ├─→ M1実装（L3 Handoff Mock）
 └─→ M2実装（Version/L3/L6 Port Mock）

L3完了 → M1実統合・Merge → M2実統合・Merge → L4完了 → L6実統合
```

- M1はHTML全体・不変Version・Handoff消費・履歴の正本、M2は管理者確認・修正Request編成・L3 Run開始を所有する。
- L3は検証済み結果・Handoff ID・Run/Attempt参照と再受渡し境界、M1は重複排除とVersion作成Transactionを所有する。配送方式はL3/L4合同G1で一意に選ぶ。
- 保存失敗後は同じ結果を冪等再処理し、新しいGenerationRequestやCodex呼出しを開始しない。
- 履歴上の任意Versionは`selected_version_id`として閲覧できるが、修正元は最小範囲として最新作成Versionだけを許可する。過去Versionからの分岐生成は要件拡張として再承認する。
- L4は`published`・`approved`・公開Version Pointerを持たずL6へ、タグ・配置をVersion Snapshotへ含めずL5へ残す。
- L4は確認画面ShellとPreview対象参照・導線を所有し、HTML描画・Sandbox・CSP・Origin・外部通信隔離はL6 Preview Surfaceへ委譲する。管理画面DOMへRaw HTMLを直接挿入しない。
- M1がVersion Schema意味、Repository、Feature-local Fixture、物理化要求を所有する。L7はMigration番号・共有File・Root依存・全体Router/DIを所有する。

### L3-M3内部（承認済み）

```text
G0 → M3-S1 ───────────┐
G0 → M1・M2詳細設計 ──┴→ L3合同G1

L1完了 + L3 G1
  ├─→ M1実装
  ├─→ M2実装
  ├─→ M3-S2（M1/M2 Fake）
  └─→ M3-S3（Command/Query Port Fake）

M1・M2 Merge → S2実統合・Merge → S3実統合・Merge → L3完了
```

- S1は共有`GenerationRequest`・`ValidatedGenerationResult`を再定義せず、M3内のRun/Attempt状態、Command/Query、永続化対応、L4 Handoffを所有する。
- 自動処理はAttempt 1〜4、手動再試行は旧Runを変えず新Runを作る。Repository障害はProvider/Validation失敗と区別し、Codex再試行予算を消費しない。
- 初回生成入力は単一の自然言語Promptとし、topic欄と追加指示欄へ固定分解しない。空白だけを拒否し、topicを表す入力を案内するが内容分類は追加しない。
- 各失敗確定時に安全な要約と「自動再試行中」を参照可能にし、UI確認を待たず次Attemptへ進む。最終失敗時だけ手動再試行を許す。
- 利用者向け取消は追加せず、System shutdown・接続断・Worker crash時の整合だけを設計する。
- Start/Manual Retryの冪等Key、`(run_id, attempt_sequence)`一意性、Attempt先行永続化、Lease/FencingまたはCASを使う。Provider側保証がない外部呼出しのExactly-onceは約束しない。
- 呼出し後・結果保存前のCrashは同一Attemptを盲目的に再実行せず`outcome_unknown`等へ確定し、必要なら別Attemptとする。L4受渡しは同じ結果を再取得・冪等再受渡しできる方式をG1で選ぶ。
- S2はRun/Attempt/必要時Handoff・Jobの意味、Repository、Feature Fixture、Job Handler、Retry/Fencingを所有する。L7はMigration番号・共有File・Worker Host・Scheduler起動・Root DIを所有する。
- S3はFeature-local Route Descriptorを持ち、Backend Guardを強制境界とする。全体Router mountはL7へ残す。

### L3-M2内部（承認済み）

```text
G0 → M2-S1 ───────────┐
G0 → M1・M3詳細設計 ──┴→ L3合同G1

L1完了 + L3 G1
  ├─→ M2-S2
  ├─→ M1実装
  └─→ M3 Mock実装

M1・M2 Merge → M3実統合 → L3完了
```

- S1は文書・fragment、DOCTYPE、最低構造、空・BOM・空白・comment/DOCTYPEのみ・平文・破損・truncated・fence・説明文混入・複数文書・文字列やscript内だけのHTML要素・Parser回復境界を決定表にする。
- Validatorは候補HTMLを補修・抽出・再serializeせず、合格結果には原則として元の`CandidateHtml`を保持する。BOM・前後空白の判定上の扱いだけをG1で明示する。
- `SafeValidationFailure`はParser固有Message・Stack・候補全文を公開せず、安定Codeと安全な要約を返す。
- S2はCanonical Fakeと機能内Corpusを所有する。M3は並行中だけM3内の最小Fakeを使い、S2 Merge後に実Validatorとの契約互換を確認する。
- HTMLの内容品質・Sanitize・CSP・外部通信隔離は形式判定に含めずL6へ、Retry・履歴・利用者向け生成要約はM3へ残す。
- Parser依存追加時はL7共有変更laneがRoot Manifest・Lockfileを先行更新し、S2をそのCommitへrebaseする。全体DI・RouterもL7が所有する。

### L3-M1内部（承認済み）

```text
G0 → M1-S1 ───────────┐
G0 → M2・M3詳細設計 ──┴→ L3 G1

L1完了 + L3 G1 → M1-S2 || M2実装 || M3 Mock実装
M1-S2 + M2 Merge → M3実統合 → L3完了
```

- M1の1試行はProvider呼出しだけとし、Attempt番号・Retry判断・永続化・利用者向け要約を持たない。
- 候補抽出はProtocol fieldの忠実な抽出に限定し、Markdown除去・HTML修復・形式判定はM2へ残す。
- 固定の生成完了timeoutや利用者取消機能は追加しない。技術上必要なhandshake/idle deadlineだけを根拠付きでG1承認する。
- S1はCodex設定の意味とCodegen採否、S2はfeature-local型・検証・生成物、L7はRoot設定・依存・Bootstrap・DI・起動終了配線を所有する。

### L3内部（承認済み）

```text
G0 → L3-M1・M2・M3詳細設計 → L3 G1

L1完了 + L3 G1
  ├─→ M1実装 ─┐
  ├─→ M2実装 ─┼─→ M3実統合 → L3完了 → L4最終統合
  └─→ M3 Mock実装 ┘
```

- M1はProviderの1呼出し、M2はHTML形式判定、M3はRun/Attempt・再試行・履歴・管理操作を所有する。
- 自動処理は初回1回＋失敗後の自動再試行最大3回、1 Run最大4 Attemptとする。
- 手動再試行は旧Runを参照する新Runとして開始し、自動再試行予算をリセットする。
- L3は初回・修正を同じGenerationRequestで扱い、L4は修正Request編成・不変版保存・版履歴を所有する。
- L4保存失敗でCodexを再実行せず、ValidatedGenerationResultの保存を冪等に再試行できる境界とする。
- M3がRun/Attempt Schema意味Owner、L7 laneがMigration番号・共有File・Root依存・全体配線を所有する。

### L2-M2内部（承認済み）

```text
G0 → L2-M2-S1 ─┐
L2-M1-S1 ───────┴→ L2 G1

L1完了 + L2 G1
  ├─→ L2-M2-S2（Identity Mock）
  └─→ L2-M2-S3（Principal Mock）

L2-M1 contract-complete Merge → S2 Merge → S3実統合Merge → L2完了
```

- 依存方向は `VerifiedGoogleIdentity → AuthorizationDecision / AdministratorPrincipal → 管理者Session → 管理面Guard` とする。
- S2は管理者判定、S3は強制Guardを所有し、S3でraw Google Claimや管理者emailを照合しない。
- 管理者emailは原則Server専用設定とし、設定未指定・不正・複数値・未検証emailはdenyする。Gmail固有alias展開や部分一致は行わない。
- S3は認可状態を保存せずSchemaを持たない。S2も原則statelessとし、永続化が不可欠な場合だけG1で理由と単一正本を示す。
- S3はGuardと保護Route descriptorを提供し、L7がL3〜L6の全管理Routeを保護Groupへmountする。
- 匿名公開RouteへGuardを適用せず、生成HTML表示ContextへPrincipal・Sessionを渡さない。

### L2-M1内部（承認済み）

```text
G0 → L2-M1-S1 ─┐
L2-M2側詳細設計 ┴→ L2 G1

L1完了 + L2 G1
  ├─→ L2-M1-S2 ──────────────┐
  └─→ L2-M1-S3（Mock並行）───┴─→ S2 Merge → S3 Merge → L2-M1完了
```

- Session発行順は `VerifiedGoogleIdentity → M2 Authorization Port → 許可時だけAdministratorPrincipal → Session ID再生成・保存 → Cookie発行` とする。
- S1はSequence・状態・脅威対策・UI Owner・Schema要否・Test Matrix、S2はGoogle Flow、S3はLogin完了とSession lifecycleを所有する。
- S2とS3はMockで並行し、S3はS2 Merge後へrebaseして実Identity・Callback連携を確認する。
- L2-M1はAuthorization Port Contract Testで完了可能とし、実M2統合はM2統合MergeまたはL2完了条件とする。
- 認証Transaction SchemaはS2、Session SchemaはS3が意味Owner。Migration番号・共有FileはL7 laneが所有する。
- Feature-local route descriptorは各小分類、全体Router mount・DI・Root共有FileはL7が所有する。

### L2内部（承認済み）

```text
G0 ─→ L2-M1・L2-M2 詳細設計 ─→ L2 G1

L1完了 + L2 G1
  ├─→ L2-M1 実装 ───────────────┐
  └─→ L2-M2 実装（M1はMock可）──┴─→ M1 contract-complete Merge → M2実統合Merge → L2完了
```

- 受け渡しは `VerifiedGoogleIdentity → AuthorizationDecision / AdministratorPrincipal → 管理者Session発行` とする。
- M1はGoogle本人確認とSession、M2は事前登録email照合と管理認可を所有し、M1でemail判定しない。
- Session期限切れ・再Login誘導はM1、認証済み非管理者の拒否UIはM2が所有する。
- 詳細設計・本実装は別PathとMockで並行可能。M2のM1待ちは実Session統合のMerge依存とする。
- L7共有変更laneはL1完了後からManifest・Lockfile、共通設定、Migration番号・共有Path、全体Router・DIを統合する。
- L2 G1でGoogle Flow、Callback防御、Session/Cookie/CSRF、管理者照合・fail-closed、認証認可Matrix、Schema要否、L6引き渡し、Test観点を確認する。

### L1-M4内部（承認済み）

```text
G0 + L1-M3完了・依存固定Commit
       ├─→ L1-M4-S1 ────────────┐
       └─→ L1-M4-S2（並行）─────┴─→ L1-M4完了 → L1完了
```

- S1とS2は分離したPath・worktreeで並行着手し、Mergeは `S1 → S2` とする。
- S2のS1待ちは着手依存ではなく最終整合のMerge依存とし、S2が契約型を実行時参照する場合だけ直接依存へ変更する。
- 分岐前にM4の必要依存を確定し、L1-M3-S1 OwnerのManifest・Lockfile固定Commitを共通起点にする。
- S1は共有識別子・契約型・Port・契約互換Test、S2はMigration runner・基盤Migration・Migration専用Fixture・Testを所有する。
- L1-M2-S1/S2/S3を意味・論理設計の正本とし、M4で再定義しない。
- L1期間中の初期Migration番号はS2だけが予約し、L1完了後はL7共有変更laneへ移管する。

### L1-M3内部（承認済み）

```text
G0 → L1-M3-S1 → L1-M3-S2 → L1-M3完了 → L1-M4
```

- S1はRoot Manifest・Lockfile、Tool Version、Build・起動・Test executor・Lint設定、最小Bootstrap、設定雛形、Codegen共通入口の単一Ownerとする。
- S2は基盤Smoke、Test support、初期共有Fixture基盤、Codegen再現性・差分検出、clean環境検証の単一Ownerとする。
- Test executorはS1、再利用可能Test harnessはS2、契約・Migration TestはM4、機能TestはL2〜L6、E2EはL7が所有する。
- S2はManifest・Lockfile、Bootstrap、設定雛形を直接変更しない。追加依存やBootstrap修正はS1 Owner Commitを先にMergeしてrebaseする。
- Codegen生成物は一括所有せず、入力正本・出力Path・生成担当・Merge担当を生成物ごとに登録する。
- G0前は破棄可能なSpikeだけを許可し、本実装Mergeは行わない。

### L1-M2内部（承認済み）

```text
L1-M2-S1 ─┬─→ L1-M2-S2 ──────────┐
           └─→ L1-M2-S3（並行）───┴─→ L1-M2完了

L1-M1完了 + L1-M2完了 → G0
```

- S2とS3はS1 Merge後の同一clean commitから別worktreeで並行着手できる。
- Merge順は `S1 → S2 → S3` とし、S3はS2 Merge後へrebaseして契約・Schema対応表を確定する。
- S2を契約形状の正本、S3を永続・一時対応表とMigration統治の正本とし、同じ項目一覧を二重管理しない。
- S3がS2の修正を必要とする場合はS2 Ownerへ戻し、Interface文書を直接変更しない。

#### L1-M2小分類の論理Path・単一Owner

| 小分類 | 所有する論理Path・Glob | 共有編集規則 |
| --- | --- | --- |
| L1-M2-S1 | `docs/domain/glossary/**`、`docs/domain/model/**`、`docs/domain/ownership/**` | Domain Indexを変更しない |
| L1-M2-S2 | `docs/contracts/interfaces/**`、`docs/contracts/examples/**` | Domain Indexを変更しない |
| L1-M2-S3 | `docs/data/logical-schema/**`、`docs/data/migration-policy/**`、`docs/domain/README.md` | Domain Indexの単一Owner |

G0では、M1-S1〜S4とM2-S1〜S3のMerge、信頼領域と利用者・データOwner、Interface提供・利用方向と依存Matrix、全状態・共有データの単一Owner、Interface概念と論理Schemaの永続・一時対応、機能固有DDと物理設計の後続残置、Migration統治と技術ADR・L7共有変更laneの整合を確認する。

### L1-M1内部（承認済み）

```text
L1-M1-S1 → L1-M1-S2 → L1-M1-S3 ─┐
                         └→ L1-M1-S4 ┴→ L1-M1完了

L1-M1完了 + L1-M2完了 → G0
```

- S1は技術非依存の制約モデルとし、Process・Container・Origin・Hosting製品は決めない。
- S2はS1の境界を評価基準に技術を選び、S1の全境界を実現可能であることを逆照合する。
- S3とS4はS2後の同一clean commitから別worktree・別Pathで並行作業できる。
- S4の直接依存はS2だけとし、S1は推移依存として扱う。
- 実装作業はS3・S4で並行するが、共有Indexと所有表を一意にするためMergeはS3を先行する。
- G0は独立Taskにせず、S1〜S4とL1-M2の成果物整合を判定する。

#### L1-M1小分類の論理Path・単一Owner候補

| 小分類 | 所有する論理Path・Glob | 共有編集規則 |
| --- | --- | --- |
| L1-M1-S1 | `docs/architecture/topology/**`、`docs/architecture/trust-boundaries/**`、`docs/architecture/data-flows/**` | 共有Indexを変更しない |
| L1-M1-S2 | `docs/architecture/decisions/**`、`docs/architecture/technology-stack/**` | 共有Indexを変更しない |
| L1-M1-S3 | `docs/architecture/modules/**`、`docs/architecture/dependencies/**`、`docs/architecture/ownership/**` | 共有IndexとPath・Glob所有表の単一Owner |
| L1-M1-S4 | `docs/architecture/configuration/**`、`docs/architecture/secrets/**` | 共有Indexを変更しない |
| 計画管理Owner | `docs/task-map.md`、Task Map議論記録、G0承認記録 | 各小分類worktreeから変更しない |

G0では、S1とS2の実現可能性、S3とL1-M2の契約Owner・利用方向、S4とS1・S2の主体・構成要素、DD-001〜DD-007の残置、Path・Globの非重複、ADR採用版の追跡可能性を確認する。

### L1内部（承認済み）

```text
L1-M1 ─┐
       ├─→ G0 ─→ L1-M3 ─→ L1-M4 ─→ L1完了
L1-M2 ─┘
```

- L1-M1とL1-M2は別Path・別Ownerで並行し、相互レビューとG0で整合する。
- G0前は技術Spike、雛形案、Mock、契約テスト案までを許可し、共通起点への本実装MergeはG0後とする。
- L1-M3とL1-M4はRoot Manifest・Lockfileの競合を避けるため直列Mergeする。
- L2〜L6の共通起点は、L1-M4までMergeされた未Commit変更のないL1完了Commitとする。

### 全体

```text
L1 横断設計・共有契約・最小基盤
 ├─→ L2 Google認証・管理認可 ───────────────┐
 ├─→ L3 HTML生成・検証・再試行 → L4 版履歴 ─┼─→ L6 公開・隔離閲覧 → L7 統合・受入
 ├─→ L5 タグ・配置 ───────────────────────────────→ L7 統合・受入
 └─→ L7 統合骨格・共有変更lane
```

- L2、L3、L5は、共有契約を変更しない範囲で相互に独立して実装・Mergeできる。
- L4は確定InterfaceとMockを使ってL3と並行実装できるが、統合MergeはL3に依存する。
- L6はStubを使ってL2、L4と並行実装できるが、統合MergeはL2・L4に依存する。L5 Placementとの横断統合はL7で行う。
- L7の統合骨格とE2E準備はL1完了Commit後に開始できるが、完了はL2〜L6に依存する。

### 着手条件と完了・Merge条件

| Task | 着手条件 | 本番実装条件 | 完了・Merge条件 | 後続利用者 |
| --- | --- | --- | --- | --- |
| L1 | なし | G0通過後に不可逆なSchema・Migration・共有基盤実装を開始 | 横断契約と最小基盤が統合済み | L2〜L7 |
| L2 | G0の認証・信頼境界契約 | L1完了Commit＋L2 G1 | 機能固有テスト合格 | L6、L7 |
| L3 | G0の生成・版受け渡し契約 | L1完了Commit＋L3 G1 | 機能固有テスト合格 | L4、L7 |
| L4 | G0のHTML・版契約 | L1完了Commit＋L4 G1。L3はMock可能 | L3との生成結果統合、機能固有テスト合格 | L6、L7 |
| L5 | G0のHTML全体・版識別契約 | L1完了Commit＋L5 G1 | 機能固有テスト合格 | L7 |
| L6 | G0の版・公開・隔離契約 | L1完了Commit＋L6 G1。L2・L4はStub可能 | L2・L4との統合、機能固有テスト合格 | L7 |
| L7 | L1完了Commit | L1完了後に骨格開始 | L2〜L6統合、M1〜M3完了、全必須Profile合格、重大な未統合なし、G2判定用証跡完成 | G2 |

### Gate／Milestone

#### G0 横断設計Gate

技術構成、信頼境界、依存方向、ドメインモデル、Schema・Migration方針、共有Interface、所有境界を確認する。不可逆な本番Schema・Migration・共通基盤実装の開始条件とし、Task IDは付けない。

#### G1 機能実装準備Gate

L2〜L6ごとに、該当する詳細設計持越事項、状態遷移、受け渡し契約、テスト観点を確認する。全機能一括Gateにはせず、通過した分類から本番実装できる。Task IDは付けない。

| Gate | 主な確認対象 |
| --- | --- |
| L2 G1 | DD-002、Google認証、管理者判定、セッション、管理認可 |
| L3 G1 | DD-001、DD-004、Codex app-server連携、HTML形式判定、再試行・ログ契約 |
| L4 G1 | DD-006の版保存部分、修正生成の編成、不変版・履歴契約 |
| L5 G1 | DD-003、タグ・配置UI、HTML全体への関連付け |
| L6 G1 | DD-005、DD-006の公開切り替え部分、DD-007、公開状態遷移、隔離境界 |

#### G2 初期リリース受入Milestone

L7の完成した統合成果物と判定用証跡を確認し、AC-001〜AC-029の必須Profileが合格し、重大な未統合や未解決blockerがない場合に通過する。G2通過を初期リリースの完了条件とし、Task IDは付けない。

### worktree計画

- L1-M1とL1-M2は、文書Pathを分離して同一のclean commitから並行作業する。
- L1-M3はRoot Manifest・Lockfile、Build・Test設定、Bootstrap、設定雛形、Fixture基盤、コード生成規則を所有する。
- L1-M4はShared Source、共有型・Port、Migration実行基盤、基盤用Migration、Schema Fixture、契約互換テストを所有する。
- L1-M4で依存追加が必要な場合は、L1-M3 Ownerの共有変更Commitを先にMergeし、L1-M4をrebaseする。
- L1完了時にPath・Glob所有表と引き渡し記録を作成し、共有ファイルOwnerをL7の共有変更laneへ移管する。
- L2〜L6の共通起点は、L1-M4までMergeされ未Commit変更がないL1完了Commitとする。
- L2〜L6は、各機能Directoryと機能固有テスト・Fixtureを所有する。
- 初期Schema・Migration、共有Interface、初期Manifest・Lockfile、コード生成規則、初期共有FixtureはL1の単一Ownerとする。
- L1完了後の共有Schema、Migration、Interface、ルートManifest、Lockfile、生成物の変更は、L7の共有変更laneで逐次取り込む。
- Migration番号・順序は共有変更laneで予約し、複数worktreeから同時作成しない。
- composition root、全体Router、DI・Registry、統合Fixture、E2EはL7の単一Ownerとし、機能worktreeから直接変更しない。
- 機能worktreeは共有変更laneのCommitへrebaseしてからMergeし、共有変更を最終統合まで溜めない。
- L1完了後にL7-M1-S1の引受とL7-M2/M3骨格を開始する。各FeatureはL7-M1-S2 Stable CommitへrebaseしてMergeし、L7-M2へ逐次配線、L7-M3へTest Packを追加する。
- 機能意味上の依存は`L3 → L4`、`L2 + L4 → L6`で、L5は独立してL7へ統合する。その後L7-M2-S1最終Mount → L7-M1 Final Baseline → L7-M2最終Runtime Smoke → L7-M3全Profile → L7完了 → G2とする。
- 具体的なFile・Directory所有表はL1-M1-S3でG0時に確定し、L1完了時にL7-M1-S1へ引き渡す。

## 責務の配置

| 内容 | 配置先 |
| --- | --- |
| 横断アーキテクチャ、信頼境界、共有契約、初期Schema方針、最小基盤 | L1 |
| Google資格情報、認証、管理者判定、セッション、管理認可 | L2 |
| HTML生成、形式検証、自動・手動再試行、生成試行要約ログ | L3 |
| 管理者確認、修正生成の編成、不変版保存、版履歴 | L4 |
| HTML全体に紐付くタグとサイト内配置 | L5 |
| 公開版参照、公開状態遷移、管理者・匿名閲覧の隔離表示 | L6 |
| 統合配線、共有変更lane、横断E2E、受入証跡 | L7 |
| 各分類固有のUnit・Contract・Component Integrationテスト | 各実装分類 |

## 対象外

- 生成HTMLによる学習効果の測定。
- HTML生成時間の上限・目標設定。
- 特定ブラウザ・端末の表示保証。
- 測定可能なアクセシビリティ要求。
- アプリ独自の対象外トピック分類。
- 詳細内部ログの利用者向け表示。
- 原典にない本番ホスティング・運用監視。

## 更新規則

1. 分割案を提示した時点で、子タスクを「議論中」として追加する。
2. 承認後に「承認済み」へ変更し、議論経緯を[Task Map議論記録](topic-to-html-task-map-discussion.md)へ追記する。
3. 実装開始・完了は「実行状態」だけを更新し、過去の承認記録を書き換えない。
4. タスクを統合・移動・削除する場合は、先に議論記録へ理由を残してから本マップを更新する。
5. 小分類は最終実行単位とし、子タスクを作らない。作業手順やチェック項目は小分類の完了条件として管理する。
6. 承認はその時点の暫定的な構造確定とし、後続分割で矛盾、欠落、重複、粒度不一致、階層誤りが判明した場合は承認済みタスクも修正候補に戻す。
7. 承認済みタスクを変更する場合は、旧構造を議論記録へ残し、影響範囲と移行対応を示した修正案を再承認してからTask Mapへ反映する。
8. Task Mapには直接依存だけでなく、着手条件、完了・Merge条件、後続利用者、Gate、worktree所有境界を記録する。
