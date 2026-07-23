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
	switch color {
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func (color Ansi) Resolve(current *Set, parents []*Set) Ansi {
	resolveParentColor := func(keyword Ansi) Ansi {
		// parents is a stack pushed tail-first (see terminal.SetParentColors):
		// the nearest ancestor is the last element, so walk back-to-front.
		for i := len(parents) - 1; i >= 0; i-- {
			parentColor := parents[i]
			if parentColor == nil {
				return Transparent
			}

			switch keyword {
			case ParentBackground:
				keyword = parentColor.Background
			case ParentForeground:
				keyword = parentColor.Foreground
			default:
				if keyword == "" {
					return Transparent
				}
				return keyword.GradientLast()
			}

			if !keyword.IsGradient() {
				continue
			}

			// a parent gradient collapses to its last stop; a keyword stop refers to
			// that SAME parent's colors, never to the child segment asking for the
			// parent color. A parentBackground/parentForeground stop walks further up
			// through the next iteration; an unresolvable self-reference degrades to
			// transparent instead of leaking a keyword the child would misresolve.
			stop := keyword.GradientLast()

			switch stop { //nolint:exhaustive
			case Foreground:
				stop = parentColor.Foreground.GradientLast()
			case Background:
				stop = parentColor.Background.GradientLast()
			}

			if stop.isKeyword() && stop != ParentBackground && stop != ParentForeground && stop != Transparent {
				return Transparent
			}

			keyword = stop
		}

		if keyword == "" {
			return Transparent
		}

		return keyword.GradientLast()
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
