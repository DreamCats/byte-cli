import { Command } from "commander";
import * as api from "./api.js";
import { formatServerDate } from "./models.js";
import type { RegionValue } from "../auth/regions.js";

export const mcpCmd = new Command("mcp").description("MCP Server 查询");

mcpCmd
  .command("list")
  .description("查询 MCP Server 列表")
  .option("-s, --search <search>", "按 PSM 搜索")
  .option("-e, --env <env>", "环境名称", "prod")
  .option("-r, --region <region>", "区域", "cn")
  .option("-l, --limit <limit>", "每页数量", "10")
  .option("-o, --offset <offset>", "偏移量", "0")
  .option("--json", "JSON 格式输出")
  .action(
    async (opts: {
      search?: string;
      env: string;
      region: string;
      limit: string;
      offset: string;
      json?: boolean;
    }) => {
      const resp = await api.getMcpServers(
        opts.search,
        opts.env,
        opts.region as RegionValue,
        parseInt(opts.limit, 10),
        parseInt(opts.offset, 10),
      );

      if (opts.json) {
        console.log(JSON.stringify(resp, null, 2));
        return;
      }

      if (resp.data.length === 0) {
        console.log("未找到 MCP Server");
        return;
      }

      console.log(`区域: ${opts.region}`);
      console.log(`环境: ${opts.env}`);
      console.log(`数量: ${resp.data.length}`);
      console.log();

      for (const server of resp.data) {
        console.log(`  ${server.name} (ID: ${server.server_id})`);
        console.log(`    PSM:        ${server.psm}`);
        if (server.description) console.log(`    描述:       ${server.description}`);
        console.log(`    负责人:     ${server.owner}`);
        console.log(`    管理员:     ${server.admins.join(", ")}`);
        console.log(`    认证:       ${server.auth_enabled ? "启用" : "禁用"}`);
        console.log(`    版本:       ${server.current_revision_id}`);
        console.log(`    更新时间:   ${formatServerDate(server.updated_at)}`);
        console.log();
      }
    },
  );

mcpCmd
  .command("call <server-id> <tool-name>")
  .description("调用 MCP Server 的工具")
  .option("-a, --arg <arg...>", "工具参数，格式: key=value（可多次使用）")
  .option("-r, --region <region>", "区域", "cn")
  .option("--json", "JSON 格式输出")
  .action(
    async (
      serverId: string,
      toolName: string,
      opts: { arg?: string[]; region: string; json?: boolean },
    ) => {
      const arguments_: Record<string, unknown> = {};
      for (const a of opts.arg ?? []) {
        const eqIdx = a.indexOf("=");
        if (eqIdx === -1) {
          console.error(`参数格式错误: ${a}，应为 key=value`);
          return;
        }
        const key = a.slice(0, eqIdx);
        const value = a.slice(eqIdx + 1);
        try {
          arguments_[key] = JSON.parse(value);
        } catch {
          arguments_[key] = value;
        }
      }

      const resp = await api.callMcpTool(
        serverId,
        toolName,
        arguments_,
        opts.region as RegionValue,
      );

      if (opts.json) {
        console.log(JSON.stringify(resp, null, 2));
        return;
      }

      console.log(`Server ID: ${serverId}`);
      console.log(`工具: ${toolName}`);
      console.log(`区域: ${opts.region}`);
      console.log();

      if (resp.result.content.length > 0) {
        for (const content of resp.result.content) {
          if (content.type === "text") {
            try {
              const data = JSON.parse(content.text);
              console.log(JSON.stringify(data, null, 2));
            } catch {
              console.log(content.text);
            }
          }
        }
      } else {
        console.log("无返回内容");
      }
    },
  );

mcpCmd
  .command("tools <server-id>")
  .description("查询 MCP Server 的工具列表")
  .option("-r, --region <region>", "区域", "cn")
  .option("--json", "JSON 格式输出")
  .action(async (serverId: string, opts: { region: string; json?: boolean }) => {
    const resp = await api.getMcpTools(serverId, opts.region as RegionValue);

    if (opts.json) {
      console.log(JSON.stringify(resp, null, 2));
      return;
    }

    const tools = resp.result.tools;
    if (tools.length === 0) {
      console.log("未找到工具");
      return;
    }

    console.log(`Server ID: ${serverId}`);
    console.log(`区域: ${opts.region}`);
    console.log(`工具数量: ${tools.length}`);
    console.log();

    for (const tool of tools) {
      console.log(`  ${tool.name}`);
      if (tool.description) console.log(`    描述: ${tool.description}`);
      if (tool.inputSchema.properties) {
        console.log("    参数:");
        for (const [paramName, paramInfo] of Object.entries(tool.inputSchema.properties)) {
          const required = tool.inputSchema.required.includes(paramName) ? " [必填]" : "";
          const desc = paramInfo.description || paramInfo.type;
          console.log(`      - ${paramName}: ${desc}${required}`);
        }
      }
      console.log();
    }
  });
