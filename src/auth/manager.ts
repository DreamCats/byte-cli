import type { Region } from "./regions.js";
import { CookieNotFoundError, TokenFetchError, InvalidResponseError } from "./exceptions.js";
import * as config from "../common/config.js";
import * as cache from "../common/cache.js";
import * as http from "../common/http.js";

const NORMAL_HEADERS: Record<string, string> = {
  Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
  "Accept-Language": "en-US,en;q=0.9",
  "Accept-Encoding": "gzip, deflate, br",
};

const CODEBASE_HEADERS: Record<string, string> = {
  Accept: "application/json, text/plain, */*",
  "Accept-Language": "zh",
  Domain: "api-server",
};

export class AuthManager {
  constructor(private region: Region) {}

  async getToken(force = false): Promise<string> {
    if (!force) {
      const cached = cache.get(this.region.value);
      if (cache.isValid(cached)) return cached!.token;
    }
    const token = await this.fetchToken();
    cache.set(this.region.value, token);
    return token;
  }

  isTokenValid(): boolean {
    const cached = cache.get(this.region.value);
    return cache.isValid(cached);
  }

  private async fetchToken(): Promise<string> {
    const cfg = config.load();
    const cookie = config.getCookie(cfg, this.region.value);
    if (!cookie) throw new CookieNotFoundError(this.region.value);

    if (this.region.isCodebase) {
      return this.fetchCodebaseToken(cookie);
    }
    return this.fetchNormalToken(cookie);
  }

  private async fetchNormalToken(cookie: string): Promise<string> {
    const resp = await http.get(this.region.authUrl, {
      ...NORMAL_HEADERS,
      Cookie: `${this.region.cookieName}=${cookie}`,
    });
    if (!resp.ok) throw new TokenFetchError(`HTTP ${resp.status}`);

    const token = resp.headers.get("x-jwt-token");
    if (!token) throw new InvalidResponseError("响应头中未找到 x-jwt-token");
    return token;
  }

  private async fetchCodebaseToken(cookie: string): Promise<string> {
    const resp = await http.get(this.region.authUrl, {
      ...CODEBASE_HEADERS,
      Cookie: `${this.region.cookieName}=${cookie}`,
    });
    if (!resp.ok) throw new TokenFetchError(`HTTP ${resp.status}`);

    const data = (await resp.json()) as { code: number; message?: string; data?: Record<string, string> };
    if (data.code !== 0) throw new TokenFetchError(data.message ?? "未知错误");

    const token = data.data?.codebase_user_jwt;
    if (!token) throw new InvalidResponseError("JSON 响应中未找到 data.codebase_user_jwt");
    return token;
  }
}
