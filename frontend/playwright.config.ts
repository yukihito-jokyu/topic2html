import { defineConfig } from "@playwright/test";

export default defineConfig({
	testDir: "./e2e",
	timeout: 30_000,
	reporter: [["list"], ["html", { open: "never" }]],
	use: {
		baseURL: "https://localhost:5173",
		ignoreHTTPSErrors: true,
		trace: "retain-on-failure",
		video: "on",
	},
});
