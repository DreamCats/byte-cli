const DEFAULT_FILTERS = [
  /_compliance_nlp_log/,
  /_compliance_whitelist_log/,
  /_compliance_source=footprint/,
  /"user_extra":\s*"[\s\S]*?"/,
  /"LogID":\s*"[^"]*"/gm,
  /"Addr":\s*"[^"]*"/gm,
  /"Client":\s*"[^"]*"/gm,
  /\{\{rip=[^}]*\}\}/,
];

export class MessageSanitizer {
  private patterns: RegExp[];

  constructor(patterns?: RegExp[]) {
    this.patterns = patterns ?? DEFAULT_FILTERS;
  }

  sanitize(text: string): string {
    let result = text;
    for (const pattern of this.patterns) {
      result = result.replace(pattern, "");
    }
    return cleanWhitespace(result);
  }
}

export class KeywordFilter {
  private keywords: string[];

  constructor(keywords: string[]) {
    this.keywords = keywords.map((k) => k.toLowerCase());
  }

  matches(text: string): boolean {
    if (this.keywords.length === 0) return true;
    const lower = text.toLowerCase();
    return this.keywords.some((kw) => lower.includes(kw));
  }
}

function cleanWhitespace(text: string): string {
  return text
    .replace(/[ \t]{2,}/g, " ")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}
