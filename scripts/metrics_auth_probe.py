#!/usr/bin/env python3
"""Probe whether the current byte-cli US JWT works for the metrics tagk API."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path


DEFAULT_METRIC = "ttec.industry.solution.sales_mask_rule_match"
DEFAULT_ENDPOINT = "https://metrics-svc-platform-ttp.tiktok-us.org/byteplot/api/v2/suggest/tagk"


def repo_root() -> Path:
    return Path(__file__).resolve().parents[1]


def get_us_token(byte_cli: str) -> str:
    if byte_cli:
        cmd = [byte_cli, "auth", "token", "--region", "us"]
    else:
        cmd = ["go", "run", "./cmd/byte-cli", "auth", "token", "--region", "us"]

    proc = subprocess.run(
        cmd,
        cwd=repo_root(),
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise RuntimeError(f"failed to get US token: {detail}")

    token = proc.stdout.strip()
    if not token:
        raise RuntimeError("byte-cli returned an empty US token")
    return token


def request_tagk(args: argparse.Namespace, token: str) -> tuple[int, dict[str, str], bytes]:
    query = urllib.parse.urlencode(
        {
            "_region": args.metrics_region,
            "metric_name": args.metric_name,
            "_tenant": args.tenant,
        }
    )
    url = f"{args.endpoint}?{query}"
    req = urllib.request.Request(
        url,
        headers={
            "Accept": "application/json, text/plain, */*",
            "Authorization": token,
            "Connection": "keep-alive",
            "Origin": "https://metrics-fe-ttp-us.tiktok-row.org",
            "Referer": "https://metrics-fe-ttp-us.tiktok-row.org/",
            "Sec-Fetch-Dest": "empty",
            "Sec-Fetch-Mode": "cors",
            "Sec-Fetch-Site": "cross-site",
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",
            "accept-language": "zh",
            "sec-ch-ua": '"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"',
            "sec-ch-ua-mobile": "?0",
            "sec-ch-ua-platform": '"macOS"',
        },
        method="GET",
    )

    try:
        with urllib.request.urlopen(req, timeout=args.timeout) as resp:
            return resp.status, dict(resp.headers), resp.read()
    except urllib.error.HTTPError as exc:
        return exc.code, dict(exc.headers), exc.read()


def print_summary(status: int, headers: dict[str, str], body: bytes) -> int:
    print(f"HTTP {status}")
    content_type = header_value(headers, "content-type")
    if content_type:
        print(f"Content-Type: {content_type}")
    for key in ("x-tt-logid", "x-request-id", "x-bd-trace-id", "server"):
        value = header_value(headers, key)
        if value:
            print(f"{key}: {value}")

    text = body.decode("utf-8", errors="replace")
    parsed = None
    try:
        parsed = json.loads(text)
    except json.JSONDecodeError:
        pass

    if isinstance(parsed, dict):
        keys = ", ".join(sorted(parsed.keys()))
        print(f"JSON keys: {keys}")
        for key in ("code", "message", "msg", "error"):
            if key in parsed:
                print(f"{key}: {parsed[key]}")
        data = parsed.get("data")
        if isinstance(data, list):
            print(f"data items: {len(data)}")
            if data:
                print("first item:", json.dumps(data[0], ensure_ascii=False)[:500])
        elif isinstance(data, dict):
            print("data keys:", ", ".join(sorted(data.keys())))
    elif isinstance(parsed, list):
        print(f"JSON list items: {len(parsed)}")
        if parsed:
            print("first item:", json.dumps(parsed[0], ensure_ascii=False)[:500])
    else:
        preview = text.replace("\n", "\\n")[:1000]
        print(f"Body preview: {preview}")

    return 0 if 200 <= status < 300 else 1


def header_value(headers: dict[str, str], name: str) -> str:
    for key, value in headers.items():
        if key.lower() == name.lower():
            return value
    return ""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Test whether byte-cli's current US JWT can authorize metrics tagk requests."
    )
    parser.add_argument("--metric-name", default=DEFAULT_METRIC)
    parser.add_argument("--metrics-region", default="US-TTP")
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--endpoint", default=DEFAULT_ENDPOINT)
    parser.add_argument("--timeout", type=float, default=15)
    parser.add_argument(
        "--byte-cli",
        default="",
        help="Optional byte-cli binary path. Defaults to `go run ./cmd/byte-cli`.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        token = get_us_token(args.byte_cli)
        print("US token: loaded")
        status, headers, body = request_tagk(args, token)
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1
    return print_summary(status, headers, body)


if __name__ == "__main__":
    raise SystemExit(main())
