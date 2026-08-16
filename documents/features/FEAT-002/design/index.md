# FEAT-002 補助設計資料索引

| 資料 | 内容 |
| --- | --- |
| [HTTP契約](http-contract.md) | 共通JSON envelope、管理operation、エラー、認可の写像。 |
| [生成operation](operations/create-generation-request.md) | 初回・修正のI/O、検証、フロー、シーケンス、DB接続。 |
| [生成要求取得operation](operations/get-generation-request.md) | 再読込み用のread-only I/O、失敗写像、シーケンス。 |
| [DB schema](database-schema.md) | table、migration、制約、index、access map。 |
| [実行時設定](runtime-configuration.md) | app-server起動設定、資格情報隔離、運用確認。 |
| [Codex app-server adapter](codex-app-server-adapter.md) | process所有、v2 JSON-RPC wire、出力抽出、cleanup。 |
| [画面設計](screen-specification.md) | 管理生成画面の状態・操作・HTTP対応。 |
| [テスト戦略](test-strategy.md) | unit、repository、HTTP、UI/E2E、境界の検証。 |

提供operationは`create_generation_request`と`get_generation_request`である。前者は管理mutation guard、後者は管理read guardを使う。複数operationを順に組み合わせる利用目的はない。生成開始の応答に最終状態と候補または最終失敗を含め、画面再読込み時だけ`get_generation_request`を使う。

監査結果: 各operationは少なくとも生成開始または結果再表示に対応する。候補の版採用、公開、隔離表示operationは含めない。これらはFEAT-003/005の責務であり、公開操作の重複・範囲外混入はない。
