package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/qpliu/qrencode-go/qrencode"
)

type RenderOptions struct {
	Scale         int
	Padding       int
	Center        bool
	TerminalWidth int
}

func main() {
	scale := flag.Int("scale", 1, "QR scale (>=1)")
	padding := flag.Int("padding", 0, "QR padding (>=0)")
	flag.Parse()

	text := strings.Join(flag.Args(), " ")

	// If no args, try piped stdin.
	if strings.TrimSpace(text) == "" {
		if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			text = strings.TrimSpace(string(b))
		}
	}

	if text == "" {
		panic("no input: pass text as args or pipe it via stdin")
	}

	grid, err := qrencode.Encode(text, qrencode.ECLevelQ)
	if err != nil {
		panic(err)
	}

	err = PrintQRCode(os.Stdout, grid, RenderOptions{
		Scale:   *scale,
		Padding: *padding,
		Center:  true,
	})
	if err != nil {
		panic(err)
	}
}

func PrintQRCode(w io.Writer, g *qrencode.BitGrid, o RenderOptions) error {
	var buffer bytes.Buffer
	if o.Scale < 1 {
		o.Scale = 1
	}
	if o.Padding < 0 {
		o.Padding = 0
	}
	if o.TerminalWidth <= 0 {
		o.TerminalWidth, _ = strconv.Atoi(os.Getenv("COLUMNS"))
	}

	renderW := (g.Width() + o.Padding*2) * o.Scale
	renderH := (g.Height() + o.Padding*2) * o.Scale

	left := 0
	if o.Center && o.TerminalWidth > renderW {
		left = (o.TerminalWidth - renderW) / 2
	}
	margin := strings.Repeat(" ", left)

	isDark := func(rx, ry int) bool {
		x := rx/o.Scale - o.Padding
		y := ry/o.Scale - o.Padding
		return x >= 0 && y >= 0 && x < g.Width() && y < g.Height() && g.Get(x, y)
	}

	for y := 0; y < renderH; y += 2 {
		if _, err := buffer.WriteString(margin); err != nil {
			return err
		}

		for x := 0; x < renderW; x++ {
			top := isDark(x, y)
			bottom := y+1 < renderH && isDark(x, y+1)

			ch := " "
			if top && bottom {
				ch = "█"
			} else if top {
				ch = "▀"
			} else if bottom {
				ch = "▄"
			}

			if _, err := buffer.WriteString(ch); err != nil {
				return err
			}
		}
		if _, err := buffer.WriteString("\n"); err != nil {
			return err
		}
	}
	fmt.Println(buffer.String())
	return nil
}
