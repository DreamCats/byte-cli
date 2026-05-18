import { load } from "./config.js";

const USER_AGENT =
  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
  "AppleWebKit/537.36 (KHTML, like Gecko) " +
  "Chrome/120.0.0.0 Safari/537.36";

const DEFAULT_TIMEOUT_MS = 30_000;

function getProxy(): string | undefined {
  const config = load();
  if (config.proxy.https) return config.proxy.https;
  if (config.proxy.http) return config.proxy.http;
  return process.env.HTTPS_PROXY ?? process.env.HTTP_PROXY;
}

export interface FetchOptions extends RequestInit {
  timeout?: number;
}

export async function request(
  url: string,
  options: FetchOptions = {},
): Promise<Response> {
  const { timeout = DEFAULT_TIMEOUT_MS, ...init } = options;

  const headers: Record<string, string> = {
    "User-Agent": USER_AGENT,
    ...(init.headers as Record<string, string> ?? {}),
  };

  const proxy = getProxy();
  // Node 18+ fetch doesn't natively support proxy, but we pass it through env
  // For corporate environments, users should set HTTPS_PROXY env var
  if (proxy) {
    process.env.HTTPS_PROXY = proxy;
    process.env.HTTP_PROXY = proxy;
  }

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeout);

  try {
    const resp = await fetch(url, {
      ...init,
      headers,
      signal: controller.signal,
      redirect: "follow",
    });
    return resp;
  } finally {
    clearTimeout(timer);
  }
}

export async function get(
  url: string,
  headers?: Record<string, string>,
): Promise<Response> {
  return request(url, { method: "GET", headers });
}

export async function post(
  url: string,
  body?: unknown,
  headers?: Record<string, string>,
): Promise<Response> {
  return request(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });
}
