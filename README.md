# appc - App Store Connect CLI

A CLI tool for fetching App Store Connect data (apps, sales reports, reviews, analytics).

## Setup

```bash
go install github.com/kenta-tanaka/appc@latest
```

Or build from source:

```bash
go build -o appc .
```

## Configuration

Set up your App Store Connect API credentials:

```bash
appc configure
```

You'll need:
- **Issuer ID** - from App Store Connect > Users and Access > Keys
- **Key ID** - from the API key you created
- **Private Key Path** - path to the `.p8` file downloaded when creating the key
- **Vendor Number** - your vendor number for sales reports

Config is saved to `~/.config/appc/config.json` with `0600` permissions.

## Usage

### List Apps

```bash
appc apps
appc apps --format csv
```

### Sales Reports

```bash
appc sales --date 2025-01-15
appc sales --date 2025-01-15 --type SALES --frequency DAILY --format csv
appc sales --date 2025-01 --frequency MONTHLY --vendor 12345678
```

### Customer Reviews

```bash
appc reviews --app <APP_ID>
appc reviews --app <APP_ID> --rating 5 --limit 50
appc reviews --app <APP_ID> --format csv > reviews.csv
```

### Analytics

```bash
appc analytics --app <APP_ID>
appc analytics --app <APP_ID> --category APP_USAGE
```

### Global Flags

- `--format json|csv` - Output format (default: json)

## Output

- Data goes to **stdout** (pipe-friendly)
- Status messages and errors go to **stderr**

```bash
# Pipe to jq for analysis
appc apps | jq '.[].name'

# Pipe to Claude Code
appc reviews --app <ID> --format json | claude "Analyze these reviews"

# Export to CSV
appc sales --date 2025-01-15 --format csv > sales.csv
```

## Development

```bash
# Run tests
go test ./...

# Run tests with verbose output
go test ./... -v
```

## Authentication

Uses JWT (ES256) tokens for the App Store Connect API. Tokens are:
- Valid for 20 minutes
- Automatically cached and refreshed
- Generated using your `.p8` private key
