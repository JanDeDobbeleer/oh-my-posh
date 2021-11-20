package main

// CachedColors is the AnsiColors Decorator that does simple color lookup caching.
// AnsiColorFromSrting calls are cheap, but not free, and having a simple cache in
// has measurable positive effect on performance.
type CachedColors struct {
	ansiColors AnsiColors
	colorCache map[cachedColorKey]AnsiColor
}

type cachedColorKey struct {
	colorString  string
	isBackground bool
}

func (c *CachedColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	if c.colorCache == nil {
		c.colorCache = make(map[cachedColorKey]AnsiColor)
	}

	key := cachedColorKey{colorString, isBackground}
	if ansiColor, ok := c.colorCache[key]; ok {
		return ansiColor
	}

	ansiColor := c.ansiColors.AnsiColorFromString(colorString, isBackground)
	c.colorCache[key] = ansiColor
	return ansiColor
}
