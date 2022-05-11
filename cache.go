package ansihtml

type contrastCache struct {
	rgb map[rune]map[rune]*string
}

func newContrastCache() *contrastCache {
	return &contrastCache{rgb: map[rune]map[rune]*string{}}
}

func (c *contrastCache) Clear() {
	c.rgb = map[rune]map[rune]*string{}
}

func (c *contrastCache) Set(bg rune, fg rune, value *string) {
	if m, ok := c.rgb[bg]; ok {
		m[fg] = value
	} else {
		m = map[rune]*string{}
		m[fg] = value
		c.rgb[bg] = m
	}
}

func (c *contrastCache) Get(bg rune, fg rune) (*string, bool) {
	m, ok := c.rgb[bg]
	if !ok {
		return nil, false
	}
	css, ok := m[fg]
	if !ok {
		return nil, false
	}
	return css, true
}
