# FEAT-002 管理HTTP契約

## 共通規則

すべて同一originの`/admin` APIであり、JSON、`Cache-Control: no-store`を返す。FEAT-001の管理session guardを再利用する。

| operation                             | guard                                           | 共通失敗                                              |
| ------------------------------------- | ----------------------------------------------- | ----------------------------------------------------- |
| `POST /admin/generation-requests`     | 認証済みsession、信頼済みOrigin、`X-CSRF-Token` | 未認証401、Origin/CSRF不正403、session store障害503。 |
| `GET /admin/generation-requests/{id}` | 認証済みsession                                 | 未認証401、session store障害503。                     |

共通失敗はFEAT-001のerror envelopeとcodeをそのまま使う。本文・URL・ログに候補HTML、CSRF token、session値、Codexの内部IDを含めない。

## POST `/admin/generation-requests` のフローチャート

```mermaid
flowchart TD
  A[POSTを受信] --> B[mutation guardを検査]
  B -->|session不正| C[401を返す]
  B -->|OriginまたはCSRF不正| D[403を返す]
  B -->|session store障害| E[503を返す]
  B -->|通過| F{request bodyを検証}
  F -->|不正| G[400 invalid_generation_requestを返す]
  F -->|有効| H{revisionか}
  H -->|はい| I{VersionSourceで利用可能か}
  I -->|いいえ| J[409 source_version_not_availableを返す]
  I -->|はい| K[生成要求を実行]
  H -->|いいえ| K
  K -->|HTML形式合格| L[201とrequest resourceを返す]
  K -->|4回すべて失敗| M[422 generation_failedを返す]
  K -->|DB永続化不能| N[500共通安全errorを返す]
```

## POST `/admin/generation-requests` のClean Architectureシーケンス図

```mermaid
sequenceDiagram
  participant UI as 管理UI
  participant H as Handler
  participant U as Generation Usecase
  participant D as Domain Rule
  participant V as VersionSource Port
  participant G as CandidateGeneration Port
  participant S as GenerationStore Port
  participant DB as PostgreSQL Adapter
  participant C as Codex Adapter

  UI->>H: POST generation-requests
  H->>H: mutation guardとrequest bodyを検証
  alt guardまたは入力が不正
    H-->>UI: 401、403、503または400
  else POSTが有効
    H->>U: create command
    opt revision
      U->>V: 合格済み版を取得
      V-->>U: 検証済みHTMLまたはsource version not available
    end
    alt 修正元を利用不可
      U-->>H: source version not available
      H-->>UI: 409
    else 初回または修正元を利用可能
      U->>D: kindと最大4試行を判定
      U->>S: T1を確定
      S->>DB: session idle更新とrunning request
      alt T1の永続化不能
        DB-->>S: storage failure
        S-->>U: storage failure
        U-->>H: safe server error
        H-->>UI: 500
      else T1を確定
        DB-->>S: committed
        loop 最大4回、最初の合格で停止
          U->>G: HTMLを生成
          G->>C: app-server interaction
          C-->>G: HTMLまたは安全な失敗
          G-->>U: HTMLまたは失敗分類
          U->>D: HTML形式と再試行可否を判定
          alt 形式合格
            U->>S: T3を確定
            S->>DB: succeeded attempt、candidate、request成功
          else 1から3回目の失敗
            U->>S: T2を確定
            S->>DB: failed attempt
          else 4回目の失敗
            U->>S: T4を確定
            S->>DB: failed attemptとrequest失敗
          end
        end
        U-->>H: resourceまたはgeneration_failed
        H-->>UI: 201または422
      end
    end
  end
```

Handlerはpresentation adapter、UsecaseとDomain Ruleはapplication/domain、VersionSource・CandidateGeneration・GenerationStoreはoutbound port、PostgreSQL AdapterとCodex Adapterはinfrastructure adapterである。DB transactionと外部app-server interactionを同時に保持しない。各portが返す内部ID、HTML本文、外部詳細errorはHandlerより外へ渡さない。

## GET `/admin/generation-requests/{id}` のフローチャート

```mermaid
flowchart TD
  A[GETを受信] --> B[read guardを検査]
  B -->|session不正| C[401を返す]
  B -->|session store障害| D[503を返す]
  B -->|通過| E{IDはUUID形式か}
  E -->|いいえ| F[400 invalid_generation_request_idを返す]
  E -->|はい| G{requestを読取る}
  G -->|存在しない| H[404 generation_request_not_foundを返す]
  G -->|DB読取り不能| I[500共通安全errorを返す]
  G -->|存在する| J[200とrequest resourceを返す]
```

`GET`の`200`は常に`Cache-Control: no-store`とする。`running`はGETだけが観測し得る中断記録であり、POSTの最終応答には使わない。

## GET `/admin/generation-requests/{id}` のClean Architectureシーケンス図

```mermaid
sequenceDiagram
  participant UI as 管理UI
  participant H as Handler
  participant U as Generation Usecase
  participant S as GenerationStore Port
  participant DB as PostgreSQL Adapter

  UI->>H: GET generation-requests ID
  H->>H: read guardとUUID形式を検証
  alt guardまたはUUID形式が不正
    H-->>UI: 401、503または400
  else 有効なGET
    H->>U: valid request ID
    U->>S: requestを取得
    S->>DB: read-only query
    alt DB読取り不能
      DB-->>S: storage failure
      S-->>U: storage failure
      U-->>H: safe server error
      H-->>UI: 500
    else requestが存在しない
      DB-->>S: not found
      S-->>U: not found
      U-->>H: not found
      H-->>UI: 404
    else requestが存在する
      DB-->>S: resource
      S-->>U: resource
      U-->>H: resource
      H-->>UI: 200
    end
  end
```

Handlerはpresentation adapter、Usecaseはapplication、GenerationStoreはoutbound port、PostgreSQL Adapterはinfrastructure adapterである。read-only queryはtransactionを開かず、候補HTMLや内部エラー詳細はHandlerより外へ渡さない。

## Resource representation

```json
{
  "id": "4bd411cf-7f83-4d0f-bc49-7e2d1c9d8787",
  "kind": "initial",
  "state": "completed_succeeded",
  "topic": "Goのgoroutine",
  "instructions": "初学者向けに説明する",
  "source_version_id": null,
  "created_at": "2026-08-16T09:00:00Z",
  "completed_at": "2026-08-16T09:00:10Z",
  "attempts": [
    {
      "number": 1,
      "outcome": "failed",
      "failure_code": "invalid_html",
      "failure_summary": "HTML文書の形式を確認できませんでした。"
    },
    {
      "number": 2,
      "outcome": "succeeded",
      "failure_code": null,
      "failure_summary": null
    }
  ],
  "candidate": {
    "id": "8d395fab-1731-40bb-9a17-941d5a4dca10",
    "validated_at": "2026-08-16T09:00:10Z"
  },
  "failure": null
}
```

resource representationの全fieldは常に存在する。`topic`、`instructions`、`source_version_id`、`candidate`、`failure`は該当しない場合に`null`である。各attemptも`number`、`outcome`、`failure_code`、`failure_summary`を常に持ち、成功attemptのfailure fieldは`null`である。候補HTML、外部応答、未検証出力はこの表現に含めない。完了済みrequestのattempt数は1〜4、`running` requestは0〜3である。成功後のattemptは含めない。

## POST `/admin/generation-requests`

初回または修正の生成を開始し、外部処理後の最終状態を返す同期operationである。

### Request body

初回:

```json
{
  "kind": "initial",
  "topic": "Goのgoroutine",
  "instructions": "初学者向けに、図の説明を含めてください"
}
```

修正:

```json
{
  "kind": "revision",
  "source_version_id": "a129546c-18c6-46bf-a110-dc525005c1a9",
  "instructions": "例を追加し、導入を短くしてください"
}
```

| field               | rule                                                                                          |
| ------------------- | --------------------------------------------------------------------------------------------- |
| `kind`              | `initial`または`revision`。                                                                   |
| `topic`             | initialで必須のtrim後非空文字列。revisionでは指定不可。                                       |
| `instructions`      | initialでは省略可能な文字列。revisionでは必須のtrim後非空文字列。                             |
| `source_version_id` | revisionで必須のUUID。initialでは指定不可。FEAT-003の参照可能な合格済み版でなければならない。 |

未知field、型不正、kindとfieldの組合せ不正は`400 invalid_generation_request`とする。空白は検証前にtrimし、保存・prompt組立には正規化後の値を使う。

### Response

形式合格時は`201 Created`とresource representationを返す。全試行失敗時は`422 Unprocessable Content`を返す。

```json
{
  "error": {
    "code": "generation_failed",
    "summary": "HTMLの生成を完了できませんでした。新しい要求としてもう一度お試しください。"
  },
  "generation_request": {
    "id": "4bd411cf-7f83-4d0f-bc49-7e2d1c9d8787",
    "kind": "initial",
    "state": "completed_failed",
    "topic": "Goのgoroutine",
    "instructions": null,
    "source_version_id": null,
    "created_at": "2026-08-16T09:00:00Z",
    "completed_at": "2026-08-16T09:00:10Z",
    "attempts": [
      {
        "number": 1,
        "outcome": "failed",
        "failure_code": "generation_unavailable",
        "failure_summary": "生成サービスに接続できませんでした。"
      },
      {
        "number": 2,
        "outcome": "failed",
        "failure_code": "invalid_html",
        "failure_summary": "HTML文書の形式を確認できませんでした。"
      },
      {
        "number": 3,
        "outcome": "failed",
        "failure_code": "generation_unavailable",
        "failure_summary": "生成サービスに接続できませんでした。"
      },
      {
        "number": 4,
        "outcome": "failed",
        "failure_code": "generation_unavailable",
        "failure_summary": "生成サービスに接続できませんでした。"
      }
    ],
    "candidate": null,
    "failure": {
      "code": "generation_unavailable",
      "summary": "生成サービスに接続できませんでした。"
    }
  }
}
```

`source_version_not_available`は`409 Conflict`、そのほかの入力不正は400であり、いずれもCodexを呼ばず生成要求を作成しない。DB永続化不能は500の共通安全errorとし、内部詳細を返さない。

## GET `/admin/generation-requests/{id}`

生成開始後の画面再読込みに最終記録を返す。`id`はUUIDである。存在しない場合は`404 generation_request_not_found`、UUID形式不正は`400 invalid_generation_request_id`を返す。成功時は`200 OK`とresource representationを返す。`running`記録を観測した場合も同じ表現で返せるが、FEAT-002の同期HTTPでは通常クライアントに返らない。

## Error code map

| code                            | HTTP                 | 表示可能なsummary                      | retry                |
| ------------------------------- | -------------------- | -------------------------------------- | -------------------- |
| `invalid_generation_request`    | 400                  | 入力内容を確認してください。           | 修正して新規要求     |
| `invalid_generation_request_id` | 400                  | 要求IDが不正です。                     | 正しい画面へ戻る     |
| `source_version_not_available`  | 409                  | 修正元の合格済み版を利用できません。   | 別の版を選ぶ         |
| `generation_failed`             | 422                  | 生成を完了できませんでした。           | 新規要求             |
| `generation_request_not_found`  | 404                  | 生成要求が見つかりません。             | 一覧・生成画面へ戻る |
| `generation_unavailable`        | representation内のみ | 生成サービスに接続できませんでした。   | 新規要求             |
| `invalid_html`                  | representation内のみ | HTML文書の形式を確認できませんでした。 | 新規要求             |

`generation_unavailable`と`invalid_html`は試行履歴・最終failureの分類であり、外部の例外文や応答本文を置換しない。`generation_failed`の最終summaryは、最後の分類に対応する定型安全文だけである。
