import { AuthManager } from "../auth/manager.js";
import { parseRegion } from "../auth/regions.js";
import * as http from "../common/http.js";
import { IDLInfoSchema, EndpointSchema, type IDLInfo, type Endpoint } from "./models.js";

const IDL_INFO_URL =
  "https://cloud.bytedance.net/api/v1/overpass/api/v3/overpass/platform/idl_info/get_latest_idl_info";
const API_LIST_URL = "https://cloud.bytedance.net/api/v1/bam/endpoint/list";

export async function getIdlInfo(psm: string): Promise<IDLInfo> {
  const token = await new AuthManager(parseRegion("cn")).getToken();
  const resp = await http.post(IDL_INFO_URL, { PSM: psm }, { "X-Jwt-Token": token });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

  const data = (await resp.json()) as Record<string, unknown>;
  const idlData = data.IDLInfo as Record<string, unknown> | undefined;
  if (!idlData) throw new Error(`PSM ${psm} 未找到 IDL 信息`);

  idlData.psm = psm;
  return IDLInfoSchema.parse(idlData);
}

export async function getApiList(psm: string): Promise<Endpoint[]> {
  const token = await new AuthManager(parseRegion("cn")).getToken();
  const resp = await http.get(`${API_LIST_URL}?psm=${encodeURIComponent(psm)}`, {
    "X-Jwt-Token": token,
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

  const data = (await resp.json()) as Record<string, unknown>;
  const endpoints = (data.data ?? []) as unknown[];
  return endpoints.map((ep) => EndpointSchema.parse(ep));
}
