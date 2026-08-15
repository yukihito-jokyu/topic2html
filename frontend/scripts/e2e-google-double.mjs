import { createSign, generateKeyPairSync, randomUUID } from "node:crypto";
import { writeFileSync } from "node:fs";
import { createServer } from "node:http";

const host = "127.0.0.1";
const port = Number.parseInt(process.env.TOPIC2HTML_E2E_GOOGLE_PORT ?? "0", 10);
let issuer = "";
const allowedEmail = "admin@example.test";
const deniedEmail = "denied@example.test";
const clientID = "e2e-client-id";
const { privateKey, publicKey } = generateKeyPairSync("rsa", {
	modulusLength: 2048,
});
const key = publicKey.export({ format: "jwk" });
const authorizations = new Map();

function encode(value) {
	return Buffer.from(JSON.stringify(value)).toString("base64url");
}

function appendQuery(endpoint, query) {
	const url = new URL(endpoint);
	for (const [key, value] of Object.entries(query))
		url.searchParams.set(key, value);

	return url.toString();
}

function parseForm(request) {
	return new Promise((resolve) => {
		let body = "";
		request.setEncoding("utf8");
		request.on("data", (chunk) => {
			body += chunk;
		});
		request.on("end", () => resolve(new URLSearchParams(body)));
	});
}

function redirect(response, location) {
	response.writeHead(303, { Location: location });
	response.end();
}

function json(response, value) {
	response.writeHead(200, { "Content-Type": "application/json" });
	response.end(JSON.stringify(value));
}

function authorizationPage(query) {
	const values = new URLSearchParams({
		redirect_uri: query.get("redirect_uri") ?? "",
		state: query.get("state") ?? "",
		nonce: query.get("nonce") ?? "",
		client_id: query.get("client_id") ?? "",
	});
	const inputs = [...values]
		.map(
			([name, value]) =>
				`<input name="${name}" type="hidden" value="${value}">`,
		)
		.join("");

	return `<!doctype html><html lang="ja"><body><main><h1>Test Google Provider</h1><p>テスト用の本人確認画面です。</p><form action="/approve" method="post">${inputs}<button type="submit">承認</button></form><form action="/deny" method="post">${inputs}<button type="submit">許可しない</button></form><form action="/cancel" method="post">${inputs}<button type="submit">キャンセル</button></form></main></body></html>`;
}

function idToken(authorization, email) {
	const header = encode({ alg: "RS256", kid: "e2e-key" });
	const payload = encode({
		iss: issuer,
		aud: authorization.clientID,
		exp: Math.floor(Date.now() / 1_000) + 300,
		nonce: authorization.nonce,
		email,
		email_verified: true,
	});
	const signingInput = `${header}.${payload}`;
	const signer = createSign("RSA-SHA256");
	signer.update(signingInput);
	signer.end();

	return `${signingInput}.${signer.sign(privateKey).toString("base64url")}`;
}

const server = createServer(async (request, response) => {
	const requestURL = new URL(request.url ?? "/", issuer);
	if (
		request.method === "GET" &&
		requestURL.pathname === "/.well-known/openid-configuration"
	) {
		json(response, {
			issuer,
			authorization_endpoint: `${issuer}/authorize`,
			token_endpoint: `${issuer}/token`,
			jwks_uri: `${issuer}/jwks`,
		});

		return;
	}
	if (request.method === "GET" && requestURL.pathname === "/jwks") {
		json(response, {
			keys: [{ kty: "RSA", kid: "e2e-key", n: key.n, e: key.e }],
		});

		return;
	}
	if (request.method === "GET" && requestURL.pathname === "/authorize") {
		response.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
		response.end(authorizationPage(requestURL.searchParams));

		return;
	}
	if (
		request.method === "POST" &&
		["/approve", "/deny", "/cancel"].includes(requestURL.pathname)
	) {
		const form = await parseForm(request);
		const redirectURI = form.get("redirect_uri");
		const state = form.get("state");
		if (
			!redirectURI ||
			!state ||
			!form.get("nonce") ||
			form.get("client_id") !== clientID
		) {
			response.writeHead(400);
			response.end();

			return;
		}
		if (requestURL.pathname === "/cancel") {
			redirect(
				response,
				appendQuery(redirectURI, { error: "access_denied", state }),
			);

			return;
		}
		const code = randomUUID();
		authorizations.set(code, {
			clientID: form.get("client_id"),
			nonce: form.get("nonce"),
			email: requestURL.pathname === "/approve" ? allowedEmail : deniedEmail,
		});
		redirect(response, appendQuery(redirectURI, { code, state }));

		return;
	}
	if (request.method === "POST" && requestURL.pathname === "/token") {
		const form = await parseForm(request);
		const authorization = authorizations.get(form.get("code"));
		if (
			!authorization ||
			form.get("client_id") !== clientID ||
			form.get("client_secret") !== "e2e-client-secret"
		) {
			response.writeHead(400);
			response.end();

			return;
		}
		authorizations.delete(form.get("code"));
		json(response, { id_token: idToken(authorization, authorization.email) });

		return;
	}
	response.writeHead(404);
	response.end();
});

server.listen(port, host, () => {
	const address = server.address();
	if (!address || typeof address === "string") {
		throw new Error("Google OAuth test double did not receive a TCP port");
	}
	issuer = `http://${host}:${address.port}`;
	const endpointFile = process.env.TOPIC2HTML_E2E_GOOGLE_ENDPOINT_FILE;
	if (endpointFile)
		writeFileSync(endpointFile, `${issuer}/.well-known/openid-configuration`);
	console.log(`Google OAuth test double is listening at ${issuer}`);
});
