package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"qrcoder/shapes"
)

func main() {
	// Flags
	content := flag.String("data", "", "Content to encode in the QR code (required)")
	output := flag.String("output", "", "Output file path (e.g., qrcode.png). If omitted, prints to terminal.")
	fgColor := flag.String("fg", "#000000", "Foreground color as hex string (e.g., '#000000' or '#000000ff')")
	bgColor := flag.String("bg", "#ffffff", "Background color as hex string (e.g., '#ffffff' or '#ffffff00' for transparent)")
	shape := flag.String("shape", "rect", "Cell shape: 'rect', 'circle', 'diamond', 'star', 'dot', 'rounded', or 'rounded-na' (neighbor-aware)")
	qrWidth := flag.Uint("width", 8, "Width of each QR block in pixels")
	borderTop := flag.Int("border-top", 4, "Top border width in blocks")
	borderRight := flag.Int("border-right", 4, "Right border width in blocks")
	borderBottom := flag.Int("border-bottom", 4, "Bottom border width in blocks")
	borderLeft := flag.Int("border-left", 4, "Left border width in blocks")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `qrcoder - Generate QR codes from the command line

Usage: qrcoder [flags]

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Examples:
  qrcoder -data "Hello, World!"
  qrcoder -data "https://example.com" -output qr.png
  qrcoder -data "Hello" -fg "#ff0000" -bg "#00ff00" -shape circle
  qrcoder -data "Hello" -output qr.png -bg "#00000000"
`)
	}
	flag.Parse()

	// Validate required flag
	if *content == "" {
		fmt.Fprintln(os.Stderr, "error: -data is required")
		flag.Usage()
		os.Exit(1)
	}

	// Determine output mode
	if *output == "" {
		printToTerminal(*content)
	} else {
		saveToFile(*content, *output, *fgColor, *bgColor, *shape, *qrWidth,
			*borderTop, *borderRight, *borderBottom, *borderLeft)
	}
}

// parseColor parses a hex color string supporting both RGB (#RRGGBB) and RGBA (#RRGGBBAA) formats.
func parseColor(hex string) (color.NRGBA, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 && len(hex) != 8 {
		return color.NRGBA{}, fmt.Errorf("invalid color %q: expected 6 or 8 hex digits (e.g., #RRGGBB or #RRGGBBAA)", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid red component in %q: %w", hex, err)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid green component in %q: %w", hex, err)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid blue component in %q: %w", hex, err)
	}

	a := uint8(255)
	if len(hex) == 8 {
		av, err := strconv.ParseUint(hex[6:8], 16, 8)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid alpha component in %q: %w", hex, err)
		}
		a = uint8(av)
	}

	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: a}, nil
}

// terminalWriter implements qrcode.Writer to render QR codes to stdout
// using Unicode half-block characters for compact display.
type terminalWriter struct{}

var _ qrcode.Writer = (*terminalWriter)(nil)

func (w *terminalWriter) Write(mat qrcode.Matrix) error {
	bitmap := mat.Bitmap()
	height := len(bitmap)
	width := len(bitmap[0])

	// Each terminal row represents 2 QR rows using half-block characters
	for row := 0; row < height; row += 2 {
		var line strings.Builder
		for col := 0; col < width; col++ {
			topSet := bitmap[row][col]
			bottomSet := false
			if row+1 < height {
				bottomSet = bitmap[row+1][col]
			}

			switch {
			case topSet && bottomSet:
				line.WriteRune('\u2588') // █ full block
			case topSet:
				line.WriteRune('\u2580') // ▀ upper half block
			case bottomSet:
				line.WriteRune('\u2584') // ▄ lower half block
			default:
				line.WriteRune(' ')
			}
		}
		fmt.Println(line.String())
	}
	return nil
}

func (w *terminalWriter) Close() error {
	return nil
}

func printToTerminal(content string) {
	qrc, err := qrcode.New(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not generate QR code: %v\n", err)
		os.Exit(1)
	}

	w := &terminalWriter{}
	if err := qrc.Save(w); err != nil {
		fmt.Fprintf(os.Stderr, "error: could not print QR code: %v\n", err)
		os.Exit(1)
	}
}

func saveToFile(content, outputPath, fgColor, bgColor, shape string,
	qrWidth uint, borderTop, borderRight, borderBottom, borderLeft int) {

	// Parse colors
	fg, err := parseColor(fgColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	bg, err := parseColor(bgColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Determine format: use PNG when any color has transparency
	hasAlpha := fg.A < 255 || bg.A < 255
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(outputPath), "."))

	opts := []standard.ImageOption{
		standard.WithQRWidth(uint8(qrWidth)),
		standard.WithBorderWidth(borderTop, borderRight, borderBottom, borderLeft),
		standard.WithFgColor(fg),
		standard.WithBgColor(bg),
	}

	if hasAlpha || ext == "png" {
		opts = append(opts, standard.WithBuiltinImageEncoder(standard.PNG_FORMAT))
	} else {
		opts = append(opts, standard.WithBuiltinImageEncoder(standard.JPEG_FORMAT))
	}

	var cellShape shapes.IShape
	switch shape {
	case "rect":
		cellShape = shapes.Rect
	case "circle":
		cellShape = shapes.Circ
	case "diamond":
		cellShape = shapes.DiamondS
	case "star":
		cellShape = shapes.StarS
	case "dot":
		cellShape = shapes.SmallDot
	case "rounded":
		cellShape = shapes.RoundedS
	case "liquid":
		cellShape = shapes.LiquidS
	case "squircle":
		cellShape = shapes.SquircleS
	case "rounded-na":
		cellShape = shapes.RoundedNAShape
	default:
		fmt.Fprintf(os.Stderr, "error: unknown shape %q\n", shape)
		os.Exit(1)
	}

	opts = append(opts, standard.WithCustomShape(cellShape))

	// Create the QR code
	qrc, err := qrcode.New(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not generate QR code: %v\n", err)
		os.Exit(1)
	}

	// Create the writer and save
	writer, err := standard.New(outputPath, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not create image writer: %v\n", err)
		os.Exit(1)
	}
	defer writer.Close()

	if err := qrc.Save(writer); err != nil {
		fmt.Fprintf(os.Stderr, "error: could not save QR code: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "QR code saved to %s\n", outputPath)
}
