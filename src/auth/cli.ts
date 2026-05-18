import { Command } from "commander";
import { AuthManager } from "./manager.js";
import { parseRegion, allRegions } from "./regions.js";
import { AuthError } from "./exceptions.js";
import * as config from "../common/config.js";

export const authCmd = new Command("auth").description("认证管理");

authCmd
  .command("login")
  .description("获取并缓存 JWT Token")
  .option("-r, --region <region>", "区域: cn/i18n/us/eu/codebase", "cn")
  .action(async (opts: { region: string }) => {
    try {
      const r = parseRegion(opts.region);
      const manager = new AuthManager(r);
      const token = await manager.getToken(true);
      console.log(`区域 ${r.value} 认证成功`);
      console.log(`Token: ${token.slice(0, 20)}...`);
    } catch (e) {
      if (e instanceof AuthError) {
        console.error(`认证错误: ${e.message}`);
        process.exit(2);
      }
      throw e;
    }
  });

authCmd
  .command("status")
  .description("查看各区域认证状态")
  .action(() => {
    const cfg = config.load();
    for (const r of allRegions()) {
      const hasCookie = !!config.getCookie(cfg, r.value);
      const manager = new AuthManager(r);
      const tokenValid = hasCookie ? manager.isTokenValid() : false;
      const cookieStatus = hasCookie ? "✓" : "✗";
      const tokenStatus = tokenValid ? "✓" : "✗";
      console.log(
        `  ${r.value.padStart(10)}  Cookie: ${cookieStatus}  Token: ${tokenStatus}`,
      );
    }
  });

authCmd
  .command("token")
  .description("输出指定区域的 JWT Token")
  .option("-r, --region <region>", "区域: cn/i18n/us/eu/codebase", "cn")
  .action(async (opts: { region: string }) => {
    try {
      const r = parseRegion(opts.region);
      const manager = new AuthManager(r);
      const t = await manager.getToken();
      console.log(t);
    } catch (e) {
      if (e instanceof AuthError) {
        console.error(`认证错误: ${e.message}`);
        process.exit(2);
      }
      throw e;
    }
  });

const configCmd = authCmd.command("config").description("配置管理");

configCmd
  .command("show")
  .description("查看当前配置")
  .action(() => {
    const cfg = config.load();
    console.log("区域 Cookie:");
    for (const r of allRegions()) {
      const cookie = config.getCookie(cfg, r.value);
      if (cookie) {
        console.log(`  ${r.value}: ${config.maskCookie(cookie)}`);
      } else {
        console.log(`  ${r.value}: (未配置)`);
      }
    }
    if (cfg.proxy.https || cfg.proxy.http) {
      console.log("\n代理:");
      if (cfg.proxy.https) console.log(`  HTTPS: ${cfg.proxy.https}`);
      if (cfg.proxy.http) console.log(`  HTTP: ${cfg.proxy.http}`);
    }
  });

configCmd
  .command("set-cookie <cookie>")
  .description("设置指定区域的 Cookie")
  .option("-r, --region <region>", "区域: cn/i18n/us/eu/codebase", "cn")
  .action((cookie: string, opts: { region: string }) => {
    const r = parseRegion(opts.region);
    const cfg = config.load();
    config.setCookie(cfg, r.value, cookie);
    config.save(cfg);
    console.log(`区域 ${r.value} Cookie 已更新`);
  });
