package service

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
)

// pdfDoc builds a minimal, valid single-page PDF document with no external
// dependencies (pure Go, CGO-free). It supports filled rectangles (used to
// render QR modules as vector graphics) and Helvetica text lines.
type pdfDoc struct {
	width   float64
	height  float64
	content bytes.Buffer
}

// newPDFDoc creates a PDF document with the given page size in PDF points.
func newPDFDoc(width, height float64) *pdfDoc {
	return &pdfDoc{width: width, height: height}
}

// setFillColor sets the non-stroking (fill) color. Components are 0.0-1.0.
func (d *pdfDoc) setFillColor(r, g, b float64) {
	fmt.Fprintf(&d.content, "%.4f %.4f %.4f rg\n", r, g, b)
}

// fillRect fills a rectangle. PDF coordinates originate at the bottom-left.
func (d *pdfDoc) fillRect(x, y, w, h float64) {
	fmt.Fprintf(&d.content, "%.3f %.3f %.3f %.3f re f\n", x, y, w, h)
}

// text draws a single line of Helvetica text at the given baseline position.
func (d *pdfDoc) text(x, y, size float64, s string) {
	fmt.Fprintf(&d.content, "BT /F1 %.2f Tf %.2f %.2f Td (%s) Tj ET\n", size, x, y, pdfEscapeString(s))
}

// render serializes the document into valid PDF bytes with a correct xref table.
func (d *pdfDoc) render() []byte {
	content := d.content.String()

	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>", d.width, d.height),
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objs)+1)
	for i, o := range objs {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}

	xrefPos := buf.Len()
	count := len(objs) + 1
	fmt.Fprintf(&buf, "xref\n0 %d\n", count)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < count; i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", count, xrefPos)

	return buf.Bytes()
}

// pdfEscapeString escapes a string for a PDF literal string object. Characters
// outside the printable ASCII range are dropped to keep the output valid.
func pdfEscapeString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '(':
			b.WriteString(`\(`)
		case ')':
			b.WriteString(`\)`)
		default:
			if r >= 0x20 && r < 0x7f {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// colorToUnit converts a color.Color into 0.0-1.0 RGB components for PDF.
func colorToUnit(c color.Color) (float64, float64, float64) {
	r, g, b, _ := c.RGBA()
	return float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0
}
