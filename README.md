# assiarius

Small CLI for polling Finviz screeners and hooking them up to Discord via webhook. Optionally send ticker statistics to an LLM (Gemini Flash 2.5 Lite by default) for sentiment identification.

## Setup

Create a `.env` file (or export env vars) with at least:

- `GEMINI_API_KEY` — required

Optional:

- `GEMINI_TIMEOUT` — request timeout (examples: `10s`, `1m`, `0s`)
- `GEMINI_MIN_INTERVAL` — minimum delay between Gemini calls when using the poll/queue flow (examples: `2s`, `500ms`)
- `GEMINI_MODEL` — override Gemini model name (default: `models/gemini-2.5-flash-lite`)

## Usage

- Poll a _news_ screener URL every 5 minutes (default):
  - `assi poll [screenerURL]`

- Poll a news screener URL with a custom interval (window matches interval):
  - `assi poll [screenerURL] 5m`
  - `assi poll [screenerURL] 1h`

- Run a screener once:
  - `assi screen [preset]`

- Run a screener once and fetch news verdicts (no queue):
  - `assi screen [preset] --news`
