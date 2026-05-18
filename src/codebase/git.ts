import { execSync } from "node:child_process";
import { URL } from "node:url";

export function inferRepoName(): string | null {
  try {
    const url = execSync("git remote get-url origin", {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    return parseRepoName(url);
  } catch {
    return null;
  }
}

export function parseRepoName(url: string): string | null {
  // ssh://git@host:port/group/project.git
  if (url.startsWith("ssh://")) {
    const parsed = new URL(url);
    return cleanPath(parsed.pathname);
  }

  // git@host:group/project.git
  if (url.includes("@") && url.includes(":")) {
    const idx = url.lastIndexOf(":");
    return cleanPath(url.slice(idx + 1));
  }

  // https://host/group/project.git
  if (url.startsWith("http")) {
    const parsed = new URL(url);
    return cleanPath(parsed.pathname);
  }

  return null;
}

function cleanPath(path: string): string | null {
  path = path.replace(/^\/+/, "").replace(/\/+$/, "");
  if (path.endsWith(".git")) path = path.slice(0, -4);
  if (!path.includes("/")) return null;
  return path;
}
