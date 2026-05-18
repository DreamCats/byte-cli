import { AuthManager } from "../auth/manager.js";
import { parseRegion, type RegionValue } from "../auth/regions.js";
import * as http from "../common/http.js";
import {
  MCPServerListResponseSchema,
  ToolListResponseSchema,
  ToolCallResponseSchema,
  type MCPServerListResponse,
  type ToolListResponse,
  type ToolCallResponse,
} from "./models.js";

const MCP_DOMAIN_MAP: Record<string, string> = {
  cn: "cloud.bytedance.net",
  us: "cloud.tiktok-us.net",
  eu: "cloud-eu.tiktok-row.net",
  i18n: "cloud.tiktok-row.net",
};

const MCP_SERVER_DOMAIN_SUFFIX: Record<string, string> = {
  cn: "mcp.bytedance.net",
  us: "mcp-usttp.tiktok-us.net",
  eu: "mcp-eu.tiktok-row.net",
  i18n: "mcp.tiktok-row.net",
};

const MCP_SERVERS_PATH = "/api/v1/aipaas/api/v1/mcp/servers";

export async function getMcpServers(
  search?: string,
  env = "prod",
  region: RegionValue = "cn",
  limit = 10,
  offset = 0,
): Promise<MCPServerListResponse> {
  const domain = MCP_DOMAIN_MAP[region];
  if (!domain) throw new Error(`MCP 不支持区域: ${region}，可选: cn/us/eu/i18n`);

  const token = await new AuthManager(parseRegion(region)).getToken();
  const params = new URLSearchParams({
    env,
    limit: String(limit),
    offset: String(offset),
    search: search ?? "",
    search_type: "own",
    sort_by: "-updated_at",
    search_fields: "psm",
  });

  const resp = await http.get(`https://${domain}${MCP_SERVERS_PATH}?${params}`, {
    "x-jwt-token": token,
    "x-og-common-path-mode": "true",
    Accept: "application/json, text/plain, */*",
    "Accept-Language": "zh",
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return MCPServerListResponseSchema.parse(await resp.json());
}

export async function getMcpTools(
  serverId: string,
  region: RegionValue = "cn",
): Promise<ToolListResponse> {
  const suffix = MCP_SERVER_DOMAIN_SUFFIX[region];
  if (!suffix) throw new Error(`MCP 不支持区域: ${region}，可选: cn/us/eu/i18n`);

  const token = await new AuthManager(parseRegion(region)).getToken();
  const resp = await http.post(
    `https://${serverId}.${suffix}/mcp`,
    { method: "tools/list", params: {}, jsonrpc: "2.0", id: 1 },
    {
      "X-Jwt-Token": token,
      "X-Mcp-Internal-Request": "true",
    },
  );
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return ToolListResponseSchema.parse(await resp.json());
}

export async function callMcpTool(
  serverId: string,
  toolName: string,
  arguments_: Record<string, unknown> | null = null,
  region: RegionValue = "cn",
): Promise<ToolCallResponse> {
  const suffix = MCP_SERVER_DOMAIN_SUFFIX[region];
  if (!suffix) throw new Error(`MCP 不支持区域: ${region}，可选: cn/us/eu/i18n`);

  const token = await new AuthManager(parseRegion(region)).getToken();
  const resp = await http.post(
    `https://${serverId}.${suffix}/mcp`,
    {
      method: "tools/call",
      params: { name: toolName, arguments: arguments_ ?? {} },
      jsonrpc: "2.0",
      id: 1,
    },
    {
      "X-Jwt-Token": token,
      "X-Mcp-Internal-Request": "true",
    },
  );
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return ToolCallResponseSchema.parse(await resp.json());
}
