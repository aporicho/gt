#!/usr/bin/env node
import { execSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const VERSION_PATH = path.join(ROOT, "VERSION");

function run(cmd) {
    console.log(`$ ${cmd}`);
    execSync(cmd, { stdio: "inherit", cwd: ROOT });
}

function bump(version, type) {
    const [major, minor, patch] = version.split(".").map(Number);
    switch (type) {
        case "major": return `${major + 1}.0.0`;
        case "minor": return `${major}.${minor + 1}.0`;
        case "patch": return `${major}.${minor}.${patch + 1}`;
        default:      return type; // 直接指定版本号
    }
}

const arg = process.argv[2];
if (!arg) {
    console.log(`用法: node scripts/release.mjs <patch|minor|major|x.y.z>`);
    console.log(`示例: node scripts/release.mjs patch`);
    process.exit(1);
}

const oldVersion = fs.readFileSync(VERSION_PATH, "utf-8").trim();
const newVersion = bump(oldVersion, arg);

console.log(`\n📦 发布 v${oldVersion} → v${newVersion}\n`);

// 1. 更新 VERSION
fs.writeFileSync(VERSION_PATH, `${newVersion}\n`);

// 2. commit
run(`git add VERSION`);
run(`git commit -m "release: v${newVersion}"`);

// 3. 打 tag
run(`git tag v${newVersion}`);

// 4. push
run(`git push origin main`);
run(`git push origin v${newVersion}`);

console.log(`\n✅ v${newVersion} 已发布！`);
console.log(`   查看进度: gh run list --repo aporicho/gtc`);
