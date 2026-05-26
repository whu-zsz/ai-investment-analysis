import json
import math
import os
import re
import sys
import time

import pandas as pd
import requests
from akshare.utils import demjson


HEADERS = {"User-Agent": "Mozilla/5.0"}
TIMEOUT = 20
PAGE_DELAY = float(os.getenv("THS_PAGE_DELAY", "1.2") or "1.2")
BOARD_DELAY = float(os.getenv("THS_BOARD_DELAY", "2.0") or "2.0")
RETRY_BASE_DELAY = float(os.getenv("THS_RETRY_BASE_DELAY", "2.0") or "2.0")
SECTOR_SPOT_URL = "http://money.finance.sina.com.cn/q/view/newFLJK.php"
SECTOR_COUNT_URL = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeStockCount"
SECTOR_DETAIL_URL = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData"
SESSION = requests.Session()


def normalize_symbol(raw_symbol: str) -> str:
    text = str(raw_symbol).strip().upper()
    if re.fullmatch(r"(SH|SZ|BJ)\d{6}", text):
        text = text[2:]
    if re.fullmatch(r"\d{6}\.(SH|SZ|BJ)", text):
        return text
    text = re.sub(r"\D", "", text)
    if len(text) > 6:
        return ""
    text = text.zfill(6)
    if text.startswith(("6", "9")):
        return f"{text}.SH"
    if text.startswith(("0", "2", "3")):
        return f"{text}.SZ"
    if text.startswith(("4", "8")):
        return f"{text}.BJ"
    return ""


def to_float(value) -> float:
    try:
        if value is None:
            return 0.0
        text = str(value).strip().replace(",", "").replace("%", "")
        if not text or text in {"--", "nan", "None"}:
            return 0.0
        return float(text)
    except Exception:
        return 0.0


def fetch_text(url: str, params=None) -> str:
    last_error = None
    for attempt in range(1, 4):
        try:
            response = SESSION.get(url, params=params, headers=HEADERS, timeout=TIMEOUT)
            response.raise_for_status()
            response.encoding = response.apparent_encoding or response.encoding
            return response.text
        except Exception as exc:
            last_error = exc
            if attempt < 3:
                time.sleep(RETRY_BASE_DELAY * attempt)
    raise last_error


def fetch_sector_spot_rows(board_type: str):
    indicator = "industry" if board_type == "industry" else "class"
    text = fetch_text(SECTOR_SPOT_URL, {"param": indicator})
    start = text.find("{")
    if start < 0:
        raise ValueError(f"invalid sector spot payload for {board_type}")
    payload = json.loads(text[start:])
    rows = []
    for value in payload.values():
        parts = str(value).split(",")
        if len(parts) < 2:
            continue
        label = parts[0].strip()
        name = parts[1].strip()
        if not label or not name:
            continue
        rows.append({"code": label, "name": name})
    return rows


def fetch_sector_total_count(label: str) -> int:
    text = fetch_text(SECTOR_COUNT_URL, {"node": label}).strip()
    return int(json.loads(text))


def fetch_sector_detail_page(label: str, page: int):
    params = {
        "page": str(page),
        "num": "80",
        "sort": "symbol",
        "asc": "1",
        "node": label,
        "symbol": "",
        "_s_r_a": "page",
    }
    text = fetch_text(SECTOR_DETAIL_URL, params)
    payload = demjson.decode(text)
    if not payload:
        return pd.DataFrame()
    return pd.DataFrame(payload)


def extract_constituents(frame: pd.DataFrame):
    constituents = []
    seen = set()
    for _, row in frame.iterrows():
        symbol = normalize_symbol(row.get("code") or row.get("symbol") or "")
        if not symbol or symbol in seen:
            continue
        seen.add(symbol)
        stock_name = str(row.get("name", "")).strip()
        constituents.append(
            {
                "symbol": symbol,
                "name": stock_name,
                "total_market_cap": to_float(row.get("mktcap")),
                "float_market_cap": to_float(row.get("nmc")),
            }
        )
    return constituents


def fetch_sector_constituents(label: str):
    total_count = fetch_sector_total_count(label)
    if total_count <= 0:
        return []
    total_pages = max(1, math.ceil(total_count / 80))
    constituents = []
    for page in range(1, total_pages + 1):
        frame = fetch_sector_detail_page(label, page)
        if frame.empty:
            break
        constituents.extend(extract_constituents(frame))
        if page < total_pages:
            time.sleep(PAGE_DELAY)
    unique = {}
    for item in constituents:
        unique[item["symbol"]] = item
    return list(unique.values())


def build_payload():
    payload = []
    requested_types = [item.strip() for item in os.getenv("THS_BOARD_TYPES", "industry,concept").split(",") if item.strip()]
    board_limit = int(os.getenv("THS_BOARD_LIMIT", "0") or "0")
    for board_type in requested_types:
        rows = fetch_sector_spot_rows(board_type)
        if board_limit > 0:
            rows = rows[:board_limit]
        for index, row in enumerate(rows):
            code = str(row.get("code", "")).strip()
            name = str(row.get("name", "")).strip()
            if not code or not name:
                continue
            try:
                constituents = fetch_sector_constituents(code)
            except Exception as exc:
                print(f"skip {board_type} {code} {name}: {exc}", file=sys.stderr)
                continue
            if not constituents:
                continue
            payload.append(
                {
                    "board_type": board_type,
                    "code": code,
                    "name": name,
                    "source": "sina_sector",
                    "company_count": len(constituents),
                    "constituents": constituents,
                }
            )
            if index < len(rows) - 1:
                time.sleep(BOARD_DELAY)
    return payload


def main():
    last_error = None
    for attempt in range(1, 3):
        try:
            payload = build_payload()
            if not payload:
                raise ValueError("empty board payload")
            sys.stdout.write(json.dumps(payload, ensure_ascii=False))
            return
        except Exception as exc:
            last_error = exc
            if attempt < 2:
                time.sleep(2 * attempt)
    raise last_error


if __name__ == "__main__":
    main()
