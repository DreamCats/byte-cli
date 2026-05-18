import { z } from "zod";

export const OwnerInfoSchema = z.object({
  id: z.number().default(0),
  username: z.string().default(""),
  status: z.number().default(0),
  created_at: z.number().default(0),
  updated_at: z.number().default(0),
});

export const SecretItemSchema = z.object({
  uid: z.string().default(""),
  secret: z.string().default(""),
  enabled: z.boolean().default(true),
  created_by: z.string().default(""),
  created_at: z.number().default(0),
});

export const ServiceAccountSpecSchema = z.object({
  name: z.string().default(""),
  node_id: z.number().default(0),
  path: z.string().default(""),
  description: z.string().default(""),
  secret: z.string().default(""),
  owners: z.array(z.string()).default([]),
  created_by: z.string().default(""),
  updated_by: z.string().default(""),
  created_at: z.number().default(0),
  updated_at: z.number().default(0),
  i18n_name: z.string().default(""),
  i18n_path: z.string().default(""),
  sensitive_type: z.number().default(0),
  owner_list: z.array(OwnerInfoSchema).default([]),
  secrets: z.array(SecretItemSchema).default([]),
});
export type ServiceAccountSpec = z.infer<typeof ServiceAccountSpecSchema>;

export const ServiceAccountSchema = z.object({
  name: z.string().default(""),
  id: z.number().default(0),
  spec: ServiceAccountSpecSchema.default({}),
});
export type ServiceAccount = z.infer<typeof ServiceAccountSchema>;

export const PageInfoSchema = z.object({
  total: z.number().default(0),
  page: z.number().default(1),
  page_size: z.number().default(10),
});

export const ServiceAccountListResponseSchema = z.object({
  error_code: z.number().default(0),
  data: z.array(ServiceAccountSchema).default([]),
  page_info: PageInfoSchema.default({}),
});
export type ServiceAccountListResponse = z.infer<typeof ServiceAccountListResponseSchema>;

export function formatTimestamp(ts: number): string {
  if (!ts) return "";
  return new Date(ts).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
