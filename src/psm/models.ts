import { z } from "zod";

const IDL_REPO_BASE = "https://code.bytedance.net";

const DOMAIN_MAP: Record<string, string> = {
  us: "cloud-ttp-us.bytedance.net",
  eu: "cloud-eu.tiktok-row.net",
  i18n: "cloud.tiktok-row.net",
};

export const IDLInfoSchema = z.object({
  psm: z.string().default(""),
  repo_name: z.string().default(""),
  idl_path: z.string().default(""),
  default_branch: z.string().default("master"),
  idl_version: z.number().default(0),
});
export type IDLInfo = z.infer<typeof IDLInfoSchema>;

export function getIdlRepoUrl(info: IDLInfo): string {
  if (!info.repo_name) return "";
  return `${IDL_REPO_BASE}/${info.repo_name}`;
}

export function getIdlRepoMainIdlUrl(info: IDLInfo): string {
  if (!info.repo_name || !info.idl_path) return "";
  return `${IDL_REPO_BASE}/${info.repo_name}/blob/${info.default_branch}/${info.idl_path}`;
}

export const EndpointSchema = z.object({
  name: z.string().default(""),
  note: z.string().default(""),
  owner: z.string().default(""),
  version: z.string().default(""),
  serializer: z.string().default(""),
  oneway: z.boolean().default(false),
  modify_time: z.number().default(0),
});
export type Endpoint = z.infer<typeof EndpointSchema>;

export function formatUpdatedAt(ep: Endpoint): string {
  if (!ep.modify_time) return "";
  return new Date(ep.modify_time * 1000).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export interface PlatformLinks {
  psm: string;
  region: string;
  tceUrl: string;
  scmUrl: string;
  tccUrl: string;
  overpassUrl: string;
}

export function generatePlatformLinks(psm: string, region: string): PlatformLinks {
  const domain = DOMAIN_MAP[region];
  if (!domain) throw new Error(`不支持的区域: ${region}，可选: us/eu/i18n`);
  const scmSearch = encodeURIComponent(psm);
  return {
    psm,
    region,
    tceUrl: `https://${domain}/tce/services?keyword=${psm}&page=1&subs_prefer=true&type=all`,
    scmUrl: `https://${domain}/scm/favor?page=1&search=${scmSearch}`,
    tccUrl: `https://${domain}/tcc/namespace/${psm}`,
    overpassUrl: `https://cloud.bytedance.net/neptune/overpass/services/${psm}`,
  };
}
