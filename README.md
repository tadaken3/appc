# appc - App Store Connect CLI

[English](README.md) | [日本語](README.ja.md)

A CLI tool for indie developers to fetch App Store Connect data — sales reports, reviews, and analytics — and pipe it into AI tools like Claude Code for growth insights.

## Installation

### Homebrew (recommended)

```bash
brew tap tadaken3/appc
brew install appc
```

### go install

```bash
go install github.com/tadaken3/appc@latest
```

### Build from source

```bash
git clone https://github.com/tadaken3/appc.git
cd appc
go build -o appc .
```

## Setup

### 1. Create an API Key

1. Open [App Store Connect](https://appstoreconnect.apple.com/) and navigate to **Users and Access > Integrations > App Store Connect API**.
2. Click **Generate API Key** and select a role (e.g., Admin, Finance).
3. Download the `.p8` private key file. **This can only be downloaded once.**
4. Note down the **Key ID** and the **Issuer ID** shown on the page.

### 2. Configure appc

Run the interactive setup:

```bash
appc configure
```

You will be prompted for:

| Prompt | Description |
|---|---|
| **Issuer ID** | UUID shown in App Store Connect > Users and Access > Keys |
| **Key ID** | Alphanumeric ID of the API key you created |
| **Private Key Path** | Path to the `.p8` file (supports `~` expansion) |
| **Vendor Number** | Your vendor number for sales reports |

Configuration is saved to `~/.config/appc/config.json` with `0600` permissions.

### Environment variables (CI / cloud environments)

In environments where running `appc configure` or placing a `.p8` file on disk
is inconvenient (CI, containers, Claude Code on the web, etc.), the
configuration can be supplied entirely through environment variables. When set,
they **override** the values in the config file, so a config file is optional.

| Variable | Description |
|---|---|
| `APPC_ISSUER_ID` | Issuer ID |
| `APPC_KEY_ID` | Key ID |
| `APPC_PRIVATE_KEY` | PEM content of the `.p8` private key (inline) |
| `APPC_PRIVATE_KEY_PATH` | Path to the `.p8` file (alternative to `APPC_PRIVATE_KEY`) |
| `APPC_VENDOR_NUMBER` | Vendor number for sales reports |

`APPC_PRIVATE_KEY` takes precedence over `APPC_PRIVATE_KEY_PATH`. Example:

```bash
export APPC_ISSUER_ID="..."
export APPC_KEY_ID="..."
export APPC_VENDOR_NUMBER="..."
export APPC_PRIVATE_KEY="$(cat AuthKey_XXXX.p8)"
appc configure --validate   # verify credentials resolve correctly
```

## Commands

### `appc apps`

List all apps in your App Store Connect account. Automatically enriches each app with its App Store rating via the iTunes Lookup API.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--country` | string | `jp` | Country code for rating lookup |

```bash
appc apps
appc apps --country us --format csv
```

### `appc lookup`

Look up app ratings and metadata from the iTunes Lookup API. Supports multiple App IDs for competitor research — no App Store Connect credentials required.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--app` | string | *(required)* | App ID(s), comma-separated |
| `--country` | string | `jp` | Country code for store lookup |

```bash
appc lookup --app 6745560143 --country jp
appc lookup --app 6745560143,123456789 --country us    # competitor research
appc lookup --app 6745560143 --format csv
```

### `appc sales`

Download sales and trends reports. Dates with no data (HTTP 404) are treated as empty results instead of errors.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--date` | string | *(required)* | Report date (`YYYY-MM-DD` or `YYYY-MM`) |
| `--from` | string | | Start date for range (`YYYY-MM-DD`, requires `--to`, daily only) |
| `--to` | string | | End date for range (`YYYY-MM-DD`, requires `--from`, daily only) |
| `--type` | string | `SALES` | Report type: `SALES`, `PRE_ORDER`, `NEWSSTAND` |
| `--frequency` | string | `DAILY` | Frequency: `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY` |
| `--sub-type` | string | `SUMMARY` | Sub type: `SUMMARY`, `DETAILED`, `OPT_IN` |
| `--vendor` | string | *(from config)* | Vendor number (overrides config value) |

*\* Either `--date` or `--from`/`--to` is required.*

```bash
appc sales --date 2025-01-15
appc sales --date 2025-01-15 --type SALES --frequency DAILY --format csv
appc sales --date 2025-01 --frequency MONTHLY --vendor 12345678
appc sales --from 2025-01-01 --to 2025-01-31
```

### `appc reviews`

Retrieve customer reviews for an app. With `--summary`, shows a rating breakdown enriched with the overall App Store rating from iTunes.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--app` | string | *(required)* | App ID |
| `--rating` | int | `0` | Filter by star rating (1–5; 0 = all) |
| `--limit` | int | `100` | Maximum number of reviews to return |
| `--summary` | bool | `false` | Show rating summary instead of full reviews |
| `--country` | string | `jp` | Country code for rating lookup (used with `--summary`) |

```bash
appc reviews --app <APP_ID>
appc reviews --app <APP_ID> --rating 5 --limit 50
appc reviews --app <APP_ID> --summary                    # includes App Store overall rating
appc reviews --app <APP_ID> --summary --country us
appc reviews --app <APP_ID> --format csv > reviews.csv
```

### `appc analytics`

Request and retrieve App Store analytics reports. Downloads segment CSV data and outputs parsed records.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--app` | string | *(required)* | App ID |
| `--category` | string | | Report category |
| `--snapshot` | bool | `false` | Use ONE_TIME_SNAPSHOT to retrieve historical data |

Available categories: `APP_USAGE`, `APP_STORE_ENGAGEMENT`, `COMMERCE`, `FRAMEWORK_USAGE`, `PERFORMANCE`

```bash
# Fetch latest analytics (ONGOING mode — data available from request creation date onward)
appc analytics --app <APP_ID> --category APP_USAGE
appc analytics --app <APP_ID> --category APP_STORE_ENGAGEMENT --format csv

# Fetch historical data (from app creation date to request date)
appc analytics --app <APP_ID> --category APP_USAGE --snapshot

# Pipe to Claude for analysis
appc analytics --app <APP_ID> --category APP_USAGE | claude "Analyze the install trends"
```

> **Note:** The first time you run `analytics`, Apple needs time to generate report instances (typically 1–2 days for ONGOING, several hours for snapshots). Subsequent runs will return data immediately.

### `appc configure`

Interactively set up or update authentication credentials. Existing values are shown as defaults — press Enter to keep them.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--validate` | bool | `false` | Validate the current configuration without modifying it |

## Global Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--format` | string | `json` | Output format: `json` or `csv` |

Applies to all commands.

## Output

- **Data** is written to **stdout** (pipe-friendly).
- **Status messages and errors** are written to **stderr**.

```bash
# Filter with jq
appc apps | jq '.[].name'

# Export to CSV
appc sales --date 2025-01-15 --format csv > sales.csv

# Pipe to Claude Code for analysis
appc reviews --app <APP_ID> --format json | claude "Summarize the sentiment of these reviews"
```

## Project Structure

```
appc/
├── main.go                 # Entry point
├── cmd/                    # CLI command definitions (cobra)
│   ├── root.go             #   Root command and global flags
│   ├── apps.go             #   appc apps
│   ├── sales.go            #   appc sales
│   ├── reviews.go          #   appc reviews
│   ├── analytics.go        #   appc analytics
│   ├── lookup.go           #   appc lookup
│   └── configure.go        #   appc configure
└── internal/
    ├── api/                # App Store Connect API clients
    │   ├── types.go        #   Shared response/paging types
    │   ├── apps.go         #   Apps endpoint
    │   ├── sales.go        #   Sales reports endpoint
    │   ├── reviews.go      #   Reviews endpoint
    │   ├── analytics.go    #   Analytics endpoint
    │   └── lookup.go       #   iTunes Lookup API
    ├── auth/               # JWT (ES256) token generation
    │   └── jwt.go
    ├── client/             # HTTP client with retry and auth
    │   └── client.go
    ├── config/             # Configuration load/save
    │   └── config.go
    └── output/             # JSON/CSV output formatting
        └── formatter.go
```

## Development

```bash
# Run all tests
go test ./...

# Verbose output
go test ./... -v
```

## Authentication

appc authenticates with the App Store Connect API using **JWT signed with ES256** (ECDSA P-256 + SHA-256).

- The token is created from your **Issuer ID**, **Key ID**, and `.p8` **private key**.
- Claims include `iss`, `iat`, `exp`, and `aud` (`appstoreconnect-v1`).
- Tokens are **valid for 20 minutes** and automatically cached.
- A cached token is refreshed when it is within 1 minute of expiry.
- Token generation is thread-safe.
