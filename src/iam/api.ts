import { AuthManager } from "../auth/manager.js";
import { parseRegion, type RegionValue } from "../auth/regions.js";
import * as http from "../common/http.js";
import {
  ServiceAccountListResponseSchema,
  ServiceAccountSchema,
  type ServiceAccount,
  type ServiceAccountListResponse,
} from "./models.js";

const IAM_DOMAIN_MAP: Record<string, string> = {
  cn: "cloud.bytedance.net",
  us: "cloud.tiktok-us.net",
  eu: "cloud-eu.tiktok-row.net",
  i18n: "cloud.tiktok-row.net",
};

const IAM_PATH = "/api/v1/iam/api/v2/service_account/list_by_page";
const IAM_SECRET_PATH = "/api/v1/iam/api/v1/service_account/secret";

export async function getServiceAccounts(
  owner?: string,
  region: RegionValue = "cn",
  page = 1,
  size = 10,
): Promise<ServiceAccountListResponse> {
  const domain = IAM_DOMAIN_MAP[region];
  if (!domain) throw new Error(`IAM 不支持区域: ${region}，可选: cn/us/eu/i18n`);

  const token = await new AuthManager(parseRegion(region)).getToken();
  const params = new URLSearchParams({
    mine_list: "1",
    page: String(page),
    size: String(size),
  });
  if (owner) {
    params.set("owner", owner);
    params.set("owner_check", "1");
  }

  const resp = await http.get(`https://${domain}${IAM_PATH}?${params}`, {
    "X-Jwt-Token": token,
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return ServiceAccountListResponseSchema.parse(await resp.json());
}

export async function getServiceAccountSecret(
  name: string,
  region: RegionValue = "cn",
): Promise<ServiceAccount> {
  const domain = IAM_DOMAIN_MAP[region];
  if (!domain) throw new Error(`IAM 不支持区域: ${region}，可选: cn/us/eu/i18n`);

  const token = await new AuthManager(parseRegion(region)).getToken();
  const resp = await http.post(
    `https://${domain}${IAM_SECRET_PATH}?name=${encodeURIComponent(name)}`,
    undefined,
    { "X-Jwt-Token": token },
  );
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

  const data = (await resp.json()) as { error_code: number; error_message?: string; data?: unknown };
  if (data.error_code !== 0) throw new Error(data.error_message ?? "查询密钥失败");
  return ServiceAccountSchema.parse(data.data);
}
