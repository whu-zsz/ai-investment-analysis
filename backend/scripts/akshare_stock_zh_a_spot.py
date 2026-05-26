import json
import math
import re
import sys
import time

import akshare as ak


def normalize_symbol(code: str) -> str:
    code = str(code).strip()
    lower = code.lower()
    if lower.startswith(("sh", "sz", "bj")):
        match = re.search(r"(sh|sz|bj)(\d+)", lower)
        if match:
            return f"{match.group(2)}.{match.group(1).upper()}"
    digits = "".join(ch for ch in code if ch.isdigit())
    if digits:
        code = digits
    if code.startswith(("43", "83", "87", "88", "92")):
        return f"{code}.BJ"
    if code.startswith(("60", "68", "90", "51", "56", "58", "11")):
        return f"{code}.SH"
    return f"{code}.SZ"


def to_float(value) -> float:
    if value is None:
        return 0.0
    if isinstance(value, float):
        if math.isnan(value):
            return 0.0
        return value
    if isinstance(value, int):
        return float(value)
    text = str(value).strip().replace(",", "")
    if not text or text == "-":
        return 0.0
    try:
        number = float(text)
    except ValueError:
        return 0.0
    if math.isnan(number):
        return 0.0
    return number


def main() -> int:
    last_error = None
    df = None
    for attempt in range(3):
        try:
            df = ak.stock_zh_a_spot()
            last_error = None
            break
        except Exception as exc:
            last_error = exc
            if attempt < 2:
                time.sleep(2 * (attempt + 1))
    if last_error is not None:
        raise last_error
    if df is None or df.empty:
        print("[]")
        return 0

    rename_map = {
        "代码": "code",
        "名称": "name",
        "最新价": "last_price",
        "涨跌额": "change_amount",
        "涨跌幅": "change_percent",
        "今开": "open_price",
        "最高": "high_price",
        "最低": "low_price",
        "昨收": "prev_close",
        "成交量": "volume",
        "成交额": "turnover",
    }
    df = df.rename(columns=rename_map)

    rows = []
    for _, row in df.iterrows():
        code = str(row.get("code", "")).strip()
        if not code:
            continue
        rows.append(
            {
                "symbol": normalize_symbol(code),
                "name": str(row.get("name", "")).strip(),
                "market": "cn_stock",
                "last_price": to_float(row.get("last_price")),
                "change_amount": to_float(row.get("change_amount")),
                "change_percent": to_float(row.get("change_percent")),
                "open_price": to_float(row.get("open_price")),
                "high_price": to_float(row.get("high_price")),
                "low_price": to_float(row.get("low_price")),
                "prev_close": to_float(row.get("prev_close")),
                "volume": to_float(row.get("volume")),
                "turnover": to_float(row.get("turnover")),
                "source": "akshare_sina",
            }
        )

    json.dump(rows, sys.stdout, ensure_ascii=False)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
