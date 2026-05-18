import { Command } from "commander";
import { AuthManager } from "../auth/manager.js";
import { parseRegion } from "../auth/regions.js";
import { query, type QueryOptions } from "./query.js";
import type { FlattenedLogEntry } from "./models.js";

export const logidCmd = new Command("logid")
  .description("通过 Log ID 查询分布式日志链路")
  .argument("<trace-id>", "Log ID / Trace ID")
  .option("-r, --region <region>", "区域: cn/i18n/us/eu", "cn")
  .option("-p, --psm <psm...>", "PSM 服务名过滤（可多次指定）")
  .option("-k, --keyword <keyword...>", "关键词过滤（可多次指定，OR 关系）")
  .option("-l, --level <level...>", "日志级别过滤（可多次指定，如 error warn）")
  .option("--limit <n>", "最多显示条数，默认 20", "20")
  .option("--max-len <len>", "消息最大长度，默认 300，0 表示不截断", "300")
  .option("--json", "JSON 格式输出（不限制条数和长度）")
  .action(
    async (
      traceId: string,
      opts: {
        region: string;
        psm?: string[];
        keyword?: string[];
        level?: string[];
        limit: string;
        maxLen: string;
        json?: boolean;
      },
    ) => {
      try {
        const r = parseRegion(opts.region);
        const manager = new AuthManager(r);
        const token = await manager.getToken();

        const isJson = opts.json;
        const options: QueryOptions = {
          psmList: opts.psm ?? [],
          keywords: opts.keyword ?? [],
          levels: opts.level ?? [],
          maxLen: isJson ? 0 : parseInt(opts.maxLen, 10),
        };

        let entries = await query(traceId, r.value, token, options);
        const total = entries.length;

        const limit = isJson ? 0 : parseInt(opts.limit, 10);
        if (limit > 0 && entries.length > limit) {
          entries = entries.slice(0, limit);
        }

        if (isJson) {
          console.log(JSON.stringify(entries, null, 2));
        } else {
          printEntries(entries, total);
        }
      } catch (e: unknown) {
        console.error(`错误: ${(e as Error).message}`);
        process.exit(1);
      }
    },
  );

function printEntries(entries: FlattenedLogEntry[], total: number): void {
  if (entries.length === 0) {
    console.log("未找到日志条目");
    return;
  }

  if (total > entries.length) {
    console.log(`共 ${total} 条日志（显示前 ${entries.length} 条，--limit 0 显示全部）:\n`);
  } else {
    console.log(`共 ${entries.length} 条日志:\n`);
  }

  for (const [i, entry] of entries.entries()) {
    console.log(`[${i + 1}] ${entry.level}  ${entry.psm}  ${entry.vregion}`);
    if (entry.location) console.log(`    位置: ${entry.location}`);
    console.log(`    ${entry.msg}`);
    console.log();
  }
}
