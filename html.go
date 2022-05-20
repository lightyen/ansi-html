package ansihtml

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
)

type writer interface {
	Write(p []byte) (n int, err error)
	WriteRune(r rune) (size int, err error)
	WriteString(s string) (size int, err error)
	Flush() error
}

type render interface {
	spanOpen(w writer, s *spanStyle) (size int64, err error)
	spanClose(w writer) (size int, err error)
	rune(w writer, c rune) (size int, err error)
	anchorOpen(w writer, a *anchor) (size int64, err error)
	anchorNext(w writer, a *anchor) (size int64, err error)
	anchorClose(w writer) (size int, err error)
}

func (c *Converter) spanOpen(w writer, s *spanStyle) (size int64, err error) {
	buf := &bytes.Buffer{}
	var classes []string
	props := map[string]string{}
	_, _ = buf.WriteString(`<span`)

	if c.isClass {
		if s.foreground != "" {
			if s.fgMode != cmRGB {
				classes = append(classes, c.classPrefix+"fg-"+s.foreground)
			} else {
				props["color"] = s.foreground
			}
		}
		if s.background != "" {
			if s.bgMode != cmRGB {
				classes = append(classes, c.classPrefix+"bg-"+s.background)
			} else {
				props["background-color"] = s.background
			}
		}
		if s.bold {
			classes = append(classes, c.classPrefix+"bold")
		}
		if s.underline {
			classes = append(classes, c.classPrefix+"underline")
		}
		if s.strike {
			classes = append(classes, c.classPrefix+"strike")
		}
		if s.italic {
			classes = append(classes, c.classPrefix+"italic")
		}
		if s.dim {
			classes = append(classes, c.classPrefix+"dim")
		}
		if s.blink {
			classes = append(classes, c.classPrefix+"blink")
		}
		if s.hidden {
			classes = append(classes, c.classPrefix+"hidden")
		}
	} else {
		if s.foreground != "" {
			props["color"] = s.foreground
		}
		if s.background != "" {
			props["background-color"] = s.background
		}
		if s.bold {
			props["font-weight"] = "bold"
		}
		if s.underline || s.strike {
			var values []string
			if s.underline {
				values = append(values, "underline")
			}
			if s.strike {
				values = append(values, "line-through")
			}
			props["text-decoration"] = strings.Join(values, " ")
		}
		if s.italic {
			props["font-style"] = "italic"
		}
		if s.hidden {
			props["opacity"] = "0"
		} else if s.dim {
			if s.background == "" {
				props["opacity"] = "0.5"
			} else if s.foreground != "" {
				props["color"] = s.foreground + "80"
			} else if c.palette.foreground != nil && c.palette.foreground.css != "" {
				props["color"] = c.palette.foreground.css + "80"
			}
		}
	}

	if len(classes) > 0 {
		_, _ = buf.WriteString(` class="`)
		for i := 0; i < len(classes)-1; i++ {
			_, _ = buf.WriteString(classes[i])
			_, _ = buf.WriteRune(' ')
		}
		if len(classes) > 0 {
			_, _ = buf.WriteString(classes[len(classes)-1])
		}
		_, _ = buf.WriteRune('"')
	}

	var keys []string
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(props) > 0 {
		_, _ = buf.WriteString(` style="`)
		for i := 0; i < len(keys)-1; i++ {
			k := keys[i]
			_, _ = buf.WriteString(k)
			_, _ = buf.WriteRune(':')
			_, _ = buf.WriteString(props[k])
			_, _ = buf.WriteRune(';')
		}
		if len(keys) > 0 {
			k := keys[len(keys)-1]
			_, _ = buf.WriteString(k)
			_, _ = buf.WriteRune(':')
			_, _ = buf.WriteString(props[k])
		}
		_, _ = buf.WriteRune('"')
	}
	_, _ = buf.WriteRune('>')
	return io.Copy(w, buf)
}

func (c *Converter) spanClose(w writer) (size int, err error) {
	return w.WriteString("</span>")
}

func (c *Converter) rune(w writer, char rune) (size int, err error) {
	if c.escapeHTML {
		switch char {
		case '<':
			return w.WriteString("&lt;")
		case '>':
			return w.WriteString("&gt;")
		case '&':
			return w.WriteString("&amp;")
		case '"':
			return w.WriteString("&quot;")
		case xSingleQuote:
			return w.WriteString("&apos;")
		}
	}
	return w.WriteRune(char)
}

func (c *Converter) anchorOpen(w writer, a *anchor) (size int64, err error) {
	buf := &bytes.Buffer{}
	var keys []string
	for k := range a.params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	_, _ = buf.WriteString(`<a `)
	_, _ = buf.WriteString(fmt.Sprintf("href=\"%s\"", a.url))
	_, _ = buf.WriteString(" class=\"")
	_, _ = buf.WriteString(c.classPrefix)
	_, _ = buf.WriteString("link\"")

	for i := 0; i < len(keys); i++ {
		_, _ = buf.WriteString(fmt.Sprintf(" %s=\"%s\"", keys[i], a.params[keys[i]]))
	}
	_, _ = buf.WriteRune('>')
	return io.Copy(w, buf)
}

func (c *Converter) anchorNext(w writer, a *anchor) (size int64, err error) {
	if n, err := c.anchorClose(w); err != nil {
		return int64(n), err
	}
	return c.anchorOpen(w, a)
}

func (c *Converter) anchorClose(w writer) (size int, err error) {
	return w.WriteString("</a>")
}
