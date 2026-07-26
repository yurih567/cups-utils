import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { renderPng, readLayout, examplesRoot } from "./cups.js";

const name = process.argv[2] || "receipt";
const outArg = process.argv[3];
const outPath = outArg
  ? path.resolve(outArg)
  : path.join(examplesRoot, "out", `${name}.png`);

const xml = readLayout(name);
const png = await renderPng(xml);

fs.mkdirSync(path.dirname(outPath), { recursive: true });
fs.writeFileSync(outPath, png);

console.log(`PNG gerado: ${outPath} (${png.length} bytes)`);
