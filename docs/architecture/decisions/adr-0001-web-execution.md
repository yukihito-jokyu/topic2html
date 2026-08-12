# ADR-0001: 生成HTMLを管理文脈から分けて実行する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Web実行方式

## 課題

Codex app-serverが生成したHTMLは、内容をアプリが信頼できない。しかし要件は、そのHTMLから外部Serviceへ通信する能力を一律に禁止していない。管理者Previewと匿名Public（一般公開）の両方で、HTMLを表示しながら、管理画面、管理API、資格情報、未公開データへ戻れない実行境界が必要である。

## 評価基準

1. `TI-01`〜`TI-06` と `RA-05`〜`RA-14` を満たせること。
2. 外部通信 `RA-03` を許可しながら、アプリ資格情報を自動付与しないこと。
3. PreviewとPublicで同じ隔離保証を使えること。
4. 表示対象の解決と、HTMLの実行を別責務にできること。
5. 新版保存・公開切替・非公開化が次の匿名解決へ反映されること。
6. 配備基盤未選定の段階で、具体OriginやContainerを不必要に固定しないこと。ここでContainerはApplication Processと依存物を隔離して配備する実行単位であり、Browser内のHTML表示枠ではない。

## 候補

### A. 管理React DOMへ直接挿入する

採用しない。`DOM`はブラウザ内の画面要素ツリーであり、管理画面と同じDOMへ未信頼HTMLを入れると、`RA-06`の管理DOM・状態への到達を構造上分離できない。HTML文字列をそのまま挿入する `dangerouslySetInnerHTML` も禁止する。

### B. Server側で静的画像へ変換する

採用しない。管理文脈との分離には有利だが、生成HTML自身の操作と外部通信を失い、要件で許容されたWebページとしての動作を満たさない。

### C. 表示対象Resolverと資格情報を持たないRendererを分離する

採用する。`Resolver`は「どのHTMLを表示してよいか」を決める役割、`Renderer`は「解決済みHTMLを実行・表示するだけ」の役割である。分離が必要なのは、公開可否を決める管理判断を未信頼HTMLの実行環境へ持ち込まないためである。

## 決定と採用理由

- Preview Resolverは管理者が指定した形式検証合格版だけを解決する。
- Public Resolverは匿名要求に対して、その時点の現在公開版だけを解決する。匿名要求へ任意の版識別子を許可しない。
- 両Resolverは、解決済みHTMLと表示に必要な最小の非秘密情報だけを共通制約のRendererへ渡す。
- Rendererから管理Frontend、管理API、認証Service、Database、Codex、管理用Client Storageへ向かう経路は作らない。
- Renderer内のHTMLから外部HTTPS通信を一律禁止しない。ただしアプリのCookie、Authorization Header、Token、秘密情報を自動付与しない。
- Rendererから管理文脈へ戻るCallback（生成HTMLの操作から管理処理を呼び返す関数）、任意Message Channel（実行文脈間で自由な命令を送る通信路）、管理画面遷移を能力として渡さない。

この決定は論理境界を固定するものであり、具体的なWeb Origin、Host、Iframe、Header、Sandbox、Container、Proxy、Cache、Capabilityは固定しない。Web OriginはScheme・Host名・Portで決まるBrowserの同一出所単位、HostはHTTP要求を受ける配備先のMachineまたはService、Iframeはある文書内に別文書を表示するHTML要素、HeaderはHTTP本文と別に渡す制御情報、SandboxはBrowserがHTMLへ許す機能を制限する仕組み、Containerは前述の配備実行単位、ProxyはBrowserと外部Serviceの間で要求を中継するServer側の役割、Cacheは正本の読取結果を一時再利用する仕組み、Capabilityは列挙した操作だけを許す権限情報である。Network Policyは配備環境で通信先・方向・Portを許可または禁止する規則であり、利用者の管理認可を代行しない。HostはOrigin中のHost名だけ、Containerは表示枠、ProxyはRendererをそれぞれ意味しない。これらを区別するのは、表示実行、配備、通信制御、利用者権限を一つの隔離策だと誤認しないためである。Iframeは候補となり得るが、それだけでは通信や資格情報を含む全境界を保証しないため、ここでは採否を決めない。

## 制約

- 許可: 生成HTML本体の表示、要件で受容された外部Serviceへの通信、外部Serviceによる閲覧者側情報の観測。
- 禁止: `RA-05`〜`RA-14`、管理React DOMへの挿入、管理Clientの再利用、管理文脈へ戻る汎用Channel。
- 条件付き: 表示用の非秘密情報は、L6でFieldを列挙し、不要な管理元データを含まないと確認できる場合だけ渡せる。
- PreviewとPublicの違いは表示対象を解決する入口だけであり、Rendererの隔離強度には差を付けない。

## Version固定方針

Web実行境界は特定Library版ではなく契約として固定する。Browser EngineはPlaywright 1.62が管理する版を試験基準としてLockfileへ固定し、実運用Browserだけに依存する非標準機能は使わない。具体方式をL6で選んだ時点で、方式に関与する実行基盤、Header、Libraryの版とSupport期間を追加ADRへ記録する。

## 見直し条件

- L6でPreview/Public共通の `RA-05`〜`RA-14` 禁止を実現できない。
- 外部通信許可と資格情報非付与を同時に保証できない。
- 生成HTMLと管理文脈の間に双方向操作が新たに必要になる。
- Public ResolverがCacheなどにより公開切替・非公開化を正しく反映できない。
- 対応Browserの標準機能だけで必要な隔離を構成できない。

## 後続への引継ぎ

- L1-M4: 用途別契約を、Go/TypeScriptの一方だけへ依存しないField定義、Canonical Fixture（両言語が同じ許可・禁止結果か確かめる正本JSON例）、Port、言語別DTO/Validatorとして物理化する。この形が必要なのは、Go BackendとTypeScript FrontendのSource型を直接共有できなくてもRendererへ渡す最小情報を機械照合するためである。
- L2: 管理者Guard、Session/Cookie境界をL6へ引き渡す。Preview/PublicのRenderer契約は所有しない。
- L4: 不変版とRenderer向けContent Portを提供し、現在公開Pointerを持たない。
- L6: Origin、Header、Sandbox、Cache、Capability、Proxyを使うかという論理方式と公開状態反映を具体化する。この判断が必要なのは、Preview/Public共通隔離と外部通信許可を満たす契約を先に示さなければ、L7が検証可能なReference/CI実行トポロジー（再現可能な参照環境と継続的試験環境の接続構成）を作れないためである。
- L7: L6の契約をReference/CI実行トポロジーの実Host・Port・Origin・TLS（Transport Layer Security、通信相手と通信内容を保護する暗号化規則）・Proxy・Network Policy・Cacheへ反映し、Cookie、Header、Storage、DOM、管理API、Private Network、任意版への到達可否をBrowserとNetwork観測で検証する。Network Policyは通信先・方向・Portの許可/禁止規則であり、管理認可の代わりにはしない。Container製品と本番Hosting製品は本ADRおよび現在のL7範囲では選定せず、本番DNS、証明書発行、CDN、監視もL7に含めない。配備要件が生じた場合だけ、Ownerを明示した後続ADRで決める。

## 公式一次資料

2026-08-12確認: [Playwright release notes](https://playwright.dev/docs/release-notes)、[Playwright browser management](https://playwright.dev/docs/browsers)。S1の論理境界と到達禁止は、公式一般資料ではなく本Repositoryの `docs/architecture/` 配下にあるIssue #4成果物を根拠とする。
