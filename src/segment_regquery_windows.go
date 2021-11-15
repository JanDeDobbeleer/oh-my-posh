//go:build windows
package main

import (
	"fmt"
)

func (r *regquery) enabled() bool {
	// Call registry code and full out "content" string.	

	r.content = fmt.Sprintf("Reg query %s /v %s", RegistryPath ,RegistryKey);
	return true;
}
