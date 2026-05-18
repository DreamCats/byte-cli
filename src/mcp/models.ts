import { z } from "zod";

export const MCPServerSchema = z.object({
  server_id: z.string(),
  psm: z.string(),
  env_name: z.string(),
  name: z.string(),
  description: z.string().default(""),
  owner: z.string(),
  subscribers: z.array(z.string()).default([]),
  current_revision_id: z.string().default(""),
  auth_enabled: z.boolean().default(true),
  allowed_psms: z.array(z.string()).default([]),
  admins: z.array(z.string()).default([]),
  created_at: z.string(),
  updated_at: z.string(),
});
export type MCPServer = z.infer<typeof MCPServerSchema>;

export const MCPServerListResponseSchema = z.object({
  code: z.number().default(0),
  error: z.string().default(""),
  data: z.array(MCPServerSchema).default([]),
});
export type MCPServerListResponse = z.infer<typeof MCPServerListResponseSchema>;

export const ToolParameterSchema: z.ZodType<ToolParameter> = z.lazy(() =>
  z.object({
    type: z.string(),
    description: z.string(),
    format: z.string(),
    properties: z.record(ToolParameterSchema).nullable(),
    items: ToolParameterSchema.nullable(),
    required: z.array(z.string()),
  }),
) as z.ZodType<ToolParameter>;
export type ToolParameter = {
  type: string;
  description: string;
  format: string;
  properties: Record<string, ToolParameter> | null;
  items: ToolParameter | null;
  required: string[];
};

export const ToolSchema = z.object({
  name: z.string(),
  description: z.string().default(""),
  inputSchema: ToolParameterSchema.default({ type: "object", description: "", format: "", properties: null, items: null, required: [] }),
  annotations: z.record(z.string()).default({}),
});
export type Tool = z.infer<typeof ToolSchema>;

export const ToolListResultSchema = z.object({
  tools: z.array(ToolSchema).default([]),
});

export const ToolListResponseSchema = z.object({
  jsonrpc: z.string().default("2.0"),
  id: z.number().default(1),
  result: ToolListResultSchema.default({ tools: [] }),
});
export type ToolListResponse = z.infer<typeof ToolListResponseSchema>;

export const ToolCallContentSchema = z.object({
  type: z.string().default("text"),
  text: z.string().default(""),
});

export const ToolCallMetaSchema = z.object({
  trace: z.record(z.string()).default({}),
});

export const ToolCallResultSchema = z.object({
  content: z.array(ToolCallContentSchema).default([]),
  meta: ToolCallMetaSchema.default({}),
});

export const ToolCallResponseSchema = z.object({
  jsonrpc: z.string().default("2.0"),
  id: z.number().default(1),
  result: ToolCallResultSchema.default({}),
});
export type ToolCallResponse = z.infer<typeof ToolCallResponseSchema>;

export function formatServerDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
