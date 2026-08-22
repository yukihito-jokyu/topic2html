# FEAT-002 要件

## 目的

管理者がトピックと任意の表現指示から解説HTMLを生成し、HTML形式を満たす結果だけを後続の版管理・隔離表示へ渡せるようにする。生成取得失敗とHTML検証失敗は同じ再生成フローで扱い、各試行と管理者向けの安全な失敗要約を記録する。

## 対象

- REQ-001–014
- BR-002–005、BR-016
- NFR-001、NFR-003–006（秘密情報の非露出、隔離境界との接続、生成時間を保証しない制約として参照）
- CON-001、CON-003、CON-004、ASM-001
- DEC-ARCH-001、DEC-ARCH-002、DEC-ARCH-003
- FEAT-001の管理read/mutation guard、同一origin HTTP契約、PostgreSQL migrationの契約

## 対象外

- 合格済み結果をコンテンツの版として保存し、公開版・履歴・非公開状態を管理すること（FEAT-003）。
- 生成HTMLを別originから管理者または閲覧者へ実際に表示すること、ならびにその隔離結合検証（FEAT-005）。
- タグ・掲載場所の管理（FEAT-004）。
- Codexアカウントの発行・権限・課金主体、隔離originの具体Hostを選定すること。

## 前提となる判断

[DEC-FEAT-001](decisions/DEC-FEAT-001.md)で選択肢Aが承認された。FEAT-002は検証済みHTML候補を所有し、FEAT-003が候補を合格済み版へ採用し、FEAT-005が候補・版を隔離表示する。この契約により、REQ-006はFEAT-005との共同受入れ条件、REQ-007の修正元はFEAT-003が確立した合格済み版となる。

Codex adapterの安全な実行/認証分離とshutdown契約は、それぞれ[DEC-FEAT-003](decisions/DEC-FEAT-003.md)、[DEC-FEAT-004](decisions/DEC-FEAT-004.md)で承認済みである。子processをGo Serverの環境・UIDから起動する実装は許可しない。
