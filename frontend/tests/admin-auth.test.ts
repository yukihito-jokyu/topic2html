import { expect, test } from "vitest";
import { AdminAuthController } from "@/features/admin-auth/hooks/useAdminAuth";
import { createAuthenticationApi } from "@/features/admin-auth/services/authentication";
import type { SameOriginPath } from "@/features/admin-auth/types";

test("認証済みでないbootstrapはログイン案内へ遷移できる値を返す", async () => {
	const api = createAuthenticationApi(async (input, init) => {
		expect(input).toBe("/admin/auth/session");
		expect(init?.credentials).toBe("same-origin");
		return new Response(JSON.stringify({ authenticated: false }), {
			status: 200,
		});
	});

	await expect(api.bootstrap()).resolves.toEqual({ authenticated: false });
});

test("bootstrapの障害応答は認証基盤障害として扱える", async () => {
	const api = createAuthenticationApi(
		async () => new Response(null, { status: 503 }),
	);

	await expect(api.bootstrap()).rejects.toThrow(
		"authentication bootstrap failed",
	);
});

test("保護操作の401はtokenを破棄してログイン案内へ遷移する", async () => {
	const transitions: string[] = [];
	const controller = new AdminAuthController(
		{
			bootstrap: async () => ({
				authenticated: true,
				csrfToken: crypto.randomUUID(),
			}),
			logout: async () => new Response(null, { status: 200 }),
		},
		(state) => transitions.push(state.kind),
		() => transitions.push("navigate-login"),
		{ kind: "loading" },
		async () => new Response(null, { status: 401 }),
	);

	await controller.bootstrap();
	expect(controller.hasCsrfToken).toBe(true);
	await controller.fetchProtected("/admin/example");
	expect(controller.hasCsrfToken).toBe(false);
	expect(controller.state).toEqual({ kind: "login", failed: false });
	expect(transitions).toContain("navigate-login");
});

test("保護操作の403はbootstrapでtokenを再取得し、503は障害状態へ遷移する", async () => {
	let bootstrapCount = 0;
	let sentCsrfHeader = false;
	let responseStatus = 403;
	const controller = new AdminAuthController(
		{
			bootstrap: async () => {
				bootstrapCount += 1;
				return { authenticated: true, csrfToken: crypto.randomUUID() };
			},
			logout: async () => new Response(null, { status: 200 }),
		},
		() => {},
		() => {},
		{ kind: "loading" },
		async (_input, init) => {
			if (init?.method === "POST") {
				sentCsrfHeader = new Headers(init.headers).has("X-CSRF-Token");
			}
			return new Response(null, { status: responseStatus });
		},
	);

	await controller.bootstrap();
	await controller.fetchProtected("/admin/example", { method: "POST" });
	expect(sentCsrfHeader).toBe(true);
	expect(bootstrapCount).toBe(2);
	expect(controller.state).toEqual({ kind: "authenticated" });
	responseStatus = 503;
	await controller.fetchProtected("/admin/example");
	expect(controller.hasCsrfToken).toBe(false);
	expect(controller.state).toEqual({ kind: "unavailable", retry: "bootstrap" });
});

test("保護操作は外部originの入力を通信前に拒否する", async () => {
	let networkCalls = 0;
	const controller = new AdminAuthController(
		{
			bootstrap: async () => ({
				authenticated: true,
				csrfToken: crypto.randomUUID(),
			}),
			logout: async () => new Response(null, { status: 200 }),
		},
		() => {},
		() => {},
		{ kind: "loading" },
		async () => {
			networkCalls += 1;
			return new Response(null, { status: 200 });
		},
	);

	await controller.bootstrap();
	for (const externalPath of [
		"//external.example/path",
		"https://external.example/path",
	]) {
		await expect(
			controller.fetchProtected(externalPath as SameOriginPath, {
				method: "POST",
			}),
		).rejects.toThrow("same-origin path");
	}
	expect(networkCalls).toBe(0);
	expect(controller.hasCsrfToken).toBe(true);
});

test("logoutの403はtokenを保持し、503はtokenを保持したまま再試行を提供する", async () => {
	const tokens: string[] = [];
	const responses = [403, 503, 200];
	const controller = new AdminAuthController(
		{
			bootstrap: async () => ({
				authenticated: true,
				csrfToken: crypto.randomUUID(),
			}),
			logout: async (token) => {
				tokens.push(token);
				return new Response(null, { status: responses.shift() });
			},
		},
		() => {},
		() => {},
		{ kind: "loading" },
	);

	await controller.bootstrap();
	await controller.logout();
	expect(controller.hasCsrfToken).toBe(true);
	expect(controller.state).toEqual({
		kind: "authenticated",
		logoutProblem: "forbidden",
	});
	await controller.logout();
	expect(controller.hasCsrfToken).toBe(true);
	expect(controller.state).toEqual({ kind: "unavailable", retry: "logout" });
	await controller.logout();
	expect(tokens).toHaveLength(3);
	expect(new Set(tokens)).toHaveLength(1);
	expect(controller.hasCsrfToken).toBe(false);
	expect(controller.state).toEqual({ kind: "login", failed: false });
});

test("logout通信例外はtokenを保持して再試行状態へ遷移する", async () => {
	const controller = new AdminAuthController(
		{
			bootstrap: async () => ({
				authenticated: true,
				csrfToken: crypto.randomUUID(),
			}),
			logout: async () => Promise.reject(new Error("unavailable")),
		},
		() => {},
		() => {},
		{ kind: "loading" },
	);

	await controller.bootstrap();
	await controller.logout();
	expect(controller.hasCsrfToken).toBe(true);
	expect(controller.state).toEqual({ kind: "unavailable", retry: "logout" });
});
