import { expect, test } from "vitest";

const runtimeSources = import.meta.glob<string>("../src/**/*.{ts,tsx}", {
	eager: true,
	query: "?raw",
	import: "default",
});

test("frontend runtime source uses neither backend internals nor server configuration", () => {
	const forbiddenReferences = [
		/\bbackend\//,
		/\b(?:pgx|postgres(?:ql)?):\/\//i,
		/\bhttps?:\/\/(?!topic2html\.invalid(?=["']))/,
		/\b(?:process|import\.meta)\.env\b/,
		/\bTOPIC2HTML_(?:DATABASE_URL|GOOGLE_CLIENT_SECRET|PROTECTION_KEY|ALLOWED_EMAIL)\b/,
	];

	for (const [path, source] of Object.entries(runtimeSources)) {
		for (const forbiddenReference of forbiddenReferences) {
			expect(source, path).not.toMatch(forbiddenReference);
		}
	}
});
