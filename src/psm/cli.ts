import { Command } from "commander";
import * as api from "./api.js";
import { generatePlatformLinks, getIdlRepoUrl, getIdlRepoMainIdlUrl, formatUpdatedAt } from "./models.js";

export const psmCmd = new Command("psm").description("PSM 服务信息查询");

psmCmd
  .command("idl <psm-name>")
  .description("查询 PSM 的 IDL 信息")
  .option("--json", "JSON 格式输出")
  .action(async (psmName: string, opts: { json?: boolean }) => {
    const info = await api.getIdlInfo(psmName);
    if (opts.json) {
      console.log(JSON.stringify(info, null, 2));
    } else {
      console.log(`PSM:            ${info.psm}`);
      console.log(`RepoName:       ${info.repo_name}`);
      console.log(`IDLPath:        ${info.idl_path}`);
      console.log(`DefaultBranch:  ${info.default_branch}`);
      console.log(`IDLVersion:     ${info.idl_version}`);
      console.log(`IDLRepoURL:     ${getIdlRepoUrl(info)}`);
      console.log(`IDLRepoMainIDLURL: ${getIdlRepoMainIdlUrl(info)}`);
    }
  });

psmCmd
  .command("api-list <psm-name>")
  .description("查询 PSM 的 API 接口列表")
  .option("--json", "JSON 格式输出")
  .action(async (psmName: string, opts: { json?: boolean }) => {
    const endpoints = await api.getApiList(psmName);
    if (opts.json) {
      console.log(JSON.stringify(endpoints, null, 2));
    } else {
      if (endpoints.length === 0) {
        console.log(`PSM ${psmName} 未找到 API 接口`);
        return;
      }
      console.log(`PSM: ${psmName}`);
      console.log(`接口数量: ${endpoints.length}`);
      console.log();
      for (const ep of endpoints) {
        console.log(`  ${ep.name}`);
        if (ep.note) console.log(`    说明:     ${ep.note}`);
        console.log(`    负责人:   ${ep.owner}`);
        console.log(`    版本:     ${ep.version}`);
        console.log(`    序列化:   ${ep.serializer}`);
        console.log(`    单向调用: ${ep.oneway ? "是" : "否"}`);
        console.log(`    更新时间: ${formatUpdatedAt(ep)}`);
        console.log();
      }
    }
  });

psmCmd
  .command("links <psm-name>")
  .description("生成 PSM 的各平台链接")
  .option("-r, --region <region>", "区域", "us")
  .option("--json", "JSON 格式输出")
  .action((psmName: string, opts: { region: string; json?: boolean }) => {
    const links = generatePlatformLinks(psmName, opts.region);
    if (opts.json) {
      console.log(JSON.stringify(links, null, 2));
    } else {
      console.log(`PSM: ${psmName}`);
      console.log(`区域: ${opts.region}`);
      console.log();
      console.log(`TCE:      ${links.tceUrl}`);
      console.log(`SCM:      ${links.scmUrl}`);
      console.log(`TCC:      ${links.tccUrl}`);
      console.log(`Overpass: ${links.overpassUrl}`);
    }
  });
