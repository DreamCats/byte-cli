import { AuthManager } from "../auth/manager.js";
import { parseRegion } from "../auth/regions.js";
import * as http from "../common/http.js";
import { parseRepo, parseMR, parseThread, type Repo, type MR, type Thread } from "./models.js";

const REPO_API_BASE = "https://codebase-api.byted.org/unstable";
const MR_API_BASE = "https://code.byted.org/api/v2/";

async function getToken(): Promise<string> {
  const manager = new AuthManager(parseRegion("codebase"));
  return manager.getToken();
}

export async function getRepoInfo(repo: string): Promise<Repo> {
  const token = await getToken();
  const encoded = encodeURIComponent(repo);
  const resp = await http.get(`${REPO_API_BASE}/repos/${encoded}`, {
    Authorization: `Codebase-User-JWT ${token}`,
    "Accept-Encoding": "gzip, deflate, br, zstd",
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return parseRepo(await resp.json());
}

export async function getMr(repoId: number, number: number): Promise<MR> {
  const token = await getToken();
  const body = {
    RepoId: repoId,
    Number: number,
    Selector: { Labels: true, CurrentUser: true, Version: true, User: true },
  };
  const resp = await http.post(
    `${MR_API_BASE}?Action=GetMergeRequest`,
    body,
    {
      "x-codebase-user-jwt": token,
    },
  );
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  return parseMR(await resp.json());
}

export async function getMrThreads(repoId: number, mrId: number): Promise<Thread[]> {
  const token = await getToken();
  const body = {
    RepoId: repoId,
    CommentableId: mrId,
    CommentableType: "merge_request",
  };
  const resp = await http.post(
    `${MR_API_BASE}?Action=ListThreads`,
    body,
    {
      "x-codebase-user-jwt": token,
    },
  );
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  const data = (await resp.json()) as Record<string, unknown>;
  const threads = (data.Threads ?? data.threads ?? []) as unknown[];
  return threads.map(parseThread);
}
