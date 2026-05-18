import { z } from "zod";

export const LogGroupSchema = z.object({
  psm: z.string().default(""),
  pod_name: z.string().default(""),
  ipv4: z.string().default(""),
  env: z.string().default(""),
  vregion: z.string().default(""),
  idc: z.string().default(""),
});

export const KVPairSchema = z.object({
  key: z.string().default(""),
  value: z.string().default(""),
});

export const LogValueSchema = z.object({
  kv_list: z.array(KVPairSchema).default([]),
  level: z.string().default(""),
});

export const LogItemSchema = z.object({
  id: z.string().default(""),
  group: LogGroupSchema.default({}),
  value: z.array(LogValueSchema).default([]),
});

export const LogMetaSchema = z.object({
  scan_time_range: z.string().default(""),
  level_list: z.array(z.string()).default([]),
});

export const LogDataSchema = z.object({
  items: z.array(LogItemSchema).default([]),
  meta: LogMetaSchema.default({}),
  tag_infos: z.array(z.record(z.unknown())).default([]),
});

export const LogQueryResponseSchema = z.object({
  data: LogDataSchema.default({}),
});

export const FlattenedLogEntrySchema = z.object({
  psm: z.string().default(""),
  pod_name: z.string().default(""),
  level: z.string().default(""),
  msg: z.string().default(""),
  location: z.string().default(""),
  env: z.string().default(""),
  vregion: z.string().default(""),
});

export type FlattenedLogEntry = z.infer<typeof FlattenedLogEntrySchema>;
export type LogItem = z.infer<typeof LogItemSchema>;
