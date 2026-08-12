import { expect, test } from "../fixtures/test";

test("最小管理画面を表示する", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "topic2html" })).toBeVisible();
  await expect(
    page.getByText("管理画面の最小起動確認用の表示です。"),
  ).toBeVisible();
});
