import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

async function startLogin(page: Page) {
	await page.goto("/admin");
	await expect(
		page.getByRole("button", { name: "Googleでログイン" }),
	).toBeVisible();
	await page.getByRole("button", { name: "Googleでログイン" }).click();
	await expect(
		page.getByRole("heading", { name: "Test Google Provider" }),
	).toBeVisible();
}

async function expectHostCookie(page: Page, name: string) {
	const cookie = (await page.context().cookies("https://localhost:5173")).find(
		(candidate) => candidate.name === name,
	);
	expect(cookie).toMatchObject({
		name,
		path: "/",
		secure: true,
		httpOnly: true,
		sameSite: "Lax",
	});
}

test("許可された利用者は実アプリケーションでログイン・logoutできる", async ({
	page,
}) => {
	await startLogin(page);
	await expectHostCookie(page, "__Host-topic2html_oauth_tx");
	await page.getByRole("button", { name: "承認" }).click();
	await expect(page).toHaveURL("https://localhost:5173/admin");
	await expect(page.getByText("認証済みです")).toBeVisible();
	await expect(page.locator("body")).not.toContainText("admin@example.test");
	await expectHostCookie(page, "__Host-topic2html_admin_session");

	await page.getByRole("button", { name: "ログアウト" }).click();
	await expect(
		page.getByRole("button", { name: "Googleでログイン" }),
	).toBeVisible();
});

test("狭い画面でもキーボードで認証操作を完了できる", async ({ page }) => {
	await page.setViewportSize({ width: 320, height: 640 });
	await page.goto("/admin");
	const login = page.getByRole("button", { name: "Googleでログイン" });
	await expect(login).toBeVisible();
	await login.focus();
	await login.press("Enter");
	await expect(
		page.getByRole("heading", { name: "Test Google Provider" }),
	).toBeVisible();
	const approve = page.getByRole("button", { name: "承認" });
	await approve.focus();
	await approve.press("Enter");
	await expect(page.getByText("認証済みです")).toBeVisible();
	const logout = page.getByRole("button", { name: "ログアウト" });
	await logout.focus();
	await logout.press("Enter");
	await expect(login).toBeVisible();
});

test("認証状態の取得中は支援技術へ状態を通知する", async ({ page }) => {
	let releaseBootstrap: () => void = () => {};
	const bootstrapReleased = new Promise<void>((resolve) => {
		releaseBootstrap = resolve;
	});
	await page.route("**/admin/auth/session", async (route) => {
		await bootstrapReleased;
		await route.fulfill({
			status: 200,
			contentType: "application/json",
			body: JSON.stringify({ authenticated: false }),
		});
	});

	const navigation = page.goto("/admin");
	await expect(page.getByRole("status")).toHaveText(
		"認証状態を確認しています。",
	);
	releaseBootstrap();
	await navigation;
	await expect(
		page.getByRole("button", { name: "Googleでログイン" }),
	).toBeVisible();
});

test("Google本人確認を取り消すと管理画面は失敗案内へ戻る", async ({ page }) => {
	await startLogin(page);
	await page.getByRole("button", { name: "キャンセル" }).click();
	await expect(page).toHaveURL(
		"https://localhost:5173/admin/login?reason=failed",
	);
	await expect(page.getByText("本人確認に失敗しました")).toBeVisible();
});

test("許可されていない利用者は管理sessionを取得できない", async ({ page }) => {
	await startLogin(page);
	await page.getByRole("button", { name: "許可しない" }).click();
	await expect(page).toHaveURL(
		"https://localhost:5173/admin/login?reason=failed",
	);
	await expect(page.getByText("本人確認に失敗しました")).toBeVisible();
});
