#!/usr/bin/env python3
import json
import os
import sys

import akshare as ak
import pandas as pd


def normalize_text(value):
    if value is None:
        return ""
    if isinstance(value, float) and pd.isna(value):
        return ""
    text = str(value).strip()
    if text.lower() in {"none", "nan"}:
        return ""
    return text


def main():
    symbol = normalize_text(os.environ.get("AKSHARE_PROFILE_SYMBOL"))
    if not symbol:
        raise ValueError("AKSHARE_PROFILE_SYMBOL is required")
    normalized = symbol.split(".")[0]
    df = ak.stock_profile_cninfo(symbol=normalized)
    if df is None or df.empty:
        print("{}")
        return
    row = df.iloc[0].to_dict()
    payload = {
        "company_name": normalize_text(row.get("公司名称")),
        "english_name": normalize_text(row.get("英文名称")),
        "market_label": normalize_text(row.get("所属市场")),
        "industry_label": normalize_text(row.get("所属行业")),
        "legal_representative": normalize_text(row.get("法人代表")),
        "registered_capital": normalize_text(row.get("注册资金")),
        "founded_at": normalize_text(row.get("成立日期")),
        "listed_at": normalize_text(row.get("上市日期")),
        "website": normalize_text(row.get("官方网站")),
        "email": normalize_text(row.get("电子邮箱")),
        "phone": normalize_text(row.get("联系电话")),
        "address": normalize_text(row.get("注册地址")),
        "office_address": normalize_text(row.get("办公地址")),
        "business": normalize_text(row.get("主营业务")),
        "business_scope": normalize_text(row.get("经营范围")),
        "introduction": normalize_text(row.get("机构简介")),
        "source": "akshare_cninfo_profile",
    }
    print(json.dumps(payload, ensure_ascii=False))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        sys.exit(1)
