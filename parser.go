package ansihtml

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Mode int

const (
	Inline Mode = iota
	Class
)

type attributes struct {
	fgIndexOrRgb rune
	bgIndexOrRgb rune
	fgMode       colorMode
	bgMode       colorMode
	bold         bool
	dim          bool
	underline    bool
	inverse      bool
	italic       bool
	strike       bool
	hidden       bool
}

type spanStyle struct {
	fgMode     colorMode
	foreground string
	bgMode     colorMode
	background string
	bold       bool
	dim        bool
	underline  bool
	italic     bool
	strike     bool
	hidden     bool
}

type anchor struct {
	url    string
	params map[string]string
}

type Converter struct {
	r                    *renderer
	minimumContrastRatio float64
	mode                 Mode
	classPrefix          string
	palette              palette
	contrastCache        *contrastCache
	prevStyle            *spanStyle
	prevAnchor           *anchor
	isSpan               bool
	styleChanged         bool
	isAnchor             bool
	attributes
}

func NewConverter(options ...Option) *Converter {
	c := &Converter{
		minimumContrastRatio: 3,
		palette:              buildDefaultPalette(),
		mode:                 Inline,
		classPrefix:          "ansi-",
		contrastCache:        newContrastCache(),
		styleChanged:         true,
		attributes: attributes{
			fgIndexOrRgb: -1,
			bgIndexOrRgb: -1,
			fgMode:       cmDEFAULT,
			bgMode:       cmDEFAULT,
			bold:         false,
			dim:          false,
			underline:    false,
			inverse:      false,
			italic:       false,
			strike:       false,
			hidden:       false,
		},
	}
	c.ApplyOptions(options...)
	return c
}

func (c *Converter) ApplyOptions(options ...Option) {
	for _, opt := range options {
		if opt != nil {
			opt(c)
		}
	}
	c.r = &renderer{classPrefix: c.classPrefix, isClass: c.mode == Class, defaultFg: c.palette.foreground}
}

func (c *Converter) Copy(dst io.Writer, src io.Reader) error {
	return c.CopyWithContext(context.Background(), dst, src)
}

func (c *Converter) CopyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	w := bufio.NewWriter(dst)
	r := bufio.NewReader(src)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		char, _, err := r.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch char {
		case xESC:
			nextChar, _, err := r.ReadRune()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if nextChar == xLeftSquareBracket {
				if err := c.readCSI(r); err != nil {
					return err
				}
				c.styleChanged = true
			} else if nextChar == xRightSquareBracket {
				if err := c.readOSC(r, w); err != nil {
					return err
				}
			}
			continue
		}

		switch char {
		case xBS:
			fallthrough
		case xVT:
			fallthrough
		case xDEL:
			continue
		}

		if err := c.rune(w, char); err != nil {
			return err
		}
	}
	if c.isSpan {
		if _, err := c.r.spanClose(w); err != nil {
			return err
		}
	}
	if c.isAnchor {
		if _, err := c.r.anchorClose(w); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (c *Converter) Reset() {
	c.contrastCache.Clear()
	c.prevStyle = nil
	c.prevAnchor = nil
	c.styleChanged = true
	c.isSpan = false
	c.isAnchor = false
	c.attributes = attributes{
		fgIndexOrRgb: -1,
		bgIndexOrRgb: -1,
		fgMode:       cmDEFAULT,
		bgMode:       cmDEFAULT,
		bold:         false,
		dim:          false,
		underline:    false,
		inverse:      false,
		italic:       false,
		strike:       false,
		hidden:       false,
	}
}

func (c *Converter) gatherStyle() (*spanStyle, error) {
	fgColor := c.fgIndexOrRgb
	bgColor := c.bgIndexOrRgb
	fgMode := c.fgMode
	bgMode := c.bgMode

	if c.inverse {
		fgColor, bgColor = bgColor, fgColor
		fgMode, bgMode = bgMode, fgMode
	}

	if fgMode == cmP256 && fgColor < 16 {
		fgMode = cmP16
	}
	if bgMode == cmP256 && bgColor < 16 {
		bgMode = cmP16
	}

	var err error
	var foreground string
	if c.mode == Class && fgMode != cmRGB {
		foreground = c.getForegroundClass(fgMode, fgColor)
	} else {
		foreground, err = c.getForegroundCSS(bgMode, bgColor, fgMode, fgColor)
	}
	if err != nil {
		return nil, err
	}

	var background string
	if c.mode == Class && bgMode != cmRGB {
		background = c.getBackgroundClass(bgMode, bgColor)
	} else {
		background, err = c.getBackgroundCSS(bgMode, bgColor)
	}
	if err != nil {
		return nil, err
	}

	style := &spanStyle{
		fgMode:     fgMode,
		foreground: foreground,
		bgMode:     bgMode,
		background: background,
		bold:       c.bold,
		dim:        c.dim,
		italic:     c.italic,
		underline:  c.underline,
		hidden:     c.hidden,
		strike:     c.strike,
	}

	return style, nil
}

func (c *Converter) rune(w writer, char rune) error {
	if c.styleChanged {
		style, err := c.gatherStyle()
		if err != nil {
			return err
		}
		if !equalStyle(c.prevStyle, style) {
			if needStyle(c.prevStyle) {
				if _, err := c.r.spanClose(w); err != nil {
					return err
				}
			}
			c.isSpan = needStyle(style)
			c.prevStyle = style
			if c.isSpan {
				if _, err := c.r.spanOpen(w, style); err != nil {
					return err
				}
			}
		}
		c.styleChanged = false
	}
	if _, err := c.r.rune(w, char); err != nil {
		return err
	}
	return w.Flush()
}

func equalStyle(a *spanStyle, b *spanStyle) bool {
	if a == nil {
		return !needStyle(b)
	}
	if b == nil {
		return !needStyle(a)
	}

	return a.foreground == b.foreground &&
		a.background == b.background &&
		a.bold == b.bold &&
		a.dim == b.dim &&
		a.italic == b.italic &&
		a.underline == b.underline &&
		a.hidden == b.hidden &&
		a.strike == b.strike
}

func needStyle(s *spanStyle) bool {
	if s == nil {
		return false
	}
	return s.foreground != "" ||
		s.background != "" ||
		s.bold ||
		s.dim ||
		s.italic ||
		s.underline ||
		s.hidden ||
		s.strike
}

func (c *Converter) resetAttributes() {
	c.fgIndexOrRgb = -1
	c.bgIndexOrRgb = -1
	c.fgMode = cmDEFAULT
	c.bgMode = cmDEFAULT
	c.bold = false
	c.dim = false
	c.italic = false
	c.underline = false
	c.inverse = false
	c.hidden = false
	c.strike = false
}

func (c *Converter) setAttributes(attrs []rune) {
	for i := 0; i < len(attrs); i++ {
		a := attrs[i]
		switch a {
		case yReset:
			c.resetAttributes()
		case yBold:
			c.bold = true
		case yDim:
			c.dim = true
		case yItalic:
			c.italic = true
		case yUnderline:
			c.underline = true
		case yInverse:
			c.inverse = true
		case yHidden:
			c.hidden = true
		case yStrike:
			c.strike = true
		case ySlowBlink:
		case yRapidBlink:
		}
		if a >= yFgBlack && a <= yFgWhite {
			c.fgIndexOrRgb = a - yFgBlack
			c.fgMode = cmP16
		} else if a >= yBgBlack && a <= yBgWhite {
			c.bgIndexOrRgb = a - yBgBlack
			c.bgMode = cmP16
		} else if a >= yBrightFgGray && a <= yBrightFgWhite {
			c.fgIndexOrRgb = 8 + a - yBrightFgGray
			c.fgMode = cmP16
		} else if a >= yBrightBgGray && a <= yBrightBgWhite {
			c.bgIndexOrRgb = 8 + a - yBrightBgGray
			c.bgMode = cmP16
		} else if a == yFgReset {
			c.fgIndexOrRgb = -1
			c.fgMode = cmDEFAULT
		} else if a == yBgReset {
			c.bgIndexOrRgb = -1
			c.bgMode = cmDEFAULT
		} else if a == yFgExt {
			if attrs[i+1] == 5 {
				c.fgMode = cmP256
				if i+2 >= len(attrs) {
					c.fgIndexOrRgb = 0
					break
				}
				c.fgIndexOrRgb = attrs[i+2]
				i += 2
			} else if attrs[i+1] == 2 {
				c.fgMode = cmRGB
				if i+4 >= len(attrs) {
					c.fgIndexOrRgb = 0
					break
				}
				c.fgIndexOrRgb = toRgb(attrs[i+2], attrs[i+3], attrs[i+4])

				i += 4
			}
		} else if a == yBgExt {
			if attrs[i+1] == 5 {
				c.bgMode = cmP256
				if i+2 >= len(attrs) {
					c.bgIndexOrRgb = 0
					break
				}
				c.bgIndexOrRgb = attrs[i+2]
				i += 2
			} else if attrs[i+1] == 2 {
				c.bgMode = cmRGB
				if i+4 >= len(attrs) {
					c.bgIndexOrRgb = 0
					break
				}
				c.bgIndexOrRgb = toRgb(attrs[i+2], attrs[i+3], attrs[i+4])
				i += 4
			}
		}
	}
}

func (c *Converter) readCSI(r io.RuneReader) (err error) {
	isEnd := func(char rune) bool {
		return char < 0x20 || char >= 0x40
	}
	var params []rune
	var num rune
	var first rune = -1
	for {
		code, _, err := r.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if first < 0 {
			first = code
		}
		if isEnd(code) {
			params = append(params, num)
			if code == xm {
				c.setAttributes(params)
			}
			break
		}
		if code == xSemiColon {
			params = append(params, num)
			num = 0
		} else if code >= '0' && code <= '9' {
			num = 10*num + (code - '0')
		} else {
			// return ErrNotSupported
			continue
		}
	}
	return
}

func (c *Converter) readOSC(r io.RuneReader, w writer) (err error) {
	var mode rune = -1
	var paramsBuilder strings.Builder
	var urlBuilder strings.Builder
	state := 0
	wrapSpan := func(cb func() error) (err error) {
		if c.isSpan {
			if _, err = c.r.spanClose(w); err != nil {
				return
			}
			if err = cb(); err != nil {
				return
			}
			if _, err = c.r.spanOpen(w, c.prevStyle); err != nil {
				return
			}
			return
		}
		return cb()
	}

	handle := func() {
		url := urlBuilder.String()
		defer func() {
			c.isAnchor = url != ""
		}()
		if url != "" {
			a := &anchor{
				url:    url,
				params: map[string]string{},
			}

			params := strings.Split(paramsBuilder.String(), ":")
			for _, str := range params {
				values := strings.Split(str, "=")
				if len(values) == 2 {
					a.params[values[0]] = values[1]
				}
			}

			if c.prevAnchor != nil {
				err = wrapSpan(func() error {
					_, err := c.r.anchorNext(w, a)
					return err
				})
				if err != nil {
					return
				}
			} else {
				err = wrapSpan(func() error {
					_, err := c.r.anchorOpen(w, a)
					return err
				})
				if err != nil {
					return
				}
			}
			c.prevAnchor = a
			return
		}
		if !c.isAnchor {
			return
		}
		err = wrapSpan(func() error {
			_, err := c.r.anchorClose(w)
			c.prevAnchor = nil
			return err
		})
	}

	for {
		code, _, err := r.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if code == xST || code == xBEL {
			if mode == 8 {
				handle()
			}
			break
		}
		if code == xESC {
			code, _, err = r.ReadRune()
			if err == io.EOF {
				return ErrUnexpected
			}
			if err != nil {
				return err
			}
			if code == xBackslash {
				if mode == 8 {
					handle()
				}
				break
			}
			return ErrUnexpected
		}
		if code == xSemiColon {
			state++
		} else if state == 0 {
			mode = code - '0'
		} else if state == 1 {
			_, _ = paramsBuilder.WriteRune(code)
		} else if state == 2 {
			_, _ = urlBuilder.WriteRune(code)
		} else {
			continue
		}
	}
	return
}

func (c *Converter) getForegroundRgb(fgColorMode colorMode, fgIndexOrRgb rune) (rune, error) {
	switch fgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		{
			if c.bold && fgIndexOrRgb < 8 {
				fgIndexOrRgb += 8
			}
			if int(fgIndexOrRgb) >= len(c.palette.colors) {
				return 0, fmt.Errorf("%w: %d", ErrColorUndefined, fgIndexOrRgb)
			}
			return c.palette.colors[fgIndexOrRgb].rgb, nil
		}
	case cmRGB:
		return fgIndexOrRgb, nil
	}
	if c.inverse {
		if c.palette.background != nil {
			return c.palette.background.rgb, nil
		}
		return 0, ErrColorUndefined
	}
	if c.palette.foreground != nil {
		return c.palette.foreground.rgb, nil
	}
	return 0, ErrColorUndefined
}

func (c *Converter) getBackgroundRgb(bgColorMode colorMode, bgIndexOrRgb rune) (rune, error) {
	switch bgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		{
			if int(bgIndexOrRgb) >= len(c.palette.colors) {
				return 0, fmt.Errorf("%w: %d", ErrColorUndefined, bgIndexOrRgb)
			}
			return c.palette.colors[bgIndexOrRgb].rgb, nil
		}
	case cmRGB:
		return bgIndexOrRgb, nil
	}
	if c.inverse {
		if c.palette.foreground != nil {
			return c.palette.foreground.rgb, nil
		}
		return 0, ErrColorUndefined
	}
	if c.palette.background != nil {
		return c.palette.background.rgb, nil
	}
	return 0, ErrColorUndefined
}

func (c *Converter) getForegroundCSS(bgColorMode colorMode, bgIndexOrRgb rune, fgColorMode colorMode, fgIndexOrRgb rune) (string, error) {
	minimumContrastCSS, ok := c.getMinimumContrastCSS(
		bgColorMode,
		bgIndexOrRgb,
		fgColorMode,
		fgIndexOrRgb,
	)
	if ok {
		return minimumContrastCSS, nil
	}
	css, err := c._getForegroundCSS(fgColorMode, fgIndexOrRgb)
	if err != nil {
		return "", err
	}
	return css, nil
}

func (c *Converter) getBackgroundCSS(bgColorMode colorMode, bgIndexOrRgb rune) (string, error) {
	switch bgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		{
			if int(bgIndexOrRgb) >= len(c.palette.colors) {
				return "", fmt.Errorf("%w: %d", ErrColorUndefined, bgIndexOrRgb)
			}
			return c.palette.colors[bgIndexOrRgb].css, nil
		}
	case cmRGB:
		return toCSS(bgIndexOrRgb), nil

	}
	if c.inverse {
		if c.palette.foreground != nil {
			return c.palette.foreground.css, nil
		}
		return "", nil
	}
	if c.palette.background != nil {
		return c.palette.background.css, nil
	}
	return "", nil
}

func (c *Converter) getMinimumContrastCSS(bgColorMode colorMode, bgIndexOrRgb rune, fgColorMode colorMode, fgIndexOrRgb rune) (string, bool) {
	if c.minimumContrastRatio <= 1 {
		return "", false
	}
	bg := bgIndexOrRgb<<8 | int32(bgColorMode)
	fg := (fgIndexOrRgb << 8) | int32(fgColorMode)
	adjustedColor, ok := c.contrastCache.Get(bg, fg)
	if ok {
		if adjustedColor != nil {
			return *adjustedColor, true
		}
		return "", false
	}

	fgRgb, fgErr := c.getForegroundRgb(fgColorMode, fgIndexOrRgb)
	bgRgb, bgErr := c.getBackgroundRgb(bgColorMode, bgIndexOrRgb)
	if fgErr != nil || bgErr != nil {
		return "", false
	}

	rgb, ok := ensureContrastRatio(fgRgb, bgRgb, c.minimumContrastRatio)
	if !ok {
		c.contrastCache.Set(bg, fg, nil)
		return "", false
	}
	css := toCSS(rgb)
	c.contrastCache.Set(bg, fg, &css)
	return css, true
}

func (c *Converter) _getForegroundCSS(fgColorMode colorMode, fgIndexOrRgb rune) (string, error) {
	switch fgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		if c.bold && fgIndexOrRgb < 8 {
			fgIndexOrRgb += 8
		}
		if int(fgIndexOrRgb) >= len(c.palette.colors) {
			return "", fmt.Errorf("%w: %d", ErrColorUndefined, fgIndexOrRgb)
		}
		return c.palette.colors[fgIndexOrRgb].css, nil

	case cmRGB:
		return toCSS(fgIndexOrRgb), nil
	}
	if c.inverse {
		if c.palette.background != nil {
			return c.palette.background.css, nil
		}
		return "", nil
	}
	if c.palette.foreground != nil {
		return c.palette.foreground.css, nil
	}
	return "", nil
}

func (c *Converter) getForegroundClass(
	fgColorMode colorMode,
	fgIndexOrRgb rune,
) string {
	switch fgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		if c.bold && c.fgIndexOrRgb < 8 {
			fgIndexOrRgb += 8
		}
		return strconv.FormatInt(int64(fgIndexOrRgb), 10)
	}
	if c.inverse {
		return "inverse"
	}
	return ""
}

func (c *Converter) getBackgroundClass(
	bgColorMode colorMode,
	bgIndexOrRgb rune,
) string {
	switch bgColorMode {
	case cmP16:
		fallthrough
	case cmP256:
		return strconv.FormatInt(int64(bgIndexOrRgb), 10)

	}
	if c.inverse {
		return "inverse"
	}
	return ""
}
