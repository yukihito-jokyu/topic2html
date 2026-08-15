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

test("許可された利用者は実アプリケーションでログイン・logoutできる", async ({
	page,
}) => {
	await startLogin(page);
	await page.getByRole("button", { name: "承認" }).click();
	await expect(page).toHaveURL("https://localhost:5173/admin");
	await expect(page.getByText("認証済みです")).toBeVisible();
	await expect(page.locator("body")).not.toContainText("admin@example.test");

	await page.getByRole("button", { name: "ログアウト" }).click();
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
