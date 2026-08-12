# 共有する機能間の操作窓口

## 目的と範囲

この文書は、機能間の操作窓口、すなわち一方の機能が他方へ業務上の依頼または読取りを渡す約束を定める。画面、HTTP・RPCなどの通信方式、JSONなどの直列化形式、保存先、内部詳細ログ、実行用Fixtureは定めない。これらを分ける理由は、同じ業務規則を異なる実装方式で利用しても、識別子・所属・公開状態の意味を変えないためである。

機械識別名は実装で正確に参照する名前であり、説明そのものではない。本書では初出時に日本語の役割と制約を併記する。値の具体的な文字種・長さ・採番方式は、論理Schemaを定める後続Taskの担当であり、ここでは空でないことと再利用しないことだけを共通条件とする。

## 共通の値と結果

### 識別子と参照

| 値 | 表すもの・必要な理由 | 許可する値、禁止する値、近い値との差 |
| --- | --- | --- |
| `HtmlId`（HTML全体識別子） | 一つの解説対象を継続して管理する親記録を指す不変の値。生成依頼、版、タグ、掲載先、公開参照を同じ対象へ結ぶために必要である。 | 空でない、発行後に別対象へ再利用しない値だけを許可する。本文や表示名を代用してはならない。`VersionId`は一つの検証済み内容を指すため、HTML全体を指す`HtmlId`とは異なる。 |
| `GenerationRequestId`（生成依頼識別子） | トピックと任意の追加指示という入力を指す値。試行結果と入力を混同せず、再試行の由来を追跡するために必要である。 | 空でない値だけを許可し、一つの`HtmlId`にのみ属する。`GenerationAttemptId`は一回の実行を指すので、同じ依頼の複数試行とは区別する。 |
| `GenerationAttemptId`（生成試行識別子） | 一回の生成実行の成功または失敗を指す値。失敗も重複なく説明するために必要である。 | 試行開始時に一度だけ発行する空でない値を許可する。失敗試行に`VersionId`を持たせてはならない。 |
| `VersionId`（検証済み版識別子） | 機械検証に合格した不変の生成内容を指す値。公開対象と修正元を本文の一致に依存せず特定するために必要である。 | 合格後に一度だけ発行する空でない値を許可する。失敗・未検証結果に発行してはならず、別`HtmlId`の版を参照してはならない。 |
| `TagId`（タグ識別子） | 再利用できる分類名の記録を指す値。改名しても付与先を維持するために必要である。 | 空でない不変の値を許可する。タグ名を関連の識別子にしてはならない。 |
| `PlacementId`（掲載先識別子） | HTML全体から独立した任意名のサイト内掲載先を指す値。同じ掲載先を複数のHTML全体で共有するために必要である。 | 空でない不変の値を許可する。版や生成試行に属させず、一つの`HtmlId`が同時に参照できる掲載先は0件または1件だけとする。 |
| `PublicationReferenceId`（公開版参照識別子） | 現在閲覧者に見せる版への参照を指す値。版の不変性と、公開・非公開という現在状態を分けるために必要である。 | 空でない不変の値を許可する。版ごとの承認済み印として代用してはならない。一つの`HtmlId`には0件または1件だけ存在できる。 |
| `UserId`（利用者識別子） | 操作主体を指す値。名前やメールアドレスの変更・同名を権限判断と混同しないために必要である。 | 空でない値を許可する。認証情報、メールアドレス、セッション値をこの値と同一視してはならない。 |

### 共通結果

各操作は、成功時に操作固有の結果を返すか、失敗時に次の安全な失敗を返す。失敗は利用者へ説明できる要約であり、スタックトレース、生成器の内部出力、認証情報などの内部詳細ログを含めない。複数の失敗を一つの曖昧な成功へ変換しない理由は、利用側が再試行、修正、権限確認を正しく選べるようにするためである。

| `FailureCode`（論理的失敗分類） | 意味・必要な理由 | 該当条件、非該当条件、利用側の扱い |
| --- | --- | --- |
| `invalid-input`（入力不適格） | 必須値の欠落または禁止された組合せを表す。提供者が不完全な依頼を処理して状態を壊さないために必要である。 | 空の必須識別子、空のトピック、初回生成に修正元を渡す場合が該当する。一時的な生成器不調は該当せず`provider-unavailable`を使う。利用側は入力を修正し、新しい依頼を作る。 |
| `not-found`（対象不在） | 指定された記録が存在しないことを表す。利用側が存在しない対象へ更新しないために必要である。 | 指定識別子の記録がない場合が該当する。存在するが別HTML全体に属する場合は`ownership-mismatch`を使う。 |
| `ownership-mismatch`（所属不一致） | 参照先が存在しても、指定した`HtmlId`に属さないことを表す。別の解説対象を誤って変更・公開しないために必要である。 | `VersionId`、生成依頼、試行が別`HtmlId`に属する場合が該当する。対象不在とは異なる。 |
| `not-validated`（未検証または検証不合格） | 版化、タグ付与、掲載先設定、公開に必要な検証済み条件がないことを表す。失敗生成物を閲覧・公開状態へ混入させないために必要である。 | 成功かつ検証済みでない候補・HTML全体に操作する場合が該当する。権限不足とは異なる。 |
| `not-authorized`（管理操作の権限なし） | 操作主体に生成・管理を許可しないことを表す。未承認内容と管理状態を保護するために必要である。 | L2が不許可と判断した管理操作が該当する。公開済み版の匿名読取り可否とは別である。 |
| `conflict`（現在状態との競合） | 同じ状態を安全に更新できないことを表す。重複したタグ付与や、後続詳細設計で定める同時更新を利用側が明示的に扱うために必要である。 | 同一の`HtmlId`と`TagId`の付与が既にある場合が該当する。入力形式が不正な場合は`invalid-input`を使う。 |
| `provider-unavailable`（生成提供者を利用不能） | 生成サービスから結果を取得できないことを表す。形式検証失敗と再試行規則を混同しないために必要である。 | 生成器が結果を返せない場合が該当する。結果を取得できたがHTML形式でない場合は`validation-failed`を使う。 |
| `validation-failed`（HTML形式の検証不合格） | 生成結果を取得できたがHTML形式の機械検証に合格しないことを表す。成功した通信と利用可能な版を区別するために必要である。 | 取得済み候補の形式不合格が該当する。生成器の利用不能とは異なる。 |
| `retry-exhausted`（自動再試行上限到達） | 一つの生成依頼について自動再試行を終了したことを表す。無限試行を防ぎ、管理者による新たな手動再試行を選べるようにするために必要である。 | 最初の試行後、生成提供者の利用不能または検証不合格に対する自動再試行が3回に達した場合が該当する。手動再試行そのものは新しい依頼・試行系列なので該当しない。 |
| `not-published`（未公開） | 閲覧者向け公開版参照がないことを表す。未承認版を外部へ見せないために必要である。 | 公開版参照が0件の場合が該当する。存在しない公開対象と区別を隠す必要がある表示方式は、公開・隔離閲覧のOwnerで定める。 |

### 機能間で渡す記録の形

次の記録は、操作の入力または成功結果として渡す業務上の値の組である。必須欄がない記録は受け付けず、任意欄は表に示す条件でだけ存在する。ここで形を固定する理由は、利用側が省略された修正元・失敗理由・公開状態を推測して、別のHTML全体へ版や現在状態を結び付けることを防ぐためである。

| 記録 | 提供者→利用者 | 必須欄 | 任意欄と存在条件 | 禁止事項・近い記録との差 |
| --- | --- | --- | --- | --- |
| `ManagementPrincipal`（管理操作主体） | L2→L3〜L6 | `UserId`、管理操作の許可を表す`allowed=true` | なし | メールアドレス、認証トークン、セッション値を含めない。`UserId`だけの未確認主体や`allowed=false`はこの記録ではなく`not-authorized`である。 |
| `GenerationRequest`（生成依頼） | L4の編成側→L3 | `HtmlId`、空でないトピック、生成種別（`initial`または`correction`） | 追加生成指示は任意。修正元`VersionId`と修正指示は`correction`だけで必須、`initial`では存在禁止。 | `GenerationAttempt`は依頼を実行した一回を表すので、入力そのものを持つ本記録と異なる。 |
| `GenerationAttemptSummary`（生成試行要約） | L3→利用側 | `HtmlId`、`GenerationRequestId`、`GenerationAttemptId`、結果（成功または失敗） | 失敗時だけ`FailureCode`と利用者向け失敗要約を必須とする。成功時には両方を持たない。 | 内部詳細ログ、`VersionId`、タグ、掲載先、公開参照を含めない。検証済み本文を渡す成功結果とは異なる。 |
| `ValidatedGenerationResult`（検証済み生成結果） | L3→L4 | `HtmlId`、`GenerationRequestId`、成功した`GenerationAttemptId`、HTML本文 | 修正生成時だけ修正元`VersionId`を必須とする。初回生成では修正元は存在しない。 | HTML形式の機械検証に合格しない結果、失敗試行、失敗要約を含めない。 |
| `ImmutableVersionReference`（不変版参照） | L4→L5・L6 | `HtmlId`、`VersionId`、成功した`GenerationAttemptId`、検証済みである事実 | 修正生成の版だけ修正元`VersionId`を持つ。 | 本文の変更権限、タグ、掲載先、公開状態を含めない。版を読む必要がある場合はL4の`ReadVersion`を使う。 |
| `PublicationReference`（公開版参照） | L6→利用側 | `PublicationReferenceId`、`HtmlId`、`VersionId` | なし。存在しないこと自体が未公開を表す。 | 版ごとの承認済み印ではない。`VersionId`は同じ`HtmlId`に属する検証済み版だけを参照できる。 |
| `AnonymousPublishedProjection`（匿名閲覧用の公開対象） | L6→匿名閲覧側 | `HtmlId`、`VersionId` | なし | 管理操作主体、認証情報、未承認版、版履歴、タグ・掲載先の変更権限を含めない。これは公開中の一版を解決する読取り値であり、`PublicationReference`の更新権限を渡す値ではない。 |
| `TagReference`（現在のタグ参照） | L5→L6または統合側 | `TagId`、空でないタグ名 | なし | `VersionId`を持たない。タグ名の改名後も関連を維持するため、関連付けは名前でなく`TagId`を使う。 |
| `PlacementReference`（現在の掲載先参照） | L5→L6または統合側 | `PlacementId`、空でない掲載先名 | なし | `HtmlId`を所有しない。掲載先はHTML全体から独立し、複数のHTML全体で共有できる。 |
| `HtmlCurrentClassification`（HTML全体の現在分類） | L5→L6または統合側 | `HtmlId`、タグ参照の集合、掲載先参照なしまたは一件 | 掲載先参照は未設定時に存在しない。タグ参照の集合は空でもよい。 | 版時点の分類履歴や公開版参照を含めない。公開版切替で値を巻き戻さない。 |

### 操作ごとの入出力と失敗

操作の「利用側」は結果を受けて次の業務処理を行う機能であり、通信方式の呼出元を意味しない。すべての管理変更操作は、L2が提供する`ManagementPrincipal`を必須入力とする。これにより、L3〜L6が権限判定を複製せず、管理者以外による状態変更を防ぐ。

| 操作 | 提供側→利用側 | 必須入力 | 任意入力 | 成功結果 | 許可する`FailureCode` |
| --- | --- | --- | --- | --- | --- |
| `ResolveManagementPrincipal`（管理操作主体の解決） | L2→L3〜L6 | `UserId`、要求する管理操作種別 | なし | `ManagementPrincipal` | `invalid-input`、`not-authorized` |
| `CreateHtml`（HTML全体作成） | L4→初回生成の編成側 | `ManagementPrincipal` | なし | 新しい`HtmlId` | `not-authorized` |
| `StartGeneration`（生成開始） | L3→L4の編成側 | `ManagementPrincipal`、`GenerationRequest` | 追加生成指示だけは`GenerationRequest`内で任意 | `GenerationRequestId`、最初の`GenerationAttemptId`、開始済み`GenerationAttemptSummary` | `invalid-input`、`not-found`、`ownership-mismatch`、`not-authorized` |
| `RetryGenerationManually`（手動再試行） | L3→L4の編成側 | `ManagementPrincipal`、終了済み`GenerationRequestId` | なし | 新しい`GenerationRequestId`、最初の`GenerationAttemptId`、開始済み`GenerationAttemptSummary` | `invalid-input`、`not-found`、`not-authorized` |
| `AcceptValidatedGenerationResult`（検証済み生成結果の受入） | L3→L4 | `ManagementPrincipal`、`ValidatedGenerationResult` | なし | `ImmutableVersionReference` | `invalid-input`、`not-found`、`ownership-mismatch`、`not-validated`、`not-authorized`、`conflict` |
| `ReadVersion`（版読取り） | L4→L5・L6 | `HtmlId`、`VersionId` | なし | `ImmutableVersionReference`と、その不変HTML本文 | `invalid-input`、`not-found`、`ownership-mismatch` |
| `CreateTag`（タグ作成） | L5→管理側 | `ManagementPrincipal`、空でないタグ名 | なし | `TagReference` | `invalid-input`、`not-authorized` |
| `RenameTag`（タグ名変更） | L5→管理側 | `ManagementPrincipal`、`TagId`、空でない新しいタグ名 | なし | 更新後の`TagReference` | `invalid-input`、`not-found`、`not-authorized` |
| `AssignTag`（タグ付与） | L5→管理側 | `ManagementPrincipal`、`HtmlId`、`TagId`、L4が返す検証済み版の存在 | なし | 更新後の`HtmlCurrentClassification` | `invalid-input`、`not-found`、`not-validated`、`not-authorized`、`conflict` |
| `UnassignTag`（タグ解除） | L5→管理側 | `ManagementPrincipal`、`HtmlId`、`TagId` | なし | 更新後の`HtmlCurrentClassification` | `invalid-input`、`not-found`、`not-authorized` |
| `CreatePlacement`（掲載先作成） | L5→管理側 | `ManagementPrincipal`、空でない掲載先名 | なし | `PlacementReference` | `invalid-input`、`not-authorized` |
| `SetPlacement`（掲載先設定） | L5→管理側 | `ManagementPrincipal`、`HtmlId`、`PlacementId`、L4が返す検証済み版の存在 | なし | 更新後の`HtmlCurrentClassification` | `invalid-input`、`not-found`、`not-validated`、`not-authorized` |
| `ReadCurrentClassification`（現在分類読取り） | L5→L6または統合側 | `HtmlId` | なし | `HtmlCurrentClassification` | `invalid-input`、`not-found` |
| `SetPublicationReference`（公開版参照設定） | L6→管理側 | `ManagementPrincipal`、`HtmlId`、`VersionId`、L4が返す検証済み版の存在 | なし | `PublicationReference` | `invalid-input`、`not-found`、`ownership-mismatch`、`not-validated`、`not-authorized` |
| `ClearPublicationReference`（公開版参照取消） | L6→管理側 | `ManagementPrincipal`、`HtmlId` | なし | `HtmlId`と公開版参照が0件である事実 | `invalid-input`、`not-found`、`not-authorized` |
| `ResolvePublishedVersion`（公開版解決） | L6→匿名閲覧側 | `HtmlId` | なし | `AnonymousPublishedProjection` | `invalid-input`、`not-found`、`not-published` |

## 操作窓口と責任

### 1. L2が提供する管理操作許可の確認

`ResolveManagementPrincipal`（管理操作主体を解決する操作）は、`UserId`と管理操作の種類を受け取り、許可された場合だけ`ManagementPrincipal`（管理操作主体）を返し、不許可なら`not-authorized`を返す。管理操作の種類は、生成、修正生成、手動再試行、版保存、タグ作成・改名・付与・解除、掲載先作成・設定、公開変更を区別する。操作ごとに許可を確認する理由は、一度の本人確認を別の管理操作の無制限な許可と誤解しないためである。

L2だけが管理者かどうかを決める。L3〜L6はこの判断を利用できるが、メールアドレス、セッション、認証トークンを保存または再判定してはならない。公開済み版を匿名で読む可否は管理操作許可ではなく、L6の公開参照で判断する。

### 2. L3が提供する生成・検証・再試行

`StartGeneration`（生成開始）は、`HtmlId`、空でないトピック、任意の追加生成指示、生成種別を受け取り、`GenerationRequestId`と開始した`GenerationAttemptId`を返す。生成種別は`initial`（初回生成）または`correction`（修正生成）である。`initial`は修正元版を持たず、`correction`は同じ`HtmlId`に属する一つの`VersionId`と空でない修正指示を持つ。初回と修正を区別する理由は、修正元の系譜を初回へ誤って付けず、修正生成の由来を追跡するためである。

L3は生成依頼、試行、HTML形式の機械検証、最初の試行後に最大3回行う自動再試行、失敗要約を所有する。成功時だけ、`ValidatedGenerationResult`（検証済み生成結果）として`HtmlId`、`GenerationRequestId`、成功した`GenerationAttemptId`、候補HTML本文、修正生成の場合だけ修正元`VersionId`をL4へ渡す。失敗時は安全な失敗要約だけを保持し、本文、版、タグ、掲載先、公開参照を作ってはならない。

`RetryGenerationManually`（手動再試行）は、`retry-exhausted`となった生成依頼を管理者が指定して、新しい`GenerationRequestId`と最初の試行を作る。前の失敗試行を成功へ書き換えず、自動再試行の回数を引き継がない。これにより、手動判断による新たな実行と、終了した自動試行の履歴を区別する。

### 3. L4が提供するHTML全体と不変版

`CreateHtml`（HTML全体作成）は、生成前に新しい空でない`HtmlId`を一度だけ発行する。親記録を先に作る理由は、初回生成依頼とその後の全試行を同じ解説対象へ所属させるためである。

`AcceptValidatedGenerationResult`（検証済み生成結果の受入）は、L3から受け取った`ValidatedGenerationResult`だけを受け入れ、空でない新しい`VersionId`を発行して不変版として保存する。提供者L4は、`HtmlId`の存在、同一所属、試行の成功・検証済み、修正元版がある場合の同一所属と循環なしを確認する。成功していない試行、検証不合格結果、別HTML全体の修正元、同じ試行による二重版化を拒否する。L4は生成依頼・試行・再試行の規則を変更せず、L3は版を直接保存しない。

`ReadVersion`（版読取り）は、`HtmlId`と`VersionId`を受け取り、所属する不変本文と系譜を返す。L5とL6は適格性確認または読取りに利用できるが、本文、`HtmlId`、`VersionId`、修正元、生成試行由来を変更・削除してはならない。

### 4. L5が提供するタグと掲載先の現在状態

`CreateTag`（タグ作成）は空でないタグ名から`TagReference`を作り、`RenameTag`（タグ名変更）は既存`TagId`の表示名だけを変え、`AssignTag`（タグ付与）と`UnassignTag`（タグ解除）は`HtmlId`と`TagId`の現在の関連を変更する。`CreatePlacement`（掲載先作成）は空でない任意名から独立した`PlacementReference`を作り、`SetPlacement`（掲載先設定）は既存`PlacementId`を一つの`HtmlId`の現在の掲載先として選ぶ。付与と掲載先設定は、`HtmlId`に少なくとも一つの検証済み版があることをL4の読取りで確認してから実行する。タグと掲載先は版の属性ではなくHTML全体の現在状態であるため、版切替・公開切替で変更または巻戻ししてはならない。失敗試行・未検証HTML全体への設定、同じタグの重複付与、複数掲載先の同時設定を禁止する。

L5は現在のタグと`PlacementId`を読取りとしてL6または統合側へ渡せる。ただし公開版を決めず、L6も掲載先・タグの正本を持たない。掲載先の作成はHTML全体と独立し、まだ参照するHTML全体がなくても許可する。

### 5. L6が提供する公開版参照と閲覧対象

`SetPublicationReference`（公開版参照設定）は、`HtmlId`と`VersionId`を受け取り、L4の読取りで版の存在、同じ`HtmlId`への所属、検証済みを確認してから、0件または1件の公開版参照を設定する。新しい検証済み版が生まれただけではこの操作を自動実行してはならない。公開中の旧版を、新版を管理者が承認するまで維持するためである。過去の検証済み版を指定することも許可し、その操作はその版の公開承認を表す。

`ClearPublicationReference`（公開版参照取消）は、指定`HtmlId`の公開版参照を0件にする。版やタグ・掲載先を削除・変更せず、閲覧者向けに非公開とする。`ResolvePublishedVersion`（公開版解決）は`HtmlId`を受け取り、公開版参照がある場合だけ同じHTML全体の`HtmlId`と`VersionId`からなる`AnonymousPublishedProjection`を返す。参照がない場合は`not-published`を返し、L6はL4の版履歴や本文を変更しない。

L6は表示隔離と閲覧対象の解決を所有するが、隔離の技術方式、通信許可、画面データ形式はこの窓口に含めない。生成HTMLの外部通信を一律に禁止しない一方、アプリの認証情報と管理データへアクセスさせない、という不変条件は後続のL6詳細設計で実現する。

## 利用順序と禁止された近道

1. 管理操作ではL2の許可判断を先に利用する。L2の不許可をL3〜L6が独自に上書きしない。
2. 初回生成ではL4が`HtmlId`を作り、L3が生成依頼と試行を作る。成功し検証済みの結果だけをL4が版として受け入れる。
3. 修正生成では、L4が同じ`HtmlId`の修正元版を指定し、L3へ修正指示を渡す。L3の成功結果をL4が新しい不変版として保存する。既存版の本文を上書きしない。
4. L5のタグ・掲載先設定とL6の公開設定は、L4が提供する検証済み版の存在を前提にする。失敗試行を対象にしない。
5. 公開版を変更しても、L5のタグ・掲載先を版時点へ戻さない。公開版参照が0件なら、L6は閲覧者向けの版を解決しない。

## 互換性規則

互換性とは、提供者を更新しても、既存の利用者が既に正しく使っている依頼・結果・失敗分類を解釈し続けられることをいう。実装形式に依存しないため、項目名ではなく値の意味と制約を基準にする。

1. 必須の入力値、成功結果、失敗分類の意味、識別子の不変・所属規則、状態Owner、禁止事項は、同じ主版の間で削除・意味変更・任意化してはならない。
2. 新しい任意の読取り値、新しい任意の操作、既存分類を変えない追加の失敗補足は追加できる。利用側は未知の任意値を保存・表示に必須とせず、無視しても既存処理が正しく完結するようにする。
3. 必須値の追加、既存値の許可範囲の縮小、失敗分類の統合または意味変更、`HtmlId`と`VersionId`の関連変更、状態Ownerの移動は互換ではない。G0または該当する後続の承認手続で、提供者と利用者が明示的に合意してから別主版として導入する。
4. `FailureCode`の未知値を受け取った利用側は、成功として扱わず、安全な一般失敗として表示・記録する。ただし自動再試行は`provider-unavailable`または`validation-failed`に限り、L3が所有する規則に従う。これにより、未知の権限・所属エラーを無限再試行しない。
5. 互換性の確認は、本文の語句一致ではなく、固定例に示す初回生成、修正生成、手動再試行、失敗、未承認、旧版維持、非公開、権限なしを既存利用側が同じ意味で扱えるかで行う。

## 対象外と後続への引渡し

本書は、内部詳細ログ、HTTP/RPC、直列化形式、画面用データ転送形式、実行可能なFixtureを定めない。これらは業務上の窓口と異なる責任であり、ここで決めると通信方式や画面都合が不変条件を変えてしまうためである。

固定例は [機能間操作窓口の固定例](../examples/shared-function-interface-examples.md) に置く。後続Taskは、この操作名・値・失敗分類・互換性規則を入力にできるが、具体的な技術方式や新しい状態遷移を推測して追加してはならない。
