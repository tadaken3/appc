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

## Commands

### `appc apps`

List all apps in your App Store Connect account.

```bash
appc apps
appc apps --format csv
```

### `appc sales`

Download sales and trends reports.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--date` | string | *(required)* | Report date (`YYYY-MM-DD` or `YYYY-MM`) |
| `--type` | string | `SALES` | Report type: `SALES`, `PRE_ORDER`, `NEWSSTAND` |
| `--frequency` | string | `DAILY` | Frequency: `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY` |
| `--sub-type` | string | `SUMMARY` | Sub type: `SUMMARY`, `DETAILED`, `OPT_IN` |
| `--vendor` | string | *(from config)* | Vendor number (overrides config value) |

```bash
appc sales --date 2025-01-15
appc sales --date 2025-01-15 --type SALES --frequency DAILY --format csv
appc sales --date 2025-01 --frequency MONTHLY --vendor 12345678
```

### `appc reviews`

Retrieve customer reviews for an app.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--app` | string | *(required)* | App ID |
| `--rating` | int | `0` | Filter by star rating (1–5; 0 = all) |
| `--limit` | int | `100` | Maximum number of reviews to return |

```bash
appc reviews --app <APP_ID>
appc reviews --app <APP_ID> --rating 5 --limit 50
appc reviews --app <APP_ID> --format csv > reviews.csv
```

### `appc analytics`

Request and retrieve App Store analytics reports.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--app` | string | *(required)* | App ID |
| `--category` | string | | Report category (e.g., `APP_USAGE`, `APP_STORE_ENGAGEMENT`) |

```bash
appc analytics --app <APP_ID>
appc analytics --app <APP_ID> --category APP_USAGE
```

### `appc configure`

Interactively set up or update authentication credentials. Existing values are shown as defaults — press Enter to keep them.

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
│   └── configure.go        #   appc configure
└── internal/
    ├── api/                # App Store Connect API clients
    │   ├── types.go        #   Shared response/paging types
    │   ├── apps.go         #   Apps endpoint
    │   ├── sales.go        #   Sales reports endpoint
    │   ├── reviews.go      #   Reviews endpoint
    │   └── analytics.go    #   Analytics endpoint
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
