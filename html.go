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

type renderer struct {
	classPrefix string
	defaultFg   *colorObject
	isClass     bool
}

func (r *renderer) spanOpen(w writer, s *spanStyle) (size int64, err error) {
	buf := &bytes.Buffer{}
	var classes []string
	props := map[string]string{}
	_, _ = buf.WriteString(`<span`)

	if r.isClass {
		if s.foreground != "" {
			if s.fgMode != cmRGB {
				classes = append(classes, r.classPrefix+"fg-"+s.foreground)
			} else {
				props["color"] = s.foreground
			}
		}
		if s.background != "" {
			if s.bgMode != cmRGB {
				classes = append(classes, r.classPrefix+"bg-"+s.background)
			} else {
				props["color"] = s.background
			}
		}
		if s.bold {
			classes = append(classes, r.classPrefix+"bold")
		}
		if s.underline {
			classes = append(classes, r.classPrefix+"underline")
		}
		if s.strike {
			classes = append(classes, r.classPrefix+"strike")
		}
		if s.italic {
			classes = append(classes, r.classPrefix+"italic")
		}
		if s.dim {
			classes = append(classes, r.classPrefix+"dim")
		}
		if s.hidden {
			classes = append(classes, r.classPrefix+"hidden")
		}
	} else {
		if s.foreground != "" {
			props["color"] = s.foreground
		}
		if s.background != "" {
			props["background-color"] = s.background
		}
		if s.bold {
			props["font-weight"] = "700"
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
			} else if r.defaultFg != nil && r.defaultFg.css != "" {
				props["color"] = r.defaultFg.css + "80"
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

func (r *renderer) spanClose(w writer) (size int, err error) {
	return w.WriteString("</span>")
}

func (r *renderer) rune(w writer, c rune) (size int, err error) {
	return w.WriteRune(c)
}

func (r *renderer) anchorOpen(w writer, a *anchor) (size int64, err error) {
	buf := &bytes.Buffer{}
	var keys []string
	for k := range a.params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	_, _ = buf.WriteString(`<a `)
	_, _ = buf.WriteString(fmt.Sprintf("href=\"%s\"", a.url))
	_, _ = buf.WriteString(" class=\"ansi-link\"")

	for i := 0; i < len(keys); i++ {
		_, _ = buf.WriteString(fmt.Sprintf(" %s=\"%s\"", keys[i], a.params[keys[i]]))
	}
	_, _ = buf.WriteRune('>')
	return io.Copy(w, buf)
}

func (r *renderer) anchorNext(w writer, a *anchor) (size int64, err error) {
	if n, err := r.anchorClose(w); err != nil {
		return int64(n), err
	}
	return r.anchorOpen(w, a)
}

func (r *renderer) anchorClose(w writer) (size int, err error) {
	return w.WriteString("</a>")
}
