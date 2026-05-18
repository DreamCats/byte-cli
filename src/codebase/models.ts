import { z } from "zod";

// PascalCase → snake_case converter
function toSnake(name: string): string {
  return name.replace(/([A-Z])/g, (_, ch, i) => (i > 0 ? "_" : "") + ch.toLowerCase());
}

function pascalToSnake(obj: unknown): unknown {
  if (Array.isArray(obj)) return obj.map(pascalToSnake);
  if (obj && typeof obj === "object") {
    return Object.fromEntries(
      Object.entries(obj).map(([k, v]) => [toSnake(k), pascalToSnake(v)]),
    );
  }
  return obj;
}

export const DepartmentSchema = z.object({
  id: z.number().default(0),
  name: z.string().default(""),
  en_name: z.string().default(""),
  tenant_key: z.string().default(""),
});
export type Department = z.infer<typeof DepartmentSchema>;

export const RepoSchema = z.object({
  id: z.number().default(0),
  name: z.string().default(""),
  platform: z.string().default(""),
  external_id: z.number().default(0),
  external_url: z.string().default(""),
  git_url: z.string().default(""),
  git_ssh_url: z.string().default(""),
  git_http_url: z.string().default(""),
  type: z.string().default(""),
  level: z.string().default(""),
  status: z.string().default(""),
  description: z.string().default(""),
  is_monorepo: z.boolean().default(false),
  merge_method: z.string().default(""),
  squash: z.string().default(""),
  department: DepartmentSchema.default({}),
  created_at: z.string().default(""),
  audit_status: z.string().default(""),
});
export type Repo = z.infer<typeof RepoSchema>;

export const DisplayNameSchema = z.object({
  content: z.string().default(""),
  i18n: z.string().default(""),
});

export const UserSchema = z.object({
  id: z.number().default(0),
  username: z.string().default(""),
  display_name: DisplayNameSchema.default({}),
  email: z.string().default(""),
});
export type User = z.infer<typeof UserSchema>;

export const CurrentUserStateSchema = z.object({
  approved: z.boolean().default(false),
  can_approve: z.boolean().default(false),
  can_merge: z.boolean().default(false),
  can_update: z.boolean().default(false),
  can_bypass: z.boolean().default(false),
});

export const LinkSchema = z.object({
  url: z.string().default(""),
  text: z.string().default(""),
});

export const MRSchema = z.object({
  id: z.number().default(0),
  number: z.number().default(0),
  status: z.string().default(""),
  source_branch_name: z.string().default(""),
  target_branch_name: z.string().default(""),
  title: z.string().default(""),
  description: z.string().default(""),
  created_by: UserSchema.default({}),
  created_at: z.string().default(""),
  updated_at: z.string().default(""),
  changes_count: z.number().default(0),
  commits_count: z.number().default(0),
  merge_method: z.string().default(""),
  draft: z.boolean().default(false),
  squash_commits: z.boolean().default(false),
  auto_merge: z.boolean().default(false),
  merge_in_progress: z.boolean().default(false),
  links: z.array(LinkSchema).default([]),
  current_user: CurrentUserStateSchema.default({}),
});
export type MR = z.infer<typeof MRSchema>;

export const PositionSchema = z.object({
  type: z.string().default(""),
  side: z.string().default(""),
  path: z.string().default(""),
  start_line: z.number().default(0),
  end_line: z.number().default(0),
});
export type Position = z.infer<typeof PositionSchema>;

export const SuggestionSchema = z.object({
  content: z.string().default(""),
  original_content: z.string().default(""),
  applied: z.boolean().default(false),
});

export const CommentSchema = z.object({
  id: z.number().default(0),
  content: z.string().default(""),
  created_at: z.string().default(""),
  updated_at: z.string().default(""),
  created_by: UserSchema.default({}),
  suggestions: z.array(SuggestionSchema).default([]),
});
export type Comment = z.infer<typeof CommentSchema>;

export const ThreadSchema = z.object({
  id: z.number().default(0),
  status: z.string().default(""),
  outdated: z.boolean().default(false),
  comments: z.array(CommentSchema).default([]),
  positions: z.array(PositionSchema).default([]),
});
export type Thread = z.infer<typeof ThreadSchema>;

export function parseRepo(data: unknown): Repo {
  return RepoSchema.parse(pascalToSnake(data));
}

export function parseMR(data: unknown): MR {
  return MRSchema.parse(pascalToSnake(data));
}

export function parseThread(data: unknown): Thread {
  return ThreadSchema.parse(pascalToSnake(data));
}
