package ansihtml

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type colorMode int32

const (
	cmDEFAULT colorMode = iota
	cmP16
	cmP256
	cmRGB
)

type Theme struct {
	Foreground    string
	Background    string
	Black         string
	Red           string
	Green         string
	Yellow        string
	Blue          string
	Magenta       string
	Cyan          string
	White         string
	Gray          string
	BrightRed     string
	BrightGreen   string
	BrightYellow  string
	BrightBlue    string
	BrightMagenta string
	BrightCyan    string
	BrightWhite   string
}

func toRgb(r rune, g rune, b rune) rune {
	return (r << 16) | (g << 8) | b
}

func toCSS(rgb int32) string {
	return fmt.Sprintf("#%06x", rgb)
}

// https://www.w3.org/TR/WCAG20/#relativeluminancedef
func relativeLuminance(rgb rune) float64 {
	return _relativeLuminance(float64((rgb>>16)&0xff), float64((rgb>>8)&0xff), float64(rgb&0xff))
}

func _relativeLuminance(r float64, g float64, b float64) float64 {
	c := [3]float64{r / 255.0, g / 255.0, b / 255.0}
	for i := 0; i < 3; i++ {
		if c[i] <= 0.03928 {
			c[i] = c[i] / 12.92
		} else {
			c[i] = math.Pow((c[i]+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*c[0] + 0.7152*c[1] + 0.0722*c[2]
}

func contrastRatio(l1 float64, l2 float64) float64 {
	if l1 < l2 {
		return (l2 + 0.05) / (l1 + 0.05)
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func ensureContrastRatio(fgRgb rune, bgRgb rune, ratio float64) (rune, bool) {
	fgL := relativeLuminance(fgRgb)
	bgL := relativeLuminance(bgRgb)
	r := contrastRatio(fgL, bgL)
	if r < ratio {
		if fgL < bgL {
			return reduceLuminance(fgRgb, bgRgb, ratio), true
		}
		return increaseLuminance(fgRgb, bgRgb, ratio), true
	}
	return 0, false
}

func reduceLuminance(fgRgb rune, bgRgb rune, ratio float64) rune {
	bgL := relativeLuminance(bgRgb)
	fgR := float64((fgRgb >> 16) & 0xff)
	fgG := float64((fgRgb >> 8) & 0xff)
	fgB := float64(fgRgb & 0xff)
	const p = 0.1
	r := contrastRatio(_relativeLuminance(fgR, fgG, fgB), bgL)
	for r < ratio && (fgR > 0 || fgG > 0 || fgB > 0) {
		fgR = fgR - math.Max(0, math.Ceil(fgR*p))
		fgG = fgG - math.Max(0, math.Ceil(fgG*p))
		fgB = fgB - math.Max(0, math.Ceil(fgB*p))
		r = contrastRatio(_relativeLuminance(fgR, fgB, fgG), bgL)
	}

	return (rune(fgR) << 16) | (rune(fgG) << 8) | rune(fgB)
}

func increaseLuminance(fgRgb rune, bgRgb rune, ratio float64) rune {
	bgL := relativeLuminance(bgRgb)
	fgR := float64((fgRgb >> 16) & 0xff)
	fgG := float64((fgRgb >> 8) & 0xff)
	fgB := float64(fgRgb & 0xff)
	const p = 0.1
	r := contrastRatio(_relativeLuminance(fgR, fgB, fgG), bgL)
	for r < ratio && (fgR < 255.0 || fgG < 255.0 || fgB < 255.0) {
		fgR = math.Min(0xff, fgR+math.Ceil((255-fgR)*p))
		fgG = math.Min(0xff, fgG+math.Ceil((255-fgG)*p))
		fgB = math.Min(0xff, fgB+math.Ceil((255-fgB)*p))
		r = contrastRatio(_relativeLuminance(fgR, fgB, fgG), bgL)
	}
	return (rune(fgR) << 16) | (rune(fgG) << 8) | rune(fgB)
}

type colorObject struct {
	rgb rune
	css string
}

type palette struct {
	foreground *colorObject
	background *colorObject
	colors     []*colorObject
}

var defaultColors = [256]string{
	"#3f4451", // 30
	"#e05561", // 31
	"#8cc265", // 32
	"#d18f52", // 33
	"#4aa5f0", // 34
	"#c162de", // 35
	"#42b3c2", // 36
	"#e6e6e6", // 37
	// bright colors
	"#4f5666", // 90
	"#ff616e", // 91
	"#a5e075", // 92
	"#f0a45d", // 93
	"#4dc4ff", // 94
	"#de73ff", // 95
	"#4cd1e0", // 96
	"#d7dae0", // 97
	// 256-colors
	"#000000",
	"#00005f",
	"#000087",
	"#0000af",
	"#0000d7",
	"#0000ff",
	"#005f00",
	"#005f5f",
	"#005f87",
	"#005faf",
	"#005fd7",
	"#005fff",
	"#008700",
	"#00875f",
	"#008787",
	"#0087af",
	"#0087d7",
	"#0087ff",
	"#00af00",
	"#00af5f",
	"#00af87",
	"#00afaf",
	"#00afd7",
	"#00afff",
	"#00d700",
	"#00d75f",
	"#00d787",
	"#00d7af",
	"#00d7d7",
	"#00d7ff",
	"#00ff00",
	"#00ff5f",
	"#00ff87",
	"#00ffaf",
	"#00ffd7",
	"#00ffff",
	"#5f0000",
	"#5f005f",
	"#5f0087",
	"#5f00af",
	"#5f00d7",
	"#5f00ff",
	"#5f5f00",
	"#5f5f5f",
	"#5f5f87",
	"#5f5faf",
	"#5f5fd7",
	"#5f5fff",
	"#5f8700",
	"#5f875f",
	"#5f8787",
	"#5f87af",
	"#5f87d7",
	"#5f87ff",
	"#5faf00",
	"#5faf5f",
	"#5faf87",
	"#5fafaf",
	"#5fafd7",
	"#5fafff",
	"#5fd700",
	"#5fd75f",
	"#5fd787",
	"#5fd7af",
	"#5fd7d7",
	"#5fd7ff",
	"#5fff00",
	"#5fff5f",
	"#5fff87",
	"#5fffaf",
	"#5fffd7",
	"#5fffff",
	"#870000",
	"#87005f",
	"#870087",
	"#8700af",
	"#8700d7",
	"#8700ff",
	"#875f00",
	"#875f5f",
	"#875f87",
	"#875faf",
	"#875fd7",
	"#875fff",
	"#878700",
	"#87875f",
	"#878787",
	"#8787af",
	"#8787d7",
	"#8787ff",
	"#87af00",
	"#87af5f",
	"#87af87",
	"#87afaf",
	"#87afd7",
	"#87afff",
	"#87d700",
	"#87d75f",
	"#87d787",
	"#87d7af",
	"#87d7d7",
	"#87d7ff",
	"#87ff00",
	"#87ff5f",
	"#87ff87",
	"#87ffaf",
	"#87ffd7",
	"#87ffff",
	"#af0000",
	"#af005f",
	"#af0087",
	"#af00af",
	"#af00d7",
	"#af00ff",
	"#af5f00",
	"#af5f5f",
	"#af5f87",
	"#af5faf",
	"#af5fd7",
	"#af5fff",
	"#af8700",
	"#af875f",
	"#af8787",
	"#af87af",
	"#af87d7",
	"#af87ff",
	"#afaf00",
	"#afaf5f",
	"#afaf87",
	"#afafaf",
	"#afafd7",
	"#afafff",
	"#afd700",
	"#afd75f",
	"#afd787",
	"#afd7af",
	"#afd7d7",
	"#afd7ff",
	"#afff00",
	"#afff5f",
	"#afff87",
	"#afffaf",
	"#afffd7",
	"#afffff",
	"#d70000",
	"#d7005f",
	"#d70087",
	"#d700af",
	"#d700d7",
	"#d700ff",
	"#d75f00",
	"#d75f5f",
	"#d75f87",
	"#d75faf",
	"#d75fd7",
	"#d75fff",
	"#d78700",
	"#d7875f",
	"#d78787",
	"#d787af",
	"#d787d7",
	"#d787ff",
	"#d7af00",
	"#d7af5f",
	"#d7af87",
	"#d7afaf",
	"#d7afd7",
	"#d7afff",
	"#d7d700",
	"#d7d75f",
	"#d7d787",
	"#d7d7af",
	"#d7d7d7",
	"#d7d7ff",
	"#d7ff00",
	"#d7ff5f",
	"#d7ff87",
	"#d7ffaf",
	"#d7ffd7",
	"#d7ffff",
	"#ff0000",
	"#ff005f",
	"#ff0087",
	"#ff00af",
	"#ff00d7",
	"#ff00ff",
	"#ff5f00",
	"#ff5f5f",
	"#ff5f87",
	"#ff5faf",
	"#ff5fd7",
	"#ff5fff",
	"#ff8700",
	"#ff875f",
	"#ff8787",
	"#ff87af",
	"#ff87d7",
	"#ff87ff",
	"#ffaf00",
	"#ffaf5f",
	"#ffaf87",
	"#ffafaf",
	"#ffafd7",
	"#ffafff",
	"#ffd700",
	"#ffd75f",
	"#ffd787",
	"#ffd7af",
	"#ffd7d7",
	"#ffd7ff",
	"#ffff00",
	"#ffff5f",
	"#ffff87",
	"#ffffaf",
	"#ffffd7",
	"#ffffff",
	"#080808",
	"#121212",
	"#1c1c1c",
	"#262626",
	"#303030",
	"#3a3a3a",
	"#444444",
	"#4e4e4e",
	"#585858",
	"#626262",
	"#6c6c6c",
	"#767676",
	"#808080",
	"#8a8a8a",
	"#949494",
	"#9e9e9e",
	"#a8a8a8",
	"#b2b2b2",
	"#bcbcbc",
	"#c6c6c6",
	"#d0d0d0",
	"#dadada",
	"#e4e4e4",
	"#eeeeee",
}

func toColorObject(hexRgb string) *colorObject {
	parseHexColor := func(rgb string) (rune, bool) {
		if strings.HasPrefix(rgb, "#") {
			rgb = rgb[1:]
		}
		switch len(rgb) {
		case 8:
			rgb = rgb[:6]
		case 6:
		case 4:
			rgb = strings.Repeat(rgb[0:1], 2) + strings.Repeat(rgb[1:2], 2) + strings.Repeat(rgb[2:3], 2)
		case 3:
			rgb = strings.Repeat(rgb[0:1], 2) + strings.Repeat(rgb[1:2], 2) + strings.Repeat(rgb[2:3], 2)
		default:
			return 0, false
		}
		xxx, err := strconv.ParseInt(rgb, 16, 32)
		if err != nil {
			return 0, false
		}
		return rune(xxx), true
	}
	rgb, ok := parseHexColor(hexRgb)
	if !ok {
		return nil
	}
	return &colorObject{
		rgb: rgb,
		css: toCSS(rgb),
	}
}

func buildDefaultPalette() palette {
	colors := [256]*colorObject{}
	for i := 0; i < len(defaultColors); i++ {
		colors[i] = toColorObject(defaultColors[i])
	}
	return palette{
		foreground: nil,
		background: nil,
		colors:     colors[:],
	}
}

func buildPalette(theme Theme) palette {
	def := func(userColor string, defaultValue string) *colorObject {
		if userColor != "" {
			return toColorObject(userColor)
		}
		return toColorObject(defaultValue)
	}
	colors := [256]*colorObject{
		def(theme.Black, defaultColors[0]),
		def(theme.Red, defaultColors[1]),
		def(theme.Green, defaultColors[2]),
		def(theme.Yellow, defaultColors[3]),
		def(theme.Blue, defaultColors[4]),
		def(theme.Magenta, defaultColors[5]),
		def(theme.Cyan, defaultColors[6]),
		def(theme.White, defaultColors[7]),
		def(theme.Gray, defaultColors[8]),
		def(theme.BrightRed, defaultColors[9]),
		def(theme.BrightGreen, defaultColors[10]),
		def(theme.BrightYellow, defaultColors[11]),
		def(theme.BrightBlue, defaultColors[12]),
		def(theme.BrightMagenta, defaultColors[13]),
		def(theme.BrightCyan, defaultColors[14]),
		def(theme.BrightWhite, defaultColors[15]),
	}

	for i := 16; i < len(defaultColors); i++ {
		colors[i] = toColorObject(defaultColors[i])
	}

	return palette{
		foreground: def(theme.Foreground, ""),
		background: def(theme.Background, ""),
		colors:     colors[:],
	}
}

// /** @return rgba */
// export function blend(fg_rgba: number, bg_rgba: number): number {
// 	const alpha = (fg_rgba & 0xff) / 255
// 	if (alpha >= 1) return fg_rgba

// 	const fgR = (fg_rgba >> 24) & 0xff
// 	const fgG = (fg_rgba >> 16) & 0xff
// 	const fgB = (fg_rgba >> 8) & 0xff
// 	const bgR = (bg_rgba >> 24) & 0xff
// 	const bgG = (bg_rgba >> 16) & 0xff
// 	const bgB = (bg_rgba >> 8) & 0xff
// 	const bg_alpha = (bg_rgba & 0xff) / 255

// 	const r = Math.round(fgR * alpha + bgR * (1 - alpha) * bg_alpha)
// 	const g = Math.round(fgG * alpha + bgG * (1 - alpha) * bg_alpha)
// 	const b = Math.round(fgB * alpha + bgB * (1 - alpha) * bg_alpha)
// 	const a = alpha + (1 - alpha) * bg_alpha
// 	return toRgba(r, g, b, a)
// 	function toRgba(r: number, g: number, b: number, a = 0xff): number {
// 		return (r << 24) | (g << 16) | (b << 8) | a
// 	}
// }
