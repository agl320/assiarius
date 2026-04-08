# assiarius

Small CLI for polling Finviz screeners and running a sentiment verdict on the latest news item per ticker. Works by aggregating tickers using finviz screener url, sorting them via priority queue based on various (to be configurable) signals, and then feeding them through a rate limited request pipeline into Gemini 3 Flash.

## Setup

Create a `.env` file (or export env vars) with at least:

- `GEMINI_API_KEY` — required

Optional:

- `GEMINI_TIMEOUT` — request timeout (examples: `10s`, `1m`, `0s`)
- `GEMINI_MIN_INTERVAL` — minimum delay between Gemini calls when using the poll/queue flow (examples: `2s`, `500ms`)

## Usage

- Poll a screener every 15 seconds (default):
  - `assi poll [screener]`

- Poll a screener with a custom interval:
  - `assi poll [screener] 30`
  - `assi poll [screener] 30s`
  - `assi poll [screener] 1m`

- Run a screener once:
  - `assi screen [preset]`

- Run a screener once and fetch news verdicts (no queue):
  - `assi screen [preset] --news`
