import { Command } from "commander";
import * as api from "./api.js";
import { inferRepoName } from "./git.js";
import type { Comment } from "./models.js";

export const codebaseCmd = new Command("codebase").description("Codebase 仓库和 MR 查询");

function resolveRepo(repo?: string): string {
  if (repo) return repo;
  const inferred = inferRepoName();
  if (!inferred) {
    console.error("无法推断仓库名，请在 git 仓库中运行或使用 -R 指定");
    process.exit(1);
  }
  console.log(`推断仓库: ${inferred}`);
  return inferred;
}

// repo subgroup
const repoCmd = codebaseCmd.command("repo").description("仓库操作");

repoCmd
  .command("info <repo-name>")
  .description("查看仓库信息，repo-name 格式: group/project")
  .option("--json", "JSON 格式输出")
  .action(async (repoName: string, opts: { json?: boolean }) => {
    try {
      const info = await api.getRepoInfo(repoName);
      if (opts.json) {
        console.log(JSON.stringify(info, null, 2));
      } else {
        console.log(`仓库: ${info.name}`);
        console.log(`ID: ${info.id}`);
        console.log(`描述: ${info.description || "(无)"}`);
        console.log(`类型: ${info.type}`);
        console.log(`状态: ${info.status}`);
        console.log(`Git SSH: ${info.git_ssh_url}`);
        console.log(`Git HTTP: ${info.git_http_url}`);
        console.log(`合并方式: ${info.merge_method}`);
        console.log(`创建时间: ${info.created_at}`);
      }
    } catch (e: unknown) {
      console.error(`错误: ${(e as Error).message}`);
      process.exit(1);
    }
  });

// mr subgroup
const mrCmd = codebaseCmd.command("mr").description("MR 操作");

mrCmd
  .command("get <number>")
  .description("查看 MR 详情")
  .option("-R, --repo <repo>", "仓库名 group/project，默认从 git remote 推断")
  .option("--json", "JSON 格式输出")
  .action(async (number: string, opts: { repo?: string; json?: boolean }) => {
    try {
      const repoName = resolveRepo(opts.repo);
      const repoInfo = await api.getRepoInfo(repoName);
      const mr = await api.getMr(repoInfo.id, parseInt(number, 10));

      if (opts.json) {
        console.log(JSON.stringify(mr, null, 2));
      } else {
        console.log(`MR !${mr.number}: ${mr.title}`);
        console.log(`状态: ${mr.status}`);
        console.log(`作者: ${mr.created_by.display_name.content} (@${mr.created_by.username})`);
        console.log(`源分支: ${mr.source_branch_name} → ${mr.target_branch_name}`);
        console.log(`变更: ${mr.changes_count} 文件, ${mr.commits_count} 提交`);
        if (mr.draft) console.log("Draft: 是");
        console.log(`创建时间: ${mr.created_at}`);
        if (mr.description) console.log(`\n描述:\n${mr.description}`);
      }
    } catch (e: unknown) {
      console.error(`错误: ${(e as Error).message}`);
      process.exit(1);
    }
  });

mrCmd
  .command("comments <number>")
  .description("查看 MR 评论")
  .option("-R, --repo <repo>", "仓库名 group/project，默认从 git remote 推断")
  .option("--unresolved", "只显示未解决的评论")
  .option("--json", "JSON 格式输出")
  .action(
    async (
      number: string,
      opts: { repo?: string; unresolved?: boolean; json?: boolean },
    ) => {
      try {
        const repoName = resolveRepo(opts.repo);
        const repoInfo = await api.getRepoInfo(repoName);
        const mr = await api.getMr(repoInfo.id, parseInt(number, 10));
        const threads = await api.getMrThreads(repoInfo.id, mr.id);

        if (opts.json) {
          console.log(JSON.stringify(threads, null, 2));
          return;
        }

        for (const thread of threads) {
          if (opts.unresolved && thread.status === "resolved") continue;

          const icon = thread.status === "unresolved" ? "●" : "●";
          const outdated = thread.outdated ? " [outdated]" : "";

          if (thread.positions.length > 0) {
            const pos = thread.positions[0];
            console.log(`\n${icon} ${pos.path}:${pos.start_line}${outdated}`);
          } else {
            console.log(`\n${icon} (general)${outdated}`);
          }

          for (const comment of thread.comments) {
            printComment(comment);
          }
        }
      } catch (e: unknown) {
        console.error(`错误: ${(e as Error).message}`);
        process.exit(1);
      }
    },
  );

function printComment(comment: Comment): void {
  const author = comment.created_by.display_name.content || comment.created_by.username;
  const header = `  [${author}] ${comment.created_at}`;

  if (comment.created_by.username === "aime") {
    printAimeComment(comment, header);
  } else {
    console.log(header);
    let preview = comment.content.slice(0, 200);
    if (comment.content.length > 200) preview += "...";
    console.log(`  ${preview}\n`);
  }
}

function printAimeComment(comment: Comment, header: string): void {
  console.log(header);

  let description = "";
  let priority = "";
  let category = "";

  for (const line of comment.content.split("\n")) {
    const trimmed = line.trim();
    if (trimmed.startsWith("**问题描述**:")) {
      description = trimmed.split(":", 1)[1]?.trim() ?? "";
    } else if (trimmed.startsWith("**优先级**:")) {
      priority = trimmed.split(":", 1)[1]?.trim() ?? "";
    } else if (trimmed.startsWith("**问题分类**:")) {
      category = trimmed.split(":", 1)[1]?.trim() ?? "";
    }
  }

  if (description) console.log(`  问题: ${description}`);
  if (priority) console.log(`  优先级: ${priority}`);
  if (category) console.log(`  分类: ${category}`);
  console.log();
}
