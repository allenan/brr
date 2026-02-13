# brr

**internet go brr**

<!-- TODO: Record demo GIF using VHS or asciinema (15-20s, dark terminal, clean font) -->
<!-- Should show: test starting → progress bars filling with sparklines → speed numbers spring-animating → latency panel → bufferbloat grade -->
<p align="center">
  <img src="assets/demo.gif" alt="brr demo" width="720">
</p>

<p align="center">
  <a href="https://github.com/allenan/brr/releases"><img src="https://img.shields.io/github/v/release/allenan/brr?style=flat-square" alt="Release"></a>
  <a href="https://github.com/allenan/brr/actions"><img src="https://img.shields.io/github/actions/workflow/status/allenan/brr/ci.yml?style=flat-square" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/allenan/brr"><img src="https://goreportcard.com/badge/github.com/allenan/brr?style=flat-square" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License"></a>
</p>

brr is a terminal-based internet speed test that goes beyond download and upload numbers. It measures **bufferbloat**, the hidden network problem that causes video calls to freeze, games to lag, and pages to hang even when your "speed" looks fine, and gives you a letter grade from A+ to F.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). No API keys. No accounts. Just run `brr`.

## Quick start

```sh
brew install allenan/tap/brr && brr
```

## Features

- **Real-time sparklines & spring-animated numbers**: watch your speeds fill in live
- **Bufferbloat grading**: A+ through F, so you know if your connection actually feels fast
- **Latency, jitter & loaded latency**: idle ping is a lie; brr measures latency under load
- **History with trend tracking**: see how your connection changes over time
- **Multiple output modes**: TUI (default), `--fullscreen`, `--json`, `--simple`
- **Accessible themes**: vivid default, colorblind-safe Okabe-Ito palette, monochrome, plus `NO_COLOR` support
- **Zero configuration**: uses Cloudflare's speed test infrastructure, no API keys needed
- **Cross-platform**: Linux, macOS, Windows on amd64 and arm64

## Installation

### Homebrew

```sh
brew install allenan/tap/brr
```

### Go

```sh
go install github.com/allenan/brr/cmd/brr@latest
```

### Binary releases

Download from [GitHub Releases](https://github.com/allenan/brr/releases). Examples:

```sh
# macOS (Apple Silicon)
curl -Lo brr.tar.gz https://github.com/allenan/brr/releases/latest/download/brr_darwin_arm64.tar.gz
tar xzf brr.tar.gz && sudo mv brr /usr/local/bin/

# Linux (amd64)
curl -Lo brr.tar.gz https://github.com/allenan/brr/releases/latest/download/brr_linux_amd64.tar.gz
tar xzf brr.tar.gz && sudo mv brr /usr/local/bin/
```

### Build from source

```sh
git clone https://github.com/allenan/brr.git && cd brr
go build -o brr ./cmd/brr
```

## Usage

### Run a test

```sh
brr
```

That's it. Results are saved to history automatically.

### Output modes

| Flag | Description |
|------|-------------|
| *(none)* | Interactive TUI with sparklines and animations |
| `--fullscreen` | TUI in alt-screen mode |
| `--json` | Machine-readable JSON output |
| `--simple` | Single summary line |

The `--simple` output is the thing you screenshot and paste into Slack:

```
$ brr --simple
↓ 308.2 Mbps  ↑ 31.4 Mbps  ⏱ 24.0ms  Bloat: A+  US → Ashburn, VA
```

### History & comparison

```sh
brr --history          # Show past runs
brr --compare          # Compare with previous result
```

### Themes

```sh
brr --theme colorblind  # Okabe-Ito safe palette
brr --theme mono        # No colors
NO_COLOR=1 brr          # Also triggers monochrome
```

### Keybindings

When the test completes:

| Key | Action |
|-----|--------|
| `h` | History |
| `e` | Export JSON |
| `c` | Compare |
| `q` | Quit |

During a test, press `q` to cancel.

<details>
<summary><strong>JSON output example</strong></summary>

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "server": {
    "ip": "203.0.113.1",
    "colo": "IAD",
    "colo_city": "Ashburn, VA",
    "location": "US",
    "isp": "Example ISP"
  },
  "download": {
    "mbps": 308.2,
    "samples": [...]
  },
  "upload": {
    "mbps": 31.4,
    "samples": [...]
  },
  "idle_latency": {
    "min_ms": 12.3,
    "max_ms": 28.1,
    "avg_ms": 18.5,
    "jitter_ms": 2.1,
    "samples": [...]
  },
  "download_latency": {
    "min_ms": 14.2,
    "max_ms": 42.7,
    "avg_ms": 22.3,
    "jitter_ms": 4.8,
    "samples": [...]
  },
  "upload_latency": {
    "min_ms": 13.8,
    "max_ms": 38.9,
    "avg_ms": 21.1,
    "jitter_ms": 3.9,
    "samples": [...]
  },
  "bufferbloat_download": "A+",
  "bufferbloat_upload": "A",
  "context_line": "Excellent for 4K streaming, video calls, and gaming"
}
```

</details>

## What is bufferbloat?

Your ISP advertises 500 Mbps. Speed tests confirm it. But your video calls still freeze.

The problem is **bufferbloat**: oversized network buffers that absorb packets during heavy traffic. Your throughput looks fine, but latency spikes to hundreds of milliseconds. Everything that needs real-time responsiveness (video calls, gaming, even scrolling a web page) suffers.

brr measures this by pinging during the download and upload phases, not just when the connection is idle. The difference between idle latency and loaded latency determines your grade:

| Grade | Latency increase | What it means |
|-------|-----------------|---------------|
| A+ | < 5 ms | Excellent. No perceptible bloat. |
| A | < 30 ms | Great. Minimal impact on real-time apps. |
| B | < 60 ms | Good. Occasional hiccups under heavy load. |
| C | < 200 ms | Fair. Video calls and gaming noticeably affected. |
| D | < 400 ms | Poor. Frequent freezes and lag spikes. |
| F | 400 ms+ | Bad. Connection becomes unusable under load. |

> A 500 Mbps connection with an F grade will feel worse for daily use than a 50 Mbps connection with an A+. Most speed tests don't measure this.

## How it works

1. **Metadata**: connect to Cloudflare's speed test infrastructure and identify the nearest edge server
2. **Idle latency**: 20 pings to establish a baseline RTT
3. **Download**: progressive transfer sizes (100 KB to 25 MB) across 16 parallel connections, sampling throughput every 100ms
4. **Upload**: same progressive strategy with upload-appropriate sizes
5. **Loaded latency**: latency probes sent every 400ms *during* download and upload phases
6. **Grading**: compare median idle latency to median loaded latency; the delta determines your bufferbloat grade

brr uses Cloudflare's speed test infrastructure, the same backend as their browser-based test.

## What brr adds

Things brr does that most speed tests don't:

- [x] Bufferbloat grading (A+ through F)
- [x] Latency measured under load, not just idle
- [x] Real-time sparkline visualizations
- [x] Built-in history with trend tracking
- [x] Export to JSON
- [x] Colorblind-safe and monochrome themes

For official ISP certification or testing against specific servers, use [Ookla's CLI](https://www.speedtest.net/apps/cli).

## Configuration

brr has no configuration file. Simplicity is a feature.

History is stored at the OS-default config path:

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/brr/history.json` |
| Linux | `~/.config/brr/history.json` |
| Windows | `%AppData%\brr\history.json` |

## Accessibility

brr ships with three themes:

- **`default`**: vivid colors for dark terminals
- **`colorblind`**: [Okabe-Ito](https://jfly.uni-koeln.de/color/) palette, distinguishable with all forms of color vision deficiency
- **`mono`**: no colors at all; auto-enabled when [`NO_COLOR`](https://no-color.org) is set

The `--simple` and `--json` output modes work well with screen readers and text processing tools.

## Contributing

Contributions welcome. Please open an issue to discuss before submitting a PR.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) for TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for terminal styling
- [Harmonica](https://github.com/charmbracelet/harmonica) for spring-physics animations
- [ntcharts](https://github.com/NimbleMarkets/ntcharts) for sparkline charts
- [Cobra](https://github.com/spf13/cobra) for CLI framework

## License

[MIT](LICENSE)
