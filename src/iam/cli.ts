import { Command } from "commander";
import * as api from "./api.js";
import { formatTimestamp } from "./models.js";
import type { RegionValue } from "../auth/regions.js";

export const iamCmd = new Command("iam").description("IAM 服务账号查询");

iamCmd
  .command("list")
  .description("查询服务账号列表")
  .option("-o, --owner <owner>", "所有者用户名（默认根据 token 自动识别）")
  .option("-r, --region <region>", "区域", "cn")
  .option("-p, --page <page>", "页码", "1")
  .option("-s, --size <size>", "每页数量", "10")
  .option("--json", "JSON 格式输出")
  .action(
    async (opts: {
      owner?: string;
      region: string;
      page: string;
      size: string;
      json?: boolean;
    }) => {
      const resp = await api.getServiceAccounts(
        opts.owner,
        opts.region as RegionValue,
        parseInt(opts.page, 10),
        parseInt(opts.size, 10),
      );

      if (opts.json) {
        console.log(JSON.stringify(resp, null, 2));
        return;
      }

      if (resp.data.length === 0) {
        console.log(`未找到 ${opts.owner ?? "当前用户"} 的服务账号`);
        return;
      }

      if (opts.owner) console.log(`所有者: ${opts.owner}`);
      console.log(`区域: ${opts.region}`);
      console.log(`总数: ${resp.page_info.total}`);
      console.log();

      for (const sa of resp.data) {
        const spec = sa.spec;
        console.log(`  ${sa.name} (ID: ${sa.id})`);
        if (spec.description) console.log(`    描述:     ${spec.description}`);
        console.log(`    路径:     ${spec.path}`);
        console.log(`    负责人:   ${spec.owners.join(", ")}`);
        console.log(`    创建人:   ${spec.created_by}`);
        console.log(`    创建时间: ${formatTimestamp(spec.created_at)}`);
        console.log(`    更新时间: ${formatTimestamp(spec.updated_at)}`);
        console.log();
      }
    },
  );

iamCmd
  .command("secret <name>")
  .description("查询服务账号密钥")
  .option("-r, --region <region>", "区域", "cn")
  .option("--json", "JSON 格式输出")
  .action(async (name: string, opts: { region: string; json?: boolean }) => {
    const sa = await api.getServiceAccountSecret(name, opts.region as RegionValue);

    if (opts.json) {
      console.log(JSON.stringify(sa, null, 2));
      return;
    }

    const spec = sa.spec;
    console.log(`服务账号: ${sa.name}`);
    console.log(`Secret:   ${spec.secret}`);
    if (spec.secrets.length > 0) {
      console.log(`密钥数量: ${spec.secrets.length}`);
      for (const s of spec.secrets) {
        const status = s.enabled ? "启用" : "禁用";
        console.log(`  ${s.uid}  ${s.secret}  [${status}]`);
      }
    }
  });
