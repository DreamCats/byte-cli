import { Command } from "commander";
import { authCmd } from "./auth/cli.js";
import { codebaseCmd } from "./codebase/cli.js";
import { logidCmd } from "./logid/cli.js";
import { psmCmd } from "./psm/cli.js";
import { iamCmd } from "./iam/cli.js";
import { mcpCmd } from "./mcp/cli.js";
import { AuthError } from "./auth/exceptions.js";

const VERSION = "0.2.0";

const program = new Command()
  .name("byte-cli")
  .version(VERSION)
  .description("字节内部开发工具统一 CLI");

program.addCommand(authCmd);
program.addCommand(codebaseCmd);
program.addCommand(logidCmd);
program.addCommand(psmCmd);
program.addCommand(iamCmd);
program.addCommand(mcpCmd);

program
  .command("version")
  .description("显示版本信息")
  .action(() => {
    console.log(`byte-cli ${VERSION}`);
  });

// Error handling
try {
  await program.parseAsync();
} catch (e: unknown) {
  if (e instanceof AuthError) {
    console.error(`认证错误: ${e.message}`);
    process.exit(2);
  }
  if (e instanceof Error) {
    console.error(`错误: ${e.message}`);
  } else {
    console.error(`错误: ${e}`);
  }
  process.exit(1);
}
