package log

type format byte

const (
	Red format = 1 << iota
	Yellow
)

type entry string

func (e *entry) Format(formats ...format) {
	str := *e
	for _, format := range formats {
		switch format {
		case Red:
			str = "\033[31m" + str
		case Yellow:
			str = "\033[33m" + str
		}
		str += "\033[0m"
	}
	*e = str
}
