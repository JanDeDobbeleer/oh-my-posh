//go:build windows
package main

import (
	"fmt"
    "golang.org/x/sys/windows"
    "unsafe"
)

func (r *regquery) enabled() bool {
    var ret bool = false
    r.content = "fail"

	// Call registry code and fill out "content" string.	
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
                var keyBufType uint32
                var keyBufSize uint32

                regQueryErr := windows.RegQueryValueEx(hKey, &regKeyUTF16[0], nil, &keyBufType, nil, &keyBufSize);

                if (regQueryErr == nil) {
                    
                    var keyBuf = make([]byte, keyBufSize, keyBufSize);

                    regQueryErr := windows.RegQueryValueEx(hKey, &regKeyUTF16[0], nil, &keyBufType, &keyBuf[0], &keyBufSize);

                    if (regQueryErr == nil) {

                        switch (keyBufType) {
                            case windows.REG_SZ:
                                var uint16p *uint16
                                uint16p = (*uint16)(unsafe.Pointer(&keyBuf[0]))  // nasty casty
                                s := windows.UTF16PtrToString(uint16p)
                                r.content = s
                                ret = true
                            case windows.REG_DWORD:
                                var uint32p *uint32
                                uint32p = (*uint32)(unsafe.Pointer(&keyBuf[0])) // more casting goodness
                                r.content = fmt.Sprintf("0x%08X", *uint32p)
                                ret = true
                            default:
                                r.content = fmt.Sprintf("default: %d, %d %d", keyBufType, keyBufSize, keyBuf[0])
                            }
                    }
                }
            }
        }

        windows.RegCloseKey(hKey);
    }
    return ret;
}
