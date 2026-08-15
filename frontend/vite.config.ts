import { defineConfig } from "vite";

const useDevelopmentTLS = process.env.TOPIC2HTML_DEV_TLS === "1";

export default defineConfig({
	resolve: {
		alias: {
			"@": new URL("./src", import.meta.url).pathname,
		},
	},
	server: {
		host: "localhost",
		port: 5173,
		strictPort: true,
		https: useDevelopmentTLS
			? {
					key: "../.certs/topic2html-dev-key.pem",
					cert: "../.certs/topic2html-dev-cert.pem",
				}
			: undefined,
		proxy: {
			"/admin/auth": "http://127.0.0.1:8080",
			"/auth": "http://127.0.0.1:8080",
		},
	},
});
