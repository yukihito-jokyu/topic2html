# Operation — `create_generation_request`

## 用語と目的

認証済み管理者の自然言語入力から、初回または修正の解説HTML候補を作る。初回はtopicと任意instructions、修正はFEAT-003の合格済みversionとinstructionsを入力にする。HTMLが形式合格した最初の出力だけを候補にし、初回を含め最大4回の失敗で停止する。

HTTP endpoint、wire schema、共通errorは[HTTP契約](../http-contract.md)、物理永続化は[DB schema](../database-schema.md)を正本とする。

## I/Oと検証

| input | initial | revision | 検証失敗時 |
| --- | --- | --- | --- |
| `kind` | `initial` | `revision` | 400。 |
| `topic` | trim後非空 | 指定不可 | 400。 |
| `instructions` | 任意 | trim後非空 | 400。 |
| `source_version_id` | 指定不可 | UUIDかつ参照可能な合格済み版 | 400または409。 |

初回promptは正規化済みtopicとinstructionsから、修正promptはVersionSourceが返す検証済み元HTMLとinstructionsからServer側で組み立てる。app-serverへ送るpromptは「完全なHTML文書だけを返す」ことを求めるが、モデル出力を信頼せず後段のHTML format validatorで必ず検証する。元HTML、prompt、未検証出力はHTTP応答・DBの試行記録・安全ログに含めない。

成功出力は`201`の候補metadata、全失敗は`422 generation_failed`と要求・試行metadataである。資格情報・Origin/CSRF・session障害は生成を開始せず、FEAT-001共通errorを返す。

## 主・代替・失敗フロー

```mermaid
flowchart TD
  A[POSTを受信] --> B{管理mutation guard}
  B -->|失敗| X[401 / 403 / 503を返す]
  B -->|通過| C{入力はkind別に有効か}
  C -->|いいえ| Y[400を返す]
  C -->|はい| D{revisionか}
  D -->|はい| E{VersionSourceで参照可能か}
  E -->|いいえ| Z[409を返す]
  E -->|はい| F[T1: idle期限更新とrequestをrunningで保存]
  D -->|いいえ| F
  F --> G[次の外部attemptを開始]
  G --> H[Codex app-serverを呼ぶ]
  H --> I{HTMLを取得できたか}
  I -->|はい| J{HTML形式は合格か}
  I -->|いいえ| K{attempt numberは4未満か}
  J -->|いいえ| K
  J -->|はい| L[T3: 成功attempt・candidate・成功requestを一括保存]
  L --> M[201を返す]
  K -->|はい| N[T2: failed attemptを保存]
  N --> G
  K -->|いいえ| O[T4: 最終failed attempt・失敗requestを一括保存]
  O --> P[422を返す]
```

| 条件 | 規則 |
| --- | --- |
| app-server起動/接続/終了/出力抽出の失敗 | `generation_unavailable`のfailed attempt。 |
| HTML parserまたは完全文書条件の不合格 | `invalid_html`のfailed attempt。本文は捨てる。 |
| 1〜3回目の失敗 | 試行を確定後、次の新しいapp-server interactionを開始する。 |
| 最初の形式合格 | 直ちに停止する。後続attemptを作らない。 |
| 4回目の失敗 | 要求を`completed_failed`にし、再開しない。 |
| requestまたはattemptのDB書込み失敗 | 外部再試行を開始・継続せず500を返す。request作成失敗なら外部呼出しなし。既にcommit済みの`running` requestはGETで読めるが再開しない。 |

## シーケンスと永続化

```mermaid
sequenceDiagram
  participant UI as 管理UI
  participant H as HTTP handler
  participant U as Generation usecase
  participant V as VersionSource
  participant G as Codex adapter
  participant DB as PostgreSQL
  UI->>H: POST initial/revision
  H->>H: session + Origin + CSRF検証
  H->>U: 正規化済みcommand
  opt revision
    U->>V: 合格済みversionを取得
    V-->>U: 検証済みHTML / 利用不可
  end
  U->>DB: T1: idle期限更新 + request(running)をcommit
  loop attempt 1..4、成功で終了
    U->>G: HTML生成
    G-->>U: HTML / 安全な失敗分類
    alt 形式合格
      U->>DB: T3: succeeded attempt + candidate + request成功をcommit
    else 1〜3回目の失敗
      U->>DB: T2: failed attemptをcommit
    else 4回目の失敗
      U->>DB: T4: failed attempt + request失敗をcommit
    end
  end
  U-->>H: 成功結果 / 最終失敗
  H-->>UI: 201 / 422
```

FEAT-001のmutation guardは、全認可・Origin・CSRF検査を通過した後だけT1でsessionのidle期限と最初の業務状態であるrequestを同一transactionで更新する。T1をcommitしてから外部I/Oを行い、DB transactionを外部呼出しにまたがせない。1〜3回目の失敗はT2としてcommitする。最初の成功はT3でのみ可視化する。4回目の失敗はT4でそのattemptとrequest終端状態を同じtransactionで確定する。どのwrite transactionでもrollback時は500を返し、rollback後に別のattemptや候補保存を行わない。

| 処理ID | transaction内のDB処理 | rollback時の結果 |
| --- | --- | --- |
| T1 | FEAT-001管理sessionのidle期限をUPDATEし、`generation_requests`へ`running`をINSERT | 全rollbackして500。外部呼出しなし。 |
| T2 | 1〜3回目の`generation_attempts`をfailedでINSERT | 500。既存requestは`running`のまま、次attemptなし。 |
| T3 | 成功attemptのINSERT、candidateのINSERT、requestを成功へUPDATE | 全rollbackして500。candidateなし、requestは`running`。 |
| T4 | 4回目failed attemptのINSERT、requestを最終失敗へUPDATE | 全rollbackして500。4回目attemptなし、requestは`running`。 |

## 不変条件・観測可能性・テスト

- requestごとのattemptは1から連番で最大4。candidateがあるなら成功attemptは一つだけでrequest stateは`completed_succeeded`。
- candidateがなければ成功attemptはなく、終了済みrequestは`completed_failed`で安全なfailureを持つ。
- 修正生成は元versionを読み取るだけで、元version、コンテンツ、公開状態を変更しない。
- 記録・応答で観測できるのはrequest ID、状態、試行番号、分類、安全要約、候補IDだけである。
- test doubleで「取得失敗→形式不合格→成功」「4回失敗」「最初の成功」「source versionなし」「commit失敗」を再現する。
