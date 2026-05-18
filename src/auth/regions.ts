export type RegionValue = "cn" | "i18n" | "us" | "eu" | "codebase";

export interface Region {
  value: RegionValue;
  authUrl: string;
  cookieName: string;
  isCodebase: boolean;
}

const AUTH_URLS: Record<RegionValue, string> = {
  cn: "https://cloud.bytedance.net/auth/api/v1/jwt",
  i18n: "https://cloud-i18n.bytedance.net/auth/api/v1/jwt",
  us: "https://cloud-ttp-us.bytedance.net/auth/api/v1/jwt",
  eu: "https://cloud-i18n.tiktok-eu.org/auth/api/v1/jwt",
  codebase: "https://bits.bytedance.net/api/v1/codebase_token",
};

const ALL_REGIONS: RegionValue[] = ["cn", "i18n", "us", "eu", "codebase"];

export function parseRegion(value: string): Region {
  const v = value.trim().toLowerCase() as RegionValue;
  if (!ALL_REGIONS.includes(v)) {
    throw new Error(
      `未知区域: ${value}，可选: ${ALL_REGIONS.join(", ")}`,
    );
  }
  return {
    value: v,
    authUrl: AUTH_URLS[v],
    cookieName: v === "codebase" ? "CAS_SESSION_API" : "CAS_SESSION",
    isCodebase: v === "codebase",
  };
}

export function allRegions(): Region[] {
  return ALL_REGIONS.map((v) => parseRegion(v));
}
