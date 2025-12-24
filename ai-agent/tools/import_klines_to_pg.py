import json
import os
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional, Tuple

import httpx
import psycopg


@dataclass(frozen=True)
class Candle:
    open_time_ms: int
    open: float
    high: float
    low: float
    close: float
    volume: Optional[float]
    raw: Dict[str, Any]

    @property
    def open_time_utc(self) -> datetime:
        return datetime.fromtimestamp(self.open_time_ms / 1000, tz=timezone.utc)


def _must_env(key: str) -> str:
    v = os.getenv(key, "").strip()
    if not v:
        raise RuntimeError(f"missing env {key}")
    return v


def _parse_date_utc(date_str: str) -> int:
    # YYYY-MM-DD -> ms timestamp (UTC 00:00:00)
    dt = datetime.strptime(date_str, "%Y-%m-%d").replace(tzinfo=timezone.utc)
    return int(dt.timestamp() * 1000)


def fetch_okx_candles(symbol: str, bar: str, start_ms: int, end_ms: int, limit: int = 100) -> List[Candle]:
    """
    OKX candles: /api/v5/market/candles
    Response item: [ts, o, h, l, c, vol, volCcy, volCcyQuote, confirm]
    Notes:
    - returned newest first
    - we page backward using before=<oldest_ts_in_batch>
    """
    out: Dict[int, Candle] = {}
    before: Optional[int] = None

    with httpx.Client(timeout=30, headers={"User-Agent": "alpha-trade/1.0"}) as client:
        while True:
            params = {"instId": symbol, "bar": bar, "limit": str(limit)}
            if before is not None:
                params["before"] = str(before)

            r = client.get("https://www.okx.com/api/v5/market/candles", params=params)
            r.raise_for_status()
            j = r.json()
            if j.get("code") != "0":
                raise RuntimeError(f"okx api error: {j}")

            batch = j.get("data") or []
            if not batch:
                break

            oldest_ts = int(batch[-1][0])
            for row in batch:
                ts = int(row[0])
                if ts < start_ms or ts > end_ms:
                    continue
                o, h, l, c = map(float, row[1:5])
                vol = None
                try:
                    vol = float(row[5])
                except Exception:
                    vol = None
                out[ts] = Candle(
                    open_time_ms=ts,
                    open=o,
                    high=h,
                    low=l,
                    close=c,
                    volume=vol,
                    raw={"okx": row},
                )

            if oldest_ts < start_ms:
                break
            before = oldest_ts

    return [out[k] for k in sorted(out.keys())]


def connect_with_retry(dsn: str, retries: int = 60, sleep_s: float = 1.0) -> psycopg.Connection:
    last_err: Optional[Exception] = None
    for _ in range(retries):
        try:
            return psycopg.connect(dsn, autocommit=True)
        except Exception as e:
            last_err = e
            time.sleep(sleep_s)
    raise RuntimeError(f"failed to connect to postgres after retries: {last_err}")


def upsert_candles(
    conn: psycopg.Connection,
    exchange: str,
    symbol: str,
    interval: str,
    candles: List[Candle],
) -> int:
    sql = """
    INSERT INTO market_candles (
      exchange, symbol, interval, open_time,
      open, high, low, close, volume, raw
    )
    VALUES (
      %(exchange)s, %(symbol)s, %(interval)s, %(open_time)s,
      %(open)s, %(high)s, %(low)s, %(close)s, %(volume)s, %(raw)s
    )
    ON CONFLICT (exchange, symbol, interval, open_time)
    DO UPDATE SET
      open = EXCLUDED.open,
      high = EXCLUDED.high,
      low = EXCLUDED.low,
      close = EXCLUDED.close,
      volume = EXCLUDED.volume,
      raw = EXCLUDED.raw
    """
    rows = []
    for c in candles:
        rows.append(
            {
                "exchange": exchange,
                "symbol": symbol,
                "interval": interval,
                "open_time": c.open_time_utc,
                "open": c.open,
                "high": c.high,
                "low": c.low,
                "close": c.close,
                "volume": c.volume,
                "raw": json.dumps(c.raw),
            }
        )

    with conn.cursor() as cur:
        cur.executemany(sql, rows)
    return len(rows)


def main() -> None:
    dsn = _must_env("PG_DSN")
    exchange = os.getenv("KLINE_EXCHANGE", "okx").strip().lower()
    symbol = os.getenv("KLINE_SYMBOL", "BTC-USDT").strip()
    interval = os.getenv("KLINE_INTERVAL", "1D").strip()
    start_date = os.getenv("KLINE_START_DATE", "2025-01-01").strip()
    end_date = os.getenv("KLINE_END_DATE", "2025-12-31").strip()

    if exchange != "okx":
        raise RuntimeError(f"only okx supported in this importer for now, got {exchange}")

    # Normalize interval -> OKX bar
    # OKX bar examples: 15m, 1H, 4H, 1D, 1W ...
    interval_norm = interval.strip()
    if interval_norm.lower() == "15m":
        bar = "15m"
        interval_store = "15m"
    elif interval_norm.upper() == "1H":
        bar = "1H"
        interval_store = "1H"
    elif interval_norm.upper() == "4H":
        bar = "4H"
        interval_store = "4H"
    elif interval_norm.upper() == "1D":
        bar = "1D"
        interval_store = "1D"
    else:
        raise RuntimeError(f"unsupported KLINE_INTERVAL={interval}. supported: 15m, 1H, 4H, 1D")

    start_ms = _parse_date_utc(start_date)
    end_ms = _parse_date_utc(end_date)

    print(f"[kline-loader] source=okx symbol={symbol} interval={interval_store} range={start_date}..{end_date}")
    candles = fetch_okx_candles(symbol=symbol, bar=bar, start_ms=start_ms, end_ms=end_ms, limit=100)
    print(f"[kline-loader] fetched candles={len(candles)}")

    conn = connect_with_retry(dsn)
    try:
        n = upsert_candles(conn, exchange=exchange, symbol=symbol, interval=interval_store, candles=candles)
        print(f"[kline-loader] upserted rows={n}")
    finally:
        conn.close()


if __name__ == "__main__":
    main()


