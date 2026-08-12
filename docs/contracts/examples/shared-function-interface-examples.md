# 機能間操作窓口の固定例

## 目的と読み方

固定例は、抽象的な操作窓口を実装前に同じ意味で解釈するための、技術形式に依存しない入出力例である。HTTPの要求・応答、JSON、画面表示、実行用Fixtureではない。`値 → 結果`は、提供者と利用者が何を渡し、どの状態を変更してよいかを確認するために必要である。

ここで使う識別子は説明用の空でない値である。値の文字列そのものに意味を持たせず、同じ値を別の記録に再利用しない。各例で`HtmlId`は解説対象、`VersionId`は検証済みで不変の内容、`GenerationAttemptId`は一回の生成実行を指す。

## 1. 初回生成と検証済み版の作成

| 順序 | 提供者と操作 | 入力 | 結果と必要な理由 |
| --- | --- | --- | --- |
| 1 | L2: `ResolveManagementPrincipal` | `UserId=admin-1`、操作=生成 | `ManagementPrincipal(UserId=admin-1, allowed=true)`。管理者だけが生成を開始できることを先に確認し、L3が認証情報を扱わないために必要である。 |
| 2 | L4: `CreateHtml` | なし | `HtmlId=html-a`。初回生成より前に親記録を作るため、後続の依頼・試行・版を同じ解説対象へ所属させられる。 |
| 3 | L3: `StartGeneration` | `HtmlId=html-a`、トピック=「太陽系」、追加生成指示なし、種別=`initial` | `GenerationRequestId=request-a1`、`GenerationAttemptId=attempt-a1`。トピックは必須で、初回なので修正元版は渡さない。 |
| 4 | L3からL4: `ValidatedGenerationResult` | `html-a`、`request-a1`、成功・検証済みの`attempt-a1`、HTML本文、修正元なし | L4へ渡せる。失敗要約や内部詳細ログは渡さず、成功かつ検証済みであることだけを版化の条件にする。 |
| 5 | L4: `AcceptValidatedGenerationResult` | 上記の検証済み生成結果 | `VersionId=version-a1`。`version-a1`は`html-a`に属する不変版であり、保存後に本文・由来を上書きしない。 |

## 2. 修正生成

| 順序 | 提供者と操作 | 入力 | 結果と必要な理由 |
| --- | --- | --- | --- |
| 1 | L2: `ResolveManagementPrincipal` | `admin-1`、操作=修正生成 | `ManagementPrincipal(UserId=admin-1, allowed=true)`。初回生成の許可と修正生成の許可を同じ値として仮定しない。 |
| 2 | L3: `StartGeneration` | `HtmlId=html-a`、トピック=「太陽系」、追加生成指示=「惑星の比較を増やす」、種別=`correction`、修正元=`VersionId=version-a1` | `GenerationRequestId=request-a2`、`GenerationAttemptId=attempt-a2`。修正元は同じ`html-a`に属する検証済み版を一つだけ指定し、初回生成とは区別する。 |
| 3 | L4: `AcceptValidatedGenerationResult` | `html-a`、成功・検証済みの`attempt-a2`、修正元=`version-a1` | `VersionId=version-a2`。`version-a1`を変更せず、新版に修正元を残す。`version-a2`を自動公開してはならない。 |

## 3. 手動再試行

| 順序 | 提供者と操作 | 入力 | 結果と必要な理由 |
| --- | --- | --- | --- |
| 1 | L3: 自動再試行 | 最初の試行後、`request-a3`で`provider-unavailable`または`validation-failed`が続く | 自動再試行が3回に達して`retry-exhausted`。上限は無限生成を防ぎ、終了を管理者へ見せるために必要である。 |
| 2 | L2: `ResolveManagementPrincipal` | `admin-1`、操作=手動再試行 | `ManagementPrincipal(UserId=admin-1, allowed=true)`。利用者が誰でも終了済みの依頼を再開できないようにする。 |
| 3 | L3: `RetryGenerationManually` | 終了済みの`GenerationRequestId=request-a3` | 新しい`GenerationRequestId=request-a4`と最初の`GenerationAttemptId=attempt-a4`。以前の失敗試行を成功に書き換えず、自動再試行の回数も引き継がない。 |

## 4. 生成提供者の失敗と検証失敗

| 条件 | L3が残す安全な失敗 | 版・現在状態への影響 | 必要な理由 |
| --- | --- | --- | --- |
| 生成器から結果を取得できない | `FailureCode=provider-unavailable`、対象試行の失敗要約 | `VersionId`を発行しない。タグ、掲載先、公開参照を作らない。 | 提供者の利用不能とHTML形式の不合格を分け、L3だけが再試行判断を行えるようにする。 |
| 結果は取得できたがHTML形式の検証に不合格 | `FailureCode=validation-failed`、対象試行の失敗要約 | `VersionId`を発行しない。失敗試行をL4へ版として渡さない。 | 通信成功を閲覧可能な版の成功と取り違えないために必要である。 |

どちらの例も内部詳細ログを利用側へ渡さない。利用側は失敗分類と要約を表示できるが、再試行の回数・終了・新しい手動再試行の規則はL3を変更してはならない。

## 5. 未承認の新版と旧版維持

前提として、`html-a`の公開版参照が`version-a1`を指し、修正生成で`version-a2`が検証済み版として保存済みだが未承認である。

| 提供者と操作 | 入力 | 結果と禁止事項 |
| --- | --- | --- |
| L6: `ResolvePublishedVersion` | `HtmlId=html-a` | `AnonymousPublishedProjection(HtmlId=html-a, VersionId=version-a1)`。`version-a2`が存在するだけで公開版参照を更新してはならない。これにより閲覧者は承認済みの旧版を継続して読む。 |
| L6: `SetPublicationReference` | `HtmlId=html-a`、`VersionId=version-a2` | L4が`version-a2`の存在、同じ`HtmlId`への所属、検証済みを確認できた場合だけ、公開版参照を`version-a2`へ設定する。これが管理者の承認である。 |
| L5: タグ・掲載先の読取り | `HtmlId=html-a` | 版切替前後で同じ現在のタグと`PlacementId`を返す。タグ・掲載先を`version-a1`または`version-a2`へ保存・巻戻ししてはならない。 |

## 6. 非公開

| 順序 | 提供者と操作 | 入力 | 結果と必要な理由 |
| --- | --- | --- | --- |
| 1 | L2: `ResolveManagementPrincipal` | `admin-1`、操作=公開取消 | `ManagementPrincipal(UserId=admin-1, allowed=true)`。公開状態の変更は管理操作なので必要である。 |
| 2 | L6: `ClearPublicationReference` | `HtmlId=html-a` | 公開版参照を0件にする。版、タグ、掲載先、生成試行を削除しない。 |
| 3 | L6: `ResolvePublishedVersion` | `HtmlId=html-a` | `not-published`。閲覧者向けに解決する版がないことを表し、未承認・取消後の版を見せないために必要である。 |

## 7. 権限なし

| 提供者と操作 | 入力 | 結果と必要な理由 |
| --- | --- | --- |
| L2: `ResolveManagementPrincipal` | `UserId=viewer-1`、操作=公開変更 | `not-authorized`。閲覧者または未登録利用者に管理状態の変更を許さない。 |
| L3/L4/L5/L6 | L2の不許可判断を伴う生成、版保存、タグ・掲載先変更、公開変更 | 操作を実行せず`not-authorized`を返す。各機能はメールアドレスやセッションを再判定せず、L2の判断だけを利用する。 |
| L6: `ResolvePublishedVersion` | `HtmlId=html-a`、公開版参照あり | `AnonymousPublishedProjection`として公開版だけを返せる。匿名閲覧可能な公開版の読取りは、管理操作の権限なしとは異なる。 |

## 8. タグと掲載先の作成・現在状態の変更

前提として、`html-a`には`version-a1`という検証済み版があり、管理操作に許可された`ManagementPrincipal(UserId=admin-1, allowed=true)`を各変更操作へ渡す。これによりL5は管理者判定を複製せず、未検証のHTML全体を分類対象へしない。

| 順序 | 提供者と操作 | 必須入力 | 成功結果または失敗 | 必要な理由 |
| --- | --- | --- | --- | --- |
| 1 | L5: `CreateTag` | 管理操作主体、タグ名=「天文学」 | `TagReference(TagId=tag-a, name=天文学)` | タグ名と不変の`TagId`を分け、後から改名しても付与先を維持する。 |
| 2 | L5: `RenameTag` | 管理操作主体、`TagId=tag-a`、新しいタグ名=「宇宙科学」 | `TagReference(TagId=tag-a, name=宇宙科学)` | `TagId`は変えない。タグ名で関連を特定すると改名時に付与先を失うためである。 |
| 3 | L5: `AssignTag` | 管理操作主体、`HtmlId=html-a`、`TagId=tag-a`、検証済み版の存在 | `HtmlCurrentClassification(HtmlId=html-a, tags=[tag-a], placement=なし)` | タグは版ではなくHTML全体の現在状態へ付ける。同じ組合せをもう一度付与する場合は`conflict`であり、重複を作らない。 |
| 4 | L5: `UnassignTag` | 管理操作主体、`HtmlId=html-a`、`TagId=tag-a` | `HtmlCurrentClassification(HtmlId=html-a, tags=[], placement=なし)` | タグを解除しても、タグ記録、版、公開参照を削除しない。 |
| 5 | L5: `CreatePlacement` | 管理操作主体、掲載先名=「学習コンテンツ」 | `PlacementReference(PlacementId=placement-a, name=学習コンテンツ)` | 掲載先はHTML全体から独立し、まだ参照先がなくても作成できる。任意名を固定候補に制限しない。 |
| 6 | L5: `SetPlacement` | 管理操作主体、`HtmlId=html-a`、`PlacementId=placement-a`、検証済み版の存在 | `HtmlCurrentClassification(HtmlId=html-a, tags=[], placement=placement-a)` | 一つのHTML全体が現在参照できる掲載先は0件または1件だけである。版や生成試行へ設定しない。 |
| 7 | L5: `ReadCurrentClassification` | `HtmlId=html-a` | 上記と同じ`HtmlCurrentClassification` | L6または統合側は読取りだけを利用し、公開版切替でタグ・掲載先を巻き戻さない。 |

## 例を使う互換性確認

提供者または利用者を更新する際は、上の各例で必須入力、成功結果、失敗分類、識別子の所属、状態Ownerが変わらないことを確認する。新しい任意読取り値は追加できるが、初回生成へ修正元を必須化する、失敗試行へ版を発行する、新版を自動公開する、閲覧者へ管理操作を許す変更は互換ではない。これらは既存利用者の判断を変えるため、別主版と明示的な承認が必要である。
