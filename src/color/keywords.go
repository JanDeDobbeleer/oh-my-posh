package color

const (
	// Transparent implies a transparent color
	Transparent Ansi = "transparent"
	// Accent is the OS accent color
	Accent Ansi = "accent"
	// ParentBackground takes the previous segment's background color
	ParentBackground Ansi = "parentBackground"
	// ParentForeground takes the previous segment's color
	ParentForeground Ansi = "parentForeground"
	// Background takes the current segment's background color
	Background Ansi = "background"
	// Foreground takes the current segment's foreground color
	Foreground Ansi = "foreground"
)

func (color Ansi) isKeyword() bool {
	switch color { //nolint: exhaustive
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func (color Ansi) Resolve(current *Set, parents []*Set) Ansi {
	resolveParentColor := func(keyword Ansi) Ansi {
		for _, parentColor := range parents {
			if parentColor == nil {
				return Transparent
			}

			switch keyword { //nolint: exhaustive
			case ParentBackground:
				keyword = parentColor.Background
			case ParentForeground:
				keyword = parentColor.Foreground
			default:
				if len(keyword) == 0 {
					return Transparent
				}
				return keyword
			}
		}

		if len(keyword) == 0 {
			return Transparent
		}

		return keyword
	}

	resolveKeyword := func(keyword Ansi) Ansi {
		switch {
		case keyword == Background && current != nil:
			return current.Background
		case keyword == Foreground && current != nil:
			return current.Foreground
		case (keyword == ParentBackground || keyword == ParentForeground) && parents != nil:
			return resolveParentColor(keyword)
		default:
			return Transparent
		}
	}

	for color.isKeyword() {
		resolved := resolveKeyword(color)
		if resolved == color {
			break
		}

		color = resolved
	}

	return color
}
