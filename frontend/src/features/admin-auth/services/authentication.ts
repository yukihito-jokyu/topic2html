import type {
	AuthenticationApi,
	SessionBootstrap,
} from "@/features/admin-auth/types";

function isAuthenticatedSession(
	value: unknown,
): value is { authenticated: true; csrf_token: string } {
	return (
		typeof value === "object" &&
		value !== null &&
		"authenticated" in value &&
		value.authenticated === true &&
		"csrf_token" in value &&
		typeof value.csrf_token === "string" &&
		value.csrf_token.length > 0
	);
}

function isAnonymousSession(value: unknown): value is { authenticated: false } {
	return (
		typeof value === "object" &&
		value !== null &&
		"authenticated" in value &&
		value.authenticated === false
	);
}

/** 同一originの管理認証HTTP契約を画面用の値へ変換する。 */
export function createAuthenticationApi(
	fetcher: typeof fetch = fetch,
): AuthenticationApi {
	return {
		async bootstrap(): Promise<SessionBootstrap> {
			const response = await fetcher("/admin/auth/session", {
				credentials: "same-origin",
			});
			if (!response.ok) {
				throw new Error("authentication bootstrap failed");
			}

			const body: unknown = await response.json();
			if (isAuthenticatedSession(body)) {
				return { authenticated: true, csrfToken: body.csrf_token };
			}
			if (isAnonymousSession(body)) {
				return { authenticated: false };
			}

			throw new Error("invalid authentication bootstrap response");
		},
		logout(csrfToken) {
			return fetcher("/admin/auth/logout", {
				method: "POST",
				credentials: "same-origin",
				headers: { "X-CSRF-Token": csrfToken },
			});
		},
	};
}
