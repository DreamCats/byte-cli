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
  .option("--max-len <len>", "消息最大长度，默认 1000，0 表示不截断", "1000")
  .option("--json", "JSON 格式输出")
  .action(
    async (
      traceId: string,
      opts: {
        region: string;
        psm?: string[];
        keyword?: string[];
        maxLen: string;
        json?: boolean;
      },
    ) => {
      try {
        const r = parseRegion(opts.region);
        const manager = new AuthManager(r);
        const token = await manager.getToken();

        const options: QueryOptions = {
          psmList: opts.psm ?? [],
          keywords: opts.keyword ?? [],
          maxLen: parseInt(opts.maxLen, 10),
        };

        const entries = await query(traceId, r.value, token, options);

        if (opts.json) {
          console.log(JSON.stringify(entries, null, 2));
        } else {
          printEntries(entries);
        }
      } catch (e: unknown) {
        console.error(`错误: ${(e as Error).message}`);
        process.exit(1);
      }
    },
  );

function printEntries(entries: FlattenedLogEntry[]): void {
  if (entries.length === 0) {
    console.log("未找到日志条目");
    return;
  }

  console.log(`共 ${entries.length} 条日志:\n`);

  for (const [i, entry] of entries.entries()) {
    console.log(`[${i + 1}] ${entry.level}  ${entry.psm}  ${entry.vregion}`);
    if (entry.location) console.log(`    位置: ${entry.location}`);
    console.log(`    ${entry.msg}`);
    console.log();
  }
}
