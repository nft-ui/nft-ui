# nft-ui

A web-based UI for managing nftables outbound traffic quotas.

## Features

- View quota rules with circular progress indicators
- Add / Edit / Delete quota rules
- Reset quota usage (single or batch)
- Manage allowed inbound ports
- Visual indicator for inbound port status
- Basic authentication support
- Read-only mode
- Auto-refresh

## Installation

### Quick Install

**Stable (main branch):**
```bash
curl -fsSL https://raw.githubusercontent.com/d3vw/nft-ui/main/install.sh | sudo bash
```

**Beta (dev branch):**
```bash
curl -fsSL https://raw.githubusercontent.com/d3vw/nft-ui/dev/install.sh | sudo bash -s -- --beta
```

### Manual Install

Download the binary from [Releases](https://github.com/d3vw/nft-ui/releases):

```bash
# Linux amd64
curl -fsSL -o nft-ui https://github.com/d3vw/nft-ui/releases/latest/download/nft-ui-linux-amd64

# Linux arm64
curl -fsSL -o nft-ui https://github.com/d3vw/nft-ui/releases/latest/download/nft-ui-linux-arm64

chmod +x nft-ui
sudo mv nft-ui /usr/local/bin/
```

## Usage

```bash
# Run with default settings (port 8080)
sudo nft-ui

# With environment variables
sudo NFT_UI_LISTEN_ADDR=:3000 NFT_UI_AUTH_USER=admin NFT_UI_AUTH_PASSWORD=secret nft-ui
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NFT_UI_LISTEN_ADDR` | `:8080` | Server listen address |
| `NFT_UI_AUTH_USER` | - | Basic auth username |
| `NFT_UI_AUTH_PASSWORD` | - | Basic auth password |
| `NFT_UI_READ_ONLY` | `false` | Disable write operations |
| `NFT_UI_REFRESH_INTERVAL` | `5` | Auto-refresh interval (seconds) |
| `NFT_UI_NFT_BINARY` | `/usr/sbin/nft` | Path to nft binary |
| `NFT_UI_TABLE_FAMILY` | `inet` | nftables family |
| `NFT_UI_TABLE_NAME` | `filter` | nftables table name |
| `NFT_UI_CHAIN_NAME` | `output` | nftables chain name |

## Systemd Service

```bash
sudo tee /etc/systemd/system/nft-ui.service > /dev/null <<EOF
[Unit]
Description=nft-ui - nftables quota management web interface
After=network.target nftables.service

[Service]
Type=simple
ExecStart=/usr/local/bin/nft-ui
Environment=NFT_UI_LISTEN_ADDR=:8080
Restart=on-failure
RestartSec=5
AmbientCapabilities=CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now nft-ui
```

## Expected nftables Rule Format

Quota rules in the `output` chain:

```
meta l4proto { tcp, udp } th sport 18444 quota over 100000 mbytes drop comment "block 18444 after 100GB"
```

Allowed port rules in the `input` chain:

```
tcp dport 8080 accept comment "nft-ui managed"
```

## Build from Source

```bash
# Requirements: Go 1.21+, Node.js 18+
make
```

## License

MIT
