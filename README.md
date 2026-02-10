# asa-cli

A command-line tool for the [Apple Search Ads](https://searchads.apple.com) Campaign Management API (v5). Manage campaigns, ad groups, keywords, negative keywords, and pull performance reports — all from your terminal.

Built to be scriptable and AI-agent friendly. Use `-o json` for machine-readable output, or default to `-o table` for a human-readable view.

## Getting Started

### Install

Requires [Go 1.24+](https://golang.org/dl/).

```bash
git clone https://github.com/TrebuhS/asa-cli.git
cd asa-cli
make install
```

This installs the `asa-cli` binary to `~/go/bin`. Make sure it's on your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"  # add to ~/.zshrc or ~/.bashrc
```

### Set Up API Access

Apple Search Ads uses OAuth2 with ES256-signed JWTs. You'll generate a key pair locally and upload the public half to Apple.

**1. Generate a key pair:**

```bash
openssl ecparam -genkey -name prime256v1 -noout -out private-key.pem
openssl ec -in private-key.pem -pubout -out public-key.pem
```

**2. Upload the public key to Apple:**

Go to [searchads.apple.com](https://searchads.apple.com) > Account Settings > API, then paste the contents of `public-key.pem` into the Public Key field. Note the **Client ID**, **Team ID**, and **Key ID** shown on that page.

**3. Store the private key:**

```bash
mkdir -p ~/.asa-cli && chmod 700 ~/.asa-cli
mv private-key.pem ~/.asa-cli/private-key.pem
chmod 600 ~/.asa-cli/private-key.pem
```

### Configure

```bash
asa-cli configure \
  --client-id "SEARCHADS.xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --team-id "SEARCHADS.xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --key-id "your-key-id" \
  --private-key-path "~/.asa-cli/private-key.pem"
```

Or run `asa-cli configure` with no flags for interactive mode.

**Org ID is optional.** If your account has one organization, it's auto-detected. For multi-org accounts, pass `--org-id` per-command or set it in config.

### Verify

```bash
asa-cli whoami
```

## Usage

### Campaigns

```bash
asa-cli campaigns list
asa-cli campaigns get 123456789
asa-cli campaigns find --filter "status=ENABLED" --sort "name:asc"
asa-cli campaigns create \
  --name "Brand - US" \
  --budget 10000 --daily-budget 100 \
  --countries US --app-id 123456789
asa-cli campaigns update 123456789 --status PAUSED --daily-budget 50
asa-cli campaigns delete 123456789
```

### Ad Groups

Scoped under a campaign with `--campaign-id`.

```bash
asa-cli adgroups list --campaign-id 123
asa-cli adgroups create --campaign-id 123 \
  --name "Exact Match" --default-bid 1.50 --cpa-goal 5.00
asa-cli adgroups update 456 --campaign-id 123 --default-bid 2.00
```

### Keywords

Scoped under a campaign and ad group.

```bash
asa-cli keywords list --campaign-id 123 --adgroup-id 456

# Bulk create
asa-cli keywords create --campaign-id 123 --adgroup-id 456 \
  --text "habit tracker" --text "daily habits" --text "habit app" \
  --match-type EXACT --bid 1.50

# Update bid
asa-cli keywords update --campaign-id 123 --adgroup-id 456 --id 789 --bid 2.00

# Delete (comma-separated)
asa-cli keywords delete 789,790,791 --campaign-id 123 --adgroup-id 456
```

### Negative Keywords

Campaign-level and ad-group-level.

```bash
# Campaign-level
asa-cli negative-keywords campaign-create --campaign-id 123 \
  --text "free" --text "cheap" --match-type EXACT
asa-cli negative-keywords campaign-list --campaign-id 123
asa-cli negative-keywords campaign-delete 789,790 --campaign-id 123

# Ad group-level
asa-cli negative-keywords adgroup-create --campaign-id 123 --adgroup-id 456 \
  --text "competitor" --match-type BROAD
```

### Reports

All reports require `--start-date` and `--end-date` (YYYY-MM-DD).

```bash
asa-cli reports campaigns --start-date 2024-01-01 --end-date 2024-01-31 --granularity DAILY
asa-cli reports adgroups  --campaign-id 123 --start-date 2024-01-01 --end-date 2024-01-31
asa-cli reports keywords  --campaign-id 123 --start-date 2024-01-01 --end-date 2024-01-31
asa-cli reports search-terms --campaign-id 123 --start-date 2024-01-01 --end-date 2024-01-31

# Group by country and device
asa-cli reports campaigns \
  --start-date 2024-01-01 --end-date 2024-01-31 \
  --granularity WEEKLY --group-by countryOrRegion,deviceClass -o json
```

Metrics: impressions, taps, installs, new downloads, redownloads, TTR, conversion rate, avg CPA, avg CPT, spend.

### Apps & Geo Search

```bash
asa-cli apps search --query "MyApp"
asa-cli apps search --query "MyApp" --owned
asa-cli geo search --query "United States"
asa-cli geo search --query "California" --country-code US
```

## Filters & Sorting

Use `--filter` with shorthand operators:

| Operator | Meaning | Example |
|----------|---------|---------|
| `=` | Equals | `status=ENABLED` |
| `~` | Contains | `name~Brand` |
| `!~` | Not contains | `name!~Test` |
| `@` | In (comma-separated) | `status@ENABLED,PAUSED` |
| `>` `<` `>=` `<=` | Comparison | `id>1000` |

Use `--sort` as `field:direction`:

```bash
asa-cli campaigns find --filter "status=ENABLED" --sort "name:asc" --limit 50
```

Use `--all` to auto-paginate and fetch every result.

## Scripting

Use `-o json` and pipe to `jq`:

```bash
# List all campaign IDs
asa-cli campaigns list -o json | jq '.[].id'

# Pause all campaigns
for id in $(asa-cli campaigns list -o json | jq -r '.[].id'); do
  asa-cli campaigns update "$id" --status PAUSED
done
```

## Configuration

Stored at `~/.asa-cli/config.yaml`. Tokens are cached at `~/.asa-cli/token_cache.json`.

### Multiple Profiles

```bash
asa-cli configure -p production --client-id "..." --team-id "..." --key-id "..." --private-key-path "..."
asa-cli campaigns list -p production
```

### Environment Variables

Override any config value:

| Variable | Description |
|----------|-------------|
| `ASA_CLIENT_ID` | Client ID |
| `ASA_TEAM_ID` | Team ID |
| `ASA_KEY_ID` | Key ID |
| `ASA_ORG_ID` | Organization ID |
| `ASA_PRIVATE_KEY_PATH` | Path to private key |

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | `json` or `table` (default: `table`) |
| `--profile` | `-p` | Named config profile |
| `--org-id` | | Organization ID (overrides config) |
| `--verbose` | `-v` | Show HTTP request/response details |
| `--no-color` | | Disable colored output |

## Contributing

```bash
git clone https://github.com/TrebuhS/asa-cli.git
cd asa-cli
go build -o asa-cli .
./asa-cli --help
```

Issues and PRs welcome.

## License

[MIT](LICENSE)
