import { readFile } from "node:fs/promises";
import tailwindcss from "@tailwindcss/postcss";
import postcss from "postcss";

const source = await readFile("src/index.css", "utf8");
const result = await postcss([tailwindcss()]).process(source, {
	from: "src/index.css",
});

if (!result.css.includes("background-color: hsl(var(--background))")) {
	throw new Error("Tailwind CSSのテーマを生成できませんでした。");
}
