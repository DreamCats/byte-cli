import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import yaml from "js-yaml";
import { z } from "zod";

export const CONFIG_DIR = path.join(os.homedir(), ".config", "byte-cli");
export const CONFIG_FILE = path.join(CONFIG_DIR, "config.yaml");

const ProxyConfigSchema = z.object({
  http: z.string().default(""),
  https: z.string().default(""),
});

const RegionConfigSchema = z.object({
  cookie: z.string().default(""),
});

const AppConfigSchema = z.object({
  regions: z.record(RegionConfigSchema).default({}),
  proxy: ProxyConfigSchema.default({ http: "", https: "" }),
});

export type AppConfig = z.infer<typeof AppConfigSchema>;

export function load(): AppConfig {
  if (!fs.existsSync(CONFIG_FILE)) {
    return AppConfigSchema.parse({});
  }
  const raw = yaml.load(fs.readFileSync(CONFIG_FILE, "utf-8")) ?? {};
  return AppConfigSchema.parse(raw);
}

export function save(config: AppConfig): void {
  fs.mkdirSync(CONFIG_DIR, { recursive: true });
  fs.writeFileSync(CONFIG_FILE, yaml.dump(config, { lineWidth: -1 }));
}

export function getCookie(config: AppConfig, region: string): string {
  return config.regions[region]?.cookie ?? "";
}

export function setCookie(config: AppConfig, region: string, cookie: string): void {
  if (!config.regions[region]) {
    config.regions[region] = { cookie: "" };
  }
  config.regions[region].cookie = cookie;
}

export function maskCookie(cookie: string): string {
  if (cookie.length <= 8) return "*".repeat(cookie.length);
  return cookie.slice(0, 4) + "*".repeat(cookie.length - 8) + cookie.slice(-4);
}
