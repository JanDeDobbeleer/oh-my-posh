//go:build !windows

package upgrade

func hideFile(_ string) error {
	return nil
}

func updateRegistry(_, _ string) {}
