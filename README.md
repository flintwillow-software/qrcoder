# qrcoder

A CLI tool for generating QR codes, built on [yeqown/go-qrcode](https://github.com/yeqown/go-qrcode).

Supports **terminal output** (Unicode block characters) and **image file output** (PNG/JPEG) with customizable colors, shapes, and borders. Colors accept both RGB (`#RRGGBB`) and RGBA (`#RRGGBBAA`) hex formats.

## Quick Start

```bash
# Print a QR code to the terminal
qrcoder -data "Hello, World!"

# Save as PNG
qrcoder -data "https://example.com" -output qr.png

# Transparent background (alpha = 00)
qrcoder -data "https://example.com" -output qr.png -bg "#00000000"

# Partially transparent background (alpha = 80)
qrcoder -data "https://example.com" -output qr.png -bg "#ffffff80"
```

## Terminal Output

When no `-output` flag is provided, the QR code is rendered directly to stdout using Unicode half-block characters:

## Image File Output

Save QR codes as PNG or JPEG. Output format is inferred from the file extension and color values — PNG is used automatically when any color has transparency.

```bash
# PNG with custom colors
qrcoder -data "https://example.com" -output qr.png -fg "#ff0000" -bg "#000000"

# JPEG with circle shape and larger blocks
qrcoder -data "https://example.com" -output qr.jpg -shape circle -width 12

# Transparent background
qrcoder -data "https://example.com" -output qr.png -bg "#00000000"

# Semi-transparent background
qrcoder -data "https://example.com" -output qr.png -bg "#00000080"
```

## Color Format

Both `-fg` and `-bg` accept hex colors in RGB or RGBA format:

| Format | Example | Description |
|--------|---------|-------------|
| `#RRGGBB` | `#ff0000` | Solid color |
| `#RRGGBBAA` | `#ff000080` | Color with alpha (00–ff) |

When any color has an alpha value less than `ff`, the output is automatically saved as PNG to preserve transparency.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-data` | *(required)* | Content to encode in the QR code |
| `-output` | *(none)* | Output file path; omit for terminal output |
| `-fg` | `#000000` | Foreground color (hex, RGB or RGBA) |
| `-bg` | `#ffffff` | Background color (hex, RGB or RGBA) |
| `-shape` | `rect` | Cell shape: `rect` or `circle` |
| `-width` | `8` | Block width in pixels |
| `-border-top` | `4` | Top border width in blocks |
| `-border-right` | `4` | Right border width in blocks |
| `-border-bottom` | `4` | Bottom border width in blocks |
| `-border-left` | `4` | Left border width in blocks |

## Building

```bash
just build
```

## Development

```bash
just test     # run tests
just fmt      # format code
just lint     # run staticcheck
just clean    # remove build artifacts
```

## Dependencies

- [yeqown/go-qrcode/v2](https://github.com/yeqown/go-qrcode) — QR code generation core
- [yeqown/go-qrcode/writer/standard](https://github.com/yeqown/go-qrcode/tree/main/writer/standard) — image file writer
