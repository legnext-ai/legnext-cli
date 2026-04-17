#!/usr/bin/env node
"use strict";

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const https = require("https");
const os = require("os");

const REPO = "legnext-ai/legnext-cli";
const BINARY = "legnext";
const VERSION = require("./package.json").version;

const PLATFORM_MAP = { darwin: "darwin", linux: "linux", win32: "windows" };
const ARCH_MAP = { x64: "amd64", arm64: "arm64" };

function fail(msg) {
  console.error(`@legnext/cli install error: ${msg}`);
  process.exit(1);
}

const platform = PLATFORM_MAP[os.platform()];
const arch = ARCH_MAP[os.arch()];
if (!platform) fail(`unsupported platform: ${os.platform()}`);
if (!arch) fail(`unsupported architecture: ${os.arch()}`);

const ext = platform === "windows" ? "zip" : "tar.gz";
const asset = `${BINARY}_${VERSION}_${platform}_${arch}.${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${asset}`;

const binDir = path.join(__dirname, "bin");
const binPath = path.join(binDir, platform === "windows" ? `${BINARY}.exe` : BINARY);

if (fs.existsSync(binPath)) {
  process.exit(0); // already installed
}

fs.mkdirSync(binDir, { recursive: true });

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "legnext-"));
const tmpFile = path.join(tmpDir, asset);

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (url) => {
      https.get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          follow(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode} downloading ${url}`));
          return;
        }
        const stream = fs.createWriteStream(dest);
        res.pipe(stream);
        stream.on("finish", () => { stream.close(); resolve(); });
        stream.on("error", reject);
      }).on("error", reject);
    };
    follow(url);
  });
}

async function main() {
  console.log(`Downloading ${BINARY} v${VERSION} (${platform}/${arch})...`);

  await download(url, tmpFile);

  if (ext === "zip") {
    execSync(`unzip -oq "${tmpFile}" -d "${tmpDir}"`);
  } else {
    execSync(`tar xzf "${tmpFile}" -C "${tmpDir}"`);
  }

  const src = path.join(tmpDir, platform === "windows" ? `${BINARY}.exe` : BINARY);
  fs.copyFileSync(src, binPath);
  fs.chmodSync(binPath, 0o755);

  // cleanup
  fs.rmSync(tmpDir, { recursive: true, force: true });

  console.log(`Installed ${BINARY} v${VERSION}`);
}

main().catch((err) => fail(err.message));
