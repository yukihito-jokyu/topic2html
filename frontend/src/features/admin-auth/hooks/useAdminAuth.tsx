import {
	createContext,
	type ReactNode,
	useContext,
	useEffect,
	useMemo,
	useRef,
	useState,
} from "react";

import { createAuthenticationApi } from "@/features/admin-auth/services/authentication";
import type {
	AdminAuthState,
	AuthenticationApi,
	SameOriginPath,
} from "@/features/admin-auth/types";

export type AdminAuthBoundary = {
	state: AdminAuthState;
	bootstrap(): Promise<void>;
	fetchProtected: AdminAuthController["fetchProtected"];
	logout(): Promise<void>;
};

const AdminAuthContext = createContext<AdminAuthBoundary | null>(null);

function assertSameOriginPath(path: SameOriginPath) {
	const trustedOrigin = "https://topic2html.invalid";
	if (
		!path.startsWith("/") ||
		path.startsWith("//") ||
		new URL(path, trustedOrigin).origin !== trustedOrigin
	) {
		throw new Error("protected requests must use a same-origin path");
	}
}

export class AdminAuthController {
	#csrfToken: string | null = null;
	#state: AdminAuthState;

	constructor(
		private readonly api: AuthenticationApi,
		private readonly onStateChange: (state: AdminAuthState) => void,
		private readonly navigateToLogin: (failed?: boolean) => void,
		initialState: AdminAuthState,
		private readonly fetcher: typeof fetch = fetch,
	) {
		this.#state = initialState;
	}

	get state() {
		return this.#state;
	}

	get hasCsrfToken() {
		return this.#csrfToken !== null;
	}

	#setState(state: AdminAuthState) {
		this.#state = state;
		this.onStateChange(state);
	}

	#showLogin(failed = false) {
		this.#csrfToken = null;
		this.navigateToLogin(failed);
		this.#setState({ kind: "login", failed });
	}

	async bootstrap() {
		this.#csrfToken = null;
		this.#setState({ kind: "loading" });
		try {
			const session = await this.api.bootstrap();
			if (!session.authenticated) {
				this.#showLogin();
				return;
			}
			this.#csrfToken = session.csrfToken;
			this.#setState({ kind: "authenticated" });
		} catch {
			this.#setState({ kind: "unavailable", retry: "bootstrap" });
		}
	}

	async handleProtectedResponse(status: number) {
		if (status === 401) {
			this.#showLogin();
			return;
		}
		if (status === 403) {
			await this.bootstrap();
			return;
		}
		if (status === 503) {
			this.#csrfToken = null;
			this.#setState({ kind: "unavailable", retry: "bootstrap" });
		}
	}

	async fetchProtected(path: SameOriginPath, init: RequestInit = {}) {
		assertSameOriginPath(path);
		const method = init.method?.toUpperCase() ?? "GET";
		const headers = new Headers(init.headers);
		if (method !== "GET" && method !== "HEAD") {
			if (!this.#csrfToken) {
				await this.bootstrap();
				throw new Error("authenticated request requires a CSRF token");
			}
			headers.set("X-CSRF-Token", this.#csrfToken);
		}

		const response = await this.fetcher(path, {
			...init,
			credentials: "same-origin",
			headers,
		});
		await this.handleProtectedResponse(response.status);
		return response;
	}

	async logout() {
		const token = this.#csrfToken;
		if (!token) {
			await this.bootstrap();
			return;
		}
		try {
			const response = await this.api.logout(token);
			if (response.ok || response.status === 401) {
				this.#showLogin();
				return;
			}
			if (response.status === 403) {
				this.#setState({ kind: "authenticated", logoutProblem: "forbidden" });
				return;
			}
			this.#setState({ kind: "unavailable", retry: "logout" });
		} catch {
			this.#setState({ kind: "unavailable", retry: "logout" });
		}
	}
}

function failedLoginLocation() {
	return (
		window.location.pathname === "/admin/login" &&
		new URLSearchParams(window.location.search).get("reason") === "failed"
	);
}

export function useAdminAuth() {
	const boundary = useContext(AdminAuthContext);
	if (!boundary) {
		throw new Error("AdminAuthProvider is required for protected operations");
	}
	return boundary;
}

export function AdminAuthProvider({
	api = createAuthenticationApi(),
	children,
	fetcher,
}: {
	api?: AuthenticationApi;
	children: ReactNode;
	fetcher?: typeof fetch;
}) {
	const initialState: AdminAuthState =
		window.location.pathname === "/admin/login"
			? { kind: "login", failed: failedLoginLocation() }
			: { kind: "loading" };
	const [state, setState] = useState<AdminAuthState>(initialState);
	const controller = useRef<AdminAuthController | null>(null);
	if (!controller.current) {
		controller.current = new AdminAuthController(
			api,
			setState,
			(failed = false) => {
				window.history.replaceState(
					null,
					"",
					failed ? "/admin/login?reason=failed" : "/admin/login",
				);
			},
			initialState,
			fetcher,
		);
	}

	useEffect(() => {
		if (window.location.pathname !== "/admin/login") {
			void controller.current?.bootstrap();
		}
	}, []);

	const boundary = useMemo<AdminAuthBoundary>(
		() => ({
			state,
			bootstrap:
				controller.current?.bootstrap.bind(controller.current) ??
				(() => Promise.resolve()),
			fetchProtected:
				controller.current?.fetchProtected.bind(controller.current) ??
				(() => Promise.reject()),
			logout:
				controller.current?.logout.bind(controller.current) ??
				(() => Promise.resolve()),
		}),
		[state],
	);

	return (
		<AdminAuthContext.Provider value={boundary}>
			{children}
		</AdminAuthContext.Provider>
	);
}
