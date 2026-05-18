import * as http from "../common/http.js";
import { MessageSanitizer, KeywordFilter } from "./filter.js";
import {
  LogQueryResponseSchema,
  type FlattenedLogEntry,
  type LogItem,
} from "./models.js";

const LOG_SERVICE_URLS: Record<string, string> = {
  us: "https://logservice-tx.tiktok-us.org/streamlog/platform/microservice/v1/query/trace",
  i18n: "https://logservice-sg.tiktok-row.org/streamlog/platform/microservice/v1/query/trace",
  eu: "https://logservice-eu-ttp.tiktok-eu.org/streamlog/platform/microservice/v1/query/trace",
};

const VREGION_MAP: Record<string, string> = {
  us: "US-TTP,US-TTP2",
  i18n: "Singapore-Central",
  eu: "EU-TTP2,US-EastRed",
};

export interface QueryOptions {
  psmList: string[];
  keywords: string[];
  maxLen?: number;
  scanSpanMin?: number;
}

export async function query(
  logid: string,
  region: string,
  token: string,
  options: QueryOptions,
): Promise<FlattenedLogEntry[]> {
  const url = LOG_SERVICE_URLS[region];
  if (!url) throw new Error(`区域 ${region} 暂不支持日志查询（仅 us/i18n/eu）`);

  const vregion = VREGION_MAP[region] ?? "";
  const body = {
    logid,
    psm_list: options.psmList,
    scan_span_in_min: options.scanSpanMin ?? 10,
    vregion,
  };

  const resp = await http.post(url, body, {
    "X-Jwt-Token": token,
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

  const data = LogQueryResponseSchema.parse(await resp.json());
  const sanitizer = new MessageSanitizer();
  const kwFilter = new KeywordFilter(options.keywords);
  const maxLen = options.maxLen ?? 1000;

  const results: FlattenedLogEntry[] = [];
  for (const item of data.data.items) {
    results.push(...extractEntries(item, sanitizer, kwFilter, maxLen));
  }
  return results;
}

function extractEntries(
  item: LogItem,
  sanitizer: MessageSanitizer,
  kwFilter: KeywordFilter,
  maxLen: number,
): FlattenedLogEntry[] {
  const entries: FlattenedLogEntry[] = [];
  const group = item.group;

  for (const val of item.value) {
    let msg = "";
    let location = "";

    for (const kv of val.kv_list) {
      if (kv.key === "_msg") msg = kv.value;
      else if (kv.key === "_location") location = kv.value;
    }

    if (!msg) continue;
    msg = sanitizer.sanitize(msg);
    if (!kwFilter.matches(msg)) continue;
    if (maxLen > 0 && msg.length > maxLen) msg = msg.slice(0, maxLen) + "...";

    entries.push({
      psm: group.psm,
      pod_name: group.pod_name,
      level: val.level,
      msg,
      location,
      env: group.env,
      vregion: group.vregion,
    });
  }
  return entries;
}
