import { expect, test } from "@playwright/test";

const stories = [
	{
		id: "components-ui-button--default",
		name: "ボタン",
		role: "button",
		accessibleName: "保存",
	},
	{
		id: "components-ui-input--default",
		name: "入力欄",
		role: "textbox",
		accessibleName: "ホスト",
	},
] as const;

for (const story of stories) {
	test(`${story.name}のStoryを表示できる`, async ({ page }) => {
		await page.goto(`/iframe.html?id=${story.id}&viewMode=story`);

		await expect(
			page.getByRole(story.role, { name: story.accessibleName }),
		).toBeVisible();
	});
}

test("確認ダイアログを開閉できる", async ({ page }) => {
	await page.goto(
		"/iframe.html?id=components-ui-alert-dialog--confirm&viewMode=story",
	);

	await page.getByRole("button", { name: "削除" }).click();
	const dialog = page.getByRole("alertdialog", {
		name: "接続プロファイルを削除",
	});
	await expect(dialog).toBeVisible();
	await dialog.getByRole("button", { name: "キャンセル" }).click();
	await expect(dialog).toBeHidden();
});
