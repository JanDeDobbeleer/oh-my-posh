//go:build windows
package main

import (
	"fmt"
    "golang.org/x/sys/windows"
)

func (r *regquery) enabled() bool {
    var ret bool = false;
    r.content = "fail";

	// Call registry code and full out "content" string.	
    registryPath := r.props.getString(RegistryPath, "")
    registryKey := r.props.getString(RegistryKey, "")

    regPathUTF16, regPathErr :=  windows.UTF16FromString(registryPath);

    if (regPathErr == nil) {
        var hKey windows.Handle;

        regOpenErr := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, &regPathUTF16[0], 0, windows.KEY_READ, &hKey)

        if (regOpenErr == nil) {
            regKeyUTF16, regKeyErr := windows.UTF16FromString(registryKey);

            if (regKeyErr == nil){
                // size first...
                var keyBufType uint32;
                var keyBufSize uint32;

                regQueryErr := windows.RegQueryValueEx(hKey, &regKeyUTF16[0], nil, &keyBufType, nil, &keyBufSize);

                if (regQueryErr == nil) {        
                    r.content = fmt.Sprintf("Reg query do it! %s /v %s, type: %d", registryPath ,registryKey, keyBufType);
                    ret = true;
                }
            }
        }

        windows.RegCloseKey(hKey);
    }
    return ret;
}
