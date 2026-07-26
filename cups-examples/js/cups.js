import { spawn } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
export const examplesRoot = path.resolve(__dirname, "..");
export const layoutsDir = path.join(examplesRoot, "layouts");
export const assetsDir = path.join(examplesRoot, "assets");
export const repoRoot = path.resolve(examplesRoot, "..");

function platformBinary(name) {
  const platform = process.platform;
  const arch = process.arch === "x64" ? "amd64" : process.arch;
  return `${name}-${platform}-${arch}${ext()}`;
}

function ext() {
  return process.platform === "win32" ? ".exe" : "";
}

/**
 * Resolve cups-printer / cups-template CLI invocation.
 * Order: env bin → dist/ → local binary → `go run` in the module folder.
 * Returns { command, argsPrefix, cwd }.
 */
export function resolveCommand(kind) {
  const envKey = kind === "print" ? "CUPS_PRINT_BIN" : "CUPS_RENDER_BIN";
  if (process.env[envKey]) {
    return { command: process.env[envKey], argsPrefix: [], cwd: undefined };
  }

  const name = kind === "print" ? "print" : "render";
  const moduleDir =
    kind === "print"
      ? path.join(repoRoot, "cups-printer")
      : path.join(repoRoot, "cups-template");

  const candidates = [];
  if (kind === "print") {
    candidates.push(
      path.join(examplesRoot, "bin", `cups-print${ext()}`),
      path.join(examplesRoot, "bin", `print${ext()}`),
    );
  }
  candidates.push(
    path.join(examplesRoot, "bin", platformBinary(name)),
    path.join(moduleDir, "dist", platformBinary(name)),
    path.join(moduleDir, name),
  );

  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return { command: candidate, argsPrefix: [], cwd: undefined };
    }
  }

  return {
    command: "go",
    argsPrefix: ["run", `./cmd/${name}`],
    cwd: moduleDir,
  };
}

function run(command, args, { input, cwd } = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      stdio: ["pipe", "pipe", "pipe"],
    });

    const stdout = [];
    const stderr = [];

    child.stdout.on("data", (chunk) => stdout.push(chunk));
    child.stderr.on("data", (chunk) => stderr.push(chunk));
    child.on("error", reject);
    child.on("close", (code) => {
      const out = Buffer.concat(stdout);
      const err = Buffer.concat(stderr).toString("utf8");
      if (code !== 0) {
        reject(
          new Error(
            `${command} ${args.join(" ")} exited ${code}${err ? `\n${err}` : ""}`,
          ),
        );
        return;
      }
      resolve({ stdout: out, stderr: err });
    });

    if (input != null) {
      child.stdin.end(input);
    } else {
      child.stdin.end();
    }
  });
}

/**
 * Render XML to PNG bytes via cups-template `render`.
 */
export async function renderPng(xml, { assets = assetsDir, dpi = 203 } = {}) {
  const { command, argsPrefix, cwd } = resolveCommand("render");
  const args = [
    ...argsPrefix,
    "-format",
    "png",
    "-dpi",
    String(dpi),
    "-template",
    "-",
    "-out",
    "-",
  ];
  if (assets) {
    args.push("-assets", assets);
  }
  const { stdout } = await run(command, args, { input: xml, cwd });
  return stdout;
}

/**
 * Load KEY=VALUE pairs from a .env file into a plain object.
 * Does not mutate process.env. Comments and blank lines are ignored.
 */
export function loadEnvFile(filePath = path.join(examplesRoot, ".env")) {
  const out = {};
  if (!fs.existsSync(filePath)) {
    return out;
  }
  const text = fs.readFileSync(filePath, "utf8");
  for (const raw of text.split("\n")) {
    const line = raw.trim();
    if (!line || line.startsWith("#")) continue;
    const cleaned = line.startsWith("export ") ? line.slice(7).trim() : line;
    const eq = cleaned.indexOf("=");
    if (eq <= 0) continue;
    const key = cleaned.slice(0, eq).trim();
    let val = cleaned.slice(eq + 1).trim();
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    } else {
      const cut = val.search(/[\s#]/);
      if (cut >= 0) val = val.slice(0, cut).trim();
    }
    out[key] = val;
  }
  return out;
}

/**
 * Print XML via cups-printer `print`.
 * Pass a plain name/IP/path in `dest` — cups-printer auto-detects the protocol.
 * If `dest`/`model` are omitted, values come from cups-examples/.env (JS side only).
 */
export async function printXml(xml, options = {}) {
  const fileEnv = loadEnvFile();
  const {
    dest = fileEnv.PRINTER || fileEnv.CUPS_PRINTER || fileEnv.CUPS_DEST,
    model = fileEnv.PRINTER_MODEL || fileEnv.CUPS_MODEL || "generic",
    assets = assetsDir,
    beep = 0,
    drawer = false,
    cut = true,
    feed = 3,
  } = options;

  if (!dest) {
    throw new Error(
      "dest is required (pass options.dest or set PRINTER in cups-examples/.env)",
    );
  }

  const { command, argsPrefix, cwd } = resolveCommand("print");
  const args = [
    ...argsPrefix,
    "-template",
    "-",
    "-model",
    model,
    "-dest",
    dest,
    "-feed",
    String(feed),
  ];

  if (assets) args.push("-assets", assets);
  if (beep > 0) args.push("-beep", String(beep));
  if (drawer) args.push("-drawer");
  if (!cut) args.push("-cut=false");

  await run(command, args, { input: xml, cwd });
}

export function layoutPath(name) {
  const file = name.endsWith(".xml") ? name : `${name}.xml`;
  return path.join(layoutsDir, file);
}

export function readLayout(name) {
  return fs.readFileSync(layoutPath(name));
}

export function writeTempPng(buffer, prefix = "cups-") {
  const file = path.join(os.tmpdir(), `${prefix}${Date.now()}.png`);
  fs.writeFileSync(file, buffer);
  return file;
}
