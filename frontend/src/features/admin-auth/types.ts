export type SessionBootstrap =
	| { authenticated: false }
	| { authenticated: true; csrfToken: string };

export type AuthenticationApi = {
	bootstrap(): Promise<SessionBootstrap>;
	logout(csrfToken: string): Promise<Response>;
};

export type AdminAuthState =
	| { kind: "loading" }
	| { kind: "login"; failed: boolean }
	| { kind: "authenticated"; logoutProblem?: "forbidden" }
	| { kind: "unavailable"; retry: "bootstrap" | "logout" };

export type SameOriginPath = `/${string}`;
