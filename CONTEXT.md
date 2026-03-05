# CONTEXT.md — AI Agent Guide for appc

`appc` is a CLI tool for fetching App Store Connect data.
Use this guide when integrating `appc` into AI agent workflows.

## Key Principles

- Always use `--format json` for machine-readable output
- Data goes to **stdout**, status/warnings go to **stderr**
- Exit code 0 = success, 1 = error
- Errors are printed to stderr as `Error: <message>` (no usage text)

## Authentication

Set these environment variables (preferred for agents):

```bash
export APPC_ISSUER_ID="your-issuer-id"
export APPC_KEY_ID="your-key-id"
export APPC_PRIVATE_KEY_PATH="/path/to/AuthKey.p8"
export APPC_VENDOR_NUMBER="your-vendor-number"
```

Environment variables take priority over the config file (`~/.config/appc/config.json`).

## Commands

### List Apps

```bash
appc apps --format json
appc apps --country us --format json
```

Returns: array of app objects with `id`, `name`, `bundle_id`, `sku`, `rating`, `rating_count`.

### Fetch Reviews

```bash
appc reviews --app 123456789 --format json
appc reviews --app 123456789 --rating 1 --limit 50 --format json
appc reviews --app 123456789 --summary --format json
```

- `--app` (required): numeric App ID
- `--rating`: filter by star rating (1-5)
- `--limit`: max reviews (default 100)
- `--summary`: return rating summary instead of individual reviews

### Fetch Sales Reports

```bash
appc sales --date 2024-01-15 --format json
appc sales --from 2024-01-01 --to 2024-01-31 --format json
appc sales --date 2024-01 --frequency MONTHLY --format json
```

- `--date`: single date (YYYY-MM-DD or YYYY-MM)
- `--from` / `--to`: date range (daily only, must be used together)
- `--frequency`: DAILY (default), WEEKLY, MONTHLY, YEARLY
- `--vendor`: override vendor number from config

### Fetch Analytics

```bash
appc analytics --app 123456789 --format json
appc analytics --app 123456789 --category APP_USAGE --format json
appc analytics --app 123456789 --snapshot --format json
```

- `--app` (required): numeric App ID
- `--category`: APP_USAGE, APP_STORE_ENGAGEMENT, COMMERCE, FRAMEWORK_USAGE, PERFORMANCE
- `--snapshot`: retrieve historical data from app creation date

### Lookup App Ratings

```bash
appc lookup --app 123456789 --format json
appc lookup --app 123456789,987654321 --country us --format json
```

- `--app` (required): one or more numeric App IDs (comma-separated)
- `--country`: store country code (default: jp)
- Does not require App Store Connect credentials

## Error Handling

- Check the exit code: 0 = success, non-zero = failure
- Parse stderr for error messages (format: `Error: <description>`)
- Warnings (non-fatal) also go to stderr (format: `warning: <description>`)
- Invalid input (non-numeric App IDs, bad dates) is rejected before any API call

## Typical Agent Workflow

```bash
# 1. Get app list to find App IDs
APPS=$(appc apps --format json)

# 2. Extract an App ID (using jq)
APP_ID=$(echo "$APPS" | jq -r '.[0].id')

# 3. Fetch reviews for analysis
appc reviews --app "$APP_ID" --format json

# 4. Fetch sales data for a date range
appc sales --from 2024-01-01 --to 2024-01-31 --format json
```
