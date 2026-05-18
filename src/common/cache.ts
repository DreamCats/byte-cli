import fs from "node:fs";
import path from "node:path";
import { CONFIG_DIR } from "./config.js";

const CACHE_DIR = path.join(CONFIG_DIR, "token_cache");

const TOKEN_TTL_MS = 60 * 60 * 1000; // 1 hour
const TOKEN_BUFFER_MS = 5 * 60 * 1000; // 5 minutes

interface CachedToken {
  token: string;
  expiresAt: number; // epoch ms
}

function cachePath(region: string): string {
  return path.join(CACHE_DIR, `${region}.json`);
}

export function get(region: string): CachedToken | null {
  const p = cachePath(region);
  if (!fs.existsSync(p)) return null;
  try {
    const data = JSON.parse(fs.readFileSync(p, "utf-8"));
    return { token: data.token, expiresAt: new Date(data.expires_at).getTime() };
  } catch {
    return null;
  }
}

export function set(region: string, token: string): void {
  fs.mkdirSync(CACHE_DIR, { recursive: true });
  const expiresAt = Date.now() + TOKEN_TTL_MS;
  const data = {
    token,
    expires_at: new Date(expiresAt).toISOString(),
  };
  fs.writeFileSync(cachePath(region), JSON.stringify(data, null, 2));
}

export function del(region: string): void {
  const p = cachePath(region);
  if (fs.existsSync(p)) fs.unlinkSync(p);
}

export function isValid(cached: CachedToken | null): boolean {
  if (!cached) return false;
  return Date.now() < cached.expiresAt - TOKEN_BUFFER_MS;
}
