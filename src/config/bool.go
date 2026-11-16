package config

// A plain *bool loses an explicit false during gob encoding of the config cache.
type Bool bool

func (b Bool) GobEncode() ([]byte, error) {
	if b {
		return []byte{1}, nil
	}

	return []byte{0}, nil
}

func (b *Bool) GobDecode(data []byte) error {
	*b = len(data) > 0 && data[0] != 0
	return nil
}
