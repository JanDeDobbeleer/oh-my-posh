package template

import "fmt"

func hresult(number int) string {
	return fmt.Sprintf("0x%04X", uint32(number))
}
