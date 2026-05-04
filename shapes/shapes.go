// Package shapes provides composable QR code cell shapes that implement
// the standard.IShape interface from go-qrcode.
//
// Shapes can adapt their rendering based on neighboring blocks via the
// DrawContext.Neighbours() bitmask, enabling context-aware designs like
// rounded corners and merged edges.
package shapes

import (
	"math"

	"github.com/yeqown/go-qrcode/writer/standard"
)

// IShape is an alias for the go-qrcode shape interface.
type IShape = standard.IShape

// Neighbour bit flags for the 8 surrounding cells in a 3x3 grid.
const (
	nTopLeft uint16 = 1 << iota // top-left
	nTop                        // top
	nTopRight                   // top-right
	nLeft                       // left
	nSelf                       // center (self)
	nRight                      // right
	nBotLeft                    // bottom-left
	nBot                        // bottom
	nBotRight                   // bottom-right
)

func hasNeighbour(neighbours uint16, flags ...uint16) bool {
	for _, f := range flags {
		if neighbours&f != 0 {
			return true
		}
	}
	return false
}

// Rectangle is the default rectangular cell shape.
type Rectangle struct{}

func (Rectangle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	ctx.DrawRectangle(x, y, float64(w), float64(h))
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (Rectangle) DrawFinder(ctx *standard.DrawContext) {
	Rectangle{}.Draw(ctx)
}

// Circle draws a full circle per cell.
type Circle struct{}

func (Circle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	radius := min(w, h) / 2
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0
	ctx.DrawCircle(cx, cy, float64(radius))
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (Circle) DrawFinder(ctx *standard.DrawContext) {
	Circle{}.Draw(ctx)
}

// Dot draws a small circle with a configurable radius percentage (0.0–1.0).
type Dot struct {
	Radius float64
}

func (d Dot) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	radius := min(w, h) / 2
	radius = int(float64(radius) * d.Radius)
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0
	ctx.DrawCircle(cx, cy, float64(radius))
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (d Dot) DrawFinder(ctx *standard.DrawContext) {
	full := Dot{Radius: 1.0}
	full.Draw(ctx)
}

// Rounded draws rounded rectangles that adapt based on neighbours.
// Outer corners get the full radius; shared edges use smaller radii.
type Rounded struct {
	CornerRadius float64
}

func (r Rounded) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	neighbours := ctx.Neighbours()

	baseR := float64(min(w, h)) * r.CornerRadius

	topOpen := !hasNeighbour(neighbours, nTop, nTopLeft, nTopRight)
	botOpen := !hasNeighbour(neighbours, nBot, nBotLeft, nBotRight)
	leftOpen := !hasNeighbour(neighbours, nLeft, nTopLeft, nBotLeft)
	rightOpen := !hasNeighbour(neighbours, nRight, nTopRight, nBotRight)

	rTL := cornerRadius(baseR, topOpen, leftOpen)
	rTR := cornerRadius(baseR, topOpen, rightOpen)
	rBR := cornerRadius(baseR, botOpen, rightOpen)
	rBL := cornerRadius(baseR, botOpen, leftOpen)

	maxR := max4(rTL, rTR, rBR, rBL)

	if maxR == 0 {
		ctx.DrawRectangle(x, y, float64(w), float64(h))
	} else {
		ctx.DrawRoundedRectangle(x, y, float64(w), float64(h), maxR)
	}
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (r Rounded) DrawFinder(ctx *standard.DrawContext) {
	Rectangle{}.Draw(ctx)
}

func cornerRadius(base float64, a, b bool) float64 {
	if a && b {
		return base
	}
	if a || b {
		return base / 2.0
	}
	return 0
}

func max4(a, b, c, d float64) float64 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	return m
}

// Diamond draws a rotated square (rhombus) per cell.
type Diamond struct{}

func (Diamond) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0

	ctx.MoveTo(cx, y)
	ctx.LineTo(x+float64(w), cy)
	ctx.LineTo(cx, y+float64(h))
	ctx.LineTo(x, cy)
	ctx.ClosePath()
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (Diamond) DrawFinder(ctx *standard.DrawContext) {
	Rectangle{}.Draw(ctx)
}

// Star draws a 5-pointed star per cell.
type Star struct{}

func (Star) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0
	outerR := float64(min(w, h)) / 2.0
	innerR := outerR * 0.4

	for i := 0; i < 10; i++ {
		r := outerR
		if i%2 == 1 {
			r = innerR
		}
		angle := math.Pi * float64(i) / 5.0 - math.Pi/2.0
		pX := cx + r*math.Cos(angle)
		pY := cy + r*math.Sin(angle)
		if i == 0 {
			ctx.MoveTo(pX, pY)
		} else {
			ctx.LineTo(pX, pY)
		}
	}
	ctx.ClosePath()
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (Star) DrawFinder(ctx *standard.DrawContext) {
	Rectangle{}.Draw(ctx)
}

// Squircle draws a superellipse (smooth between circle and rectangle).
// Squint: 0.0 = circle, 1.0 = rectangle.
type Squircle struct {
	Squint float64
}

func (s Squircle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0
	rx, ry := float64(w)/2.0, float64(h)/2.0

	n := 2.0 + (1.0-s.Squint)*10.0

	steps := 64
	for i := 0; i < steps; i++ {
		t := 2.0 * math.Pi * float64(i) / float64(steps)
		cosT := math.Abs(math.Cos(t))
		sinT := math.Abs(math.Sin(t))
		px := cx + rx*math.Pow(cosT, 2.0/n)*signF(math.Cos(t))
		py := cy + ry*math.Pow(sinT, 2.0/n)*signF(math.Sin(t))
		if i == 0 {
			ctx.MoveTo(px, py)
		} else {
			ctx.LineTo(px, py)
		}
	}
	ctx.ClosePath()
	ctx.SetColor(ctx.Color())
	ctx.Fill()
}

func (s Squircle) DrawFinder(ctx *standard.DrawContext) {
	Rectangle{}.Draw(ctx)
}

func signF(f float64) float64 {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

// NeighborAware wraps any shape and scales it based on the number of
// set neighbours. Fewer neighbours = smaller cell, more = larger.
type NeighborAware struct {
	Shape    IShape
	MinScale float64
	MaxScale float64
}

func (na NeighborAware) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	neighbours := ctx.Neighbours()

	count := 0
	for i := uint16(1); i < 256; i <<= 1 {
		if i != nSelf && neighbours&i != 0 {
			count++
		}
	}

	scale := na.MinScale + (na.MaxScale-na.MinScale)*float64(count)/8.0
	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0

	// Scale the gg context around the center, draw, then restore.
	ctx.ScaleAbout(scale, scale, cx, cy)
	na.Shape.Draw(ctx)
	ctx.ScaleAbout(1.0/scale, 1.0/scale, cx, cy)
}

func (na NeighborAware) DrawFinder(ctx *standard.DrawContext) {
	na.Shape.Draw(ctx)
}

// Builder provides a fluent API for constructing shapes.
type Builder struct {
	baseShape IShape
	naMin     float64
	naMax     float64
	naSet     bool
}

func NewBuilder() *Builder {
	return &Builder{
		baseShape: Rectangle{},
		naMin:     0.5,
		naMax:     1.0,
	}
}

func (b *Builder) Shape(s IShape) *Builder {
	b.baseShape = s
	return b
}

func (b *Builder) Dot(radius float64) *Builder {
	b.baseShape = &Dot{Radius: radius}
	return b
}

func (b *Builder) Rounded(cornerRadius float64) *Builder {
	b.baseShape = &Rounded{CornerRadius: cornerRadius}
	return b
}

func (b *Builder) Squircle(squint float64) *Builder {
	b.baseShape = &Squircle{Squint: squint}
	return b
}

func (b *Builder) NeighborAware(minScale, maxScale float64) *Builder {
	b.naMin = minScale
	b.naMax = maxScale
	b.naSet = true
	return b
}

func (b *Builder) Build() IShape {
	if b.naSet {
		return NeighborAware{
			Shape:    b.baseShape,
			MinScale: b.naMin,
			MaxScale: b.naMax,
		}
	}
	return b.baseShape
}

// Predefined shapes for quick access.
var (
	Rect           IShape = Rectangle{}
	Circ           IShape = Circle{}
	DiamondS       IShape = Diamond{}
	StarS          IShape = Star{}
	SmallDot       IShape = Dot{Radius: 0.5}
	RoundedS       IShape = Rounded{CornerRadius: 0.3}
	SquircleS      IShape = Squircle{Squint: 0.5}
	RoundedNAShape IShape = NewBuilder().Rounded(0.3).NeighborAware(0.5, 1.0).Build()
)
