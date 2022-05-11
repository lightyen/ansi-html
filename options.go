package ansihtml

type Option func(*Converter)

func SetTheme(theme Theme) Option {
	return func(c *Converter) {
		c.palette = buildPalette(theme)
	}
}

func SetMode(mode Mode) Option {
	return func(c *Converter) {
		c.mode = mode
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

type Options struct {
	Mode                 Mode
	ClassPrefix          string
	MinimumContrastRatio float64
	Theme                Theme
}

func SetOptions(opts Options) Option {
	return func(c *Converter) {
		c.mode = opts.Mode
		c.minimumContrastRatio = opts.MinimumContrastRatio
		if opts.MinimumContrastRatio < 1 {
			c.minimumContrastRatio = 3
		}
		c.classPrefix = opts.ClassPrefix
		c.palette = buildPalette(opts.Theme)
	}
}
