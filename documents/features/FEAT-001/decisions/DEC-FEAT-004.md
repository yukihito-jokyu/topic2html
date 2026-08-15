# DEC-FEAT-004 — CSRF ciphertext migration時の既存管理sessionの扱い

状態: **承認済み（2026-08-15、L3）**

## 課題

`admin_session_bootstrap`が同期CSRF tokenを返すには、Serverが平文を復元できる必要がある。既存の`admin_sessions`は照合用hashだけを持つため、既に発行済みの管理sessionからCSRF tokenを復元できない。CSRF tokenのServer保護ciphertextを追加するmigrationでは、既存sessionに安全なciphertextをbackfillできない。

## 選択肢

### A. 既存管理sessionをmigration時に失効し、再ログインを求める（推奨）

`002_admin_session_csrf_ciphertext`でciphertext列を追加し、ciphertextを持たない未失効sessionを同一transactionで失効する。既存利用者は次回bootstrapまたは管理操作で再ログインし、新規sessionはhashとciphertextを揃えて発行する。

CSRF平文を保存・推測・代替発行せず、既存のServer側sessionと同期CSRF方式を保つ。影響は、migration時点でログイン済みの管理者に一度だけ再ログインを求めることである。

### B. 既存管理sessionを継続利用する

既存sessionには復元用ciphertextがないため、CSRF tokenの平文保存、互換用の別CSRF方式、またはhashを置換するtoken再発行が必要になる。いずれも承認済みのデータ保護または公開挙動を追加変更するため、本migrationの範囲では採用しない。

## 推奨

**選択肢Aを推奨する。** 認証・CSRF保護を弱めず、既存hashから平文を復元できない不変条件を守れる唯一の安全な移行である。

## 承認後の影響

- `002_admin_session_csrf_ciphertext`を追加し、`001`を変更・再適用しない。
- migration後の新規管理sessionはCSRF hashとciphertextを保存する。
- ciphertextがないlegacy sessionは失効済みとして扱い、再ログインへ収束する。
- migration、bootstrap、CSRF照合、再ログインのintegration/HTTP/E2E検証を追加する。
