package ansihtml

type Option func(*Converter)

func SetTheme(theme Theme) Option {
	return func(c *Converter) {
		c.palette = buildPalette(theme)
	}
}

func SetMode(mode Mode) Option {
	return func(c *Converter) {
		c.isClass = mode == Class
	}
}

func SetMinimumContrastRatio(ratio float64) Option {
	return func(c *Converter) {
		c.minimumContrastRatio = ratio
	}
}

func SetClassPrefix(prefix string) Option {
	return func(c *Converter) {
		c.classPrefix = prefix
	}
}

func SetEscapeHTML(b bool) Option {
	return func(c *Converter) {
		c.escapeHTML = b
	}
}

type Options struct {
	Mode                 Mode
	ClassPrefix          string
	MinimumContrastRatio float64
	Theme                Theme
	EscapeHTML           bool
}

func SetOptions(opts Options) Option {
	return func(c *Converter) {
		c.isClass = opts.Mode == Class
		c.minimumContrastRatio = opts.MinimumContrastRatio
		if opts.MinimumContrastRatio < 1 {
			c.minimumContrastRatio = 3
		}
		c.classPrefix = opts.ClassPrefix
		c.palette = buildPalette(opts.Theme)
		c.escapeHTML = opts.EscapeHTML
	}
}
