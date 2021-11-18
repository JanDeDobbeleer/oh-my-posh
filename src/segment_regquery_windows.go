//go:build windows
package main

import (
	"fmt"
    "golang.org/x/sys/windows"
    "unsafe"
)

// TODO
//  Code tidyup
//      Move registry code out?
//      Make nested if/else more easy to understand
//  Test cases
//      Include registry calls?
//  Docs

func (r *regquery) getHKEYValueFromAbbrString(abbr string) windows.Handle {
    var ret windows.Handle = 0

    switch (abbr) {
        case "HKCR":
            ret = windows.HKEY_CLASSES_ROOT
        case "HKCC":
            ret = windows.HKEY_CURRENT_CONFIG
        case "HKCU":
            ret = windows.HKEY_CURRENT_USER
        case "HKLM":
            ret = windows.HKEY_LOCAL_MACHINE
        case "HKU":
            ret = windows.HKEY_USERS
    }

    return ret
}

func (r *regquery) enabled() bool {

    var enableSegment bool = false
    var failed bool = false

	// Call registry code and fill out "content" string.
    registryRoot := r.props.getString(RegistryRoot, "")
    registryRootHKEYValue := r.getHKEYValueFromAbbrString(registryRoot)
    
    registryPath := r.props.getString(RegistryPath, "")
    registryKey := r.props.getString(RegistryKey, "")

    // Fallback behaviour
    failBehaviour := r.props.getString(QueryFailBehaviour, "hide_segment")
    fallbackString := r.props.getString(QueryFailFallbackString, "")

    regPathUTF16, regPathErr :=  windows.UTF16FromString(registryPath);

    if (regPathErr == nil) {
        var hKey windows.Handle;
                                  
        regOpenErr := windows.RegOpenKeyEx(registryRootHKEYValue, &regPathUTF16[0], 0, windows.KEY_READ, &hKey)

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
                                enableSegment = true
                            case windows.REG_DWORD:
                                var uint32p *uint32
                                uint32p = (*uint32)(unsafe.Pointer(&keyBuf[0])) // more casting goodness
                                
                                // add a property for the sprintf format string so it can be defined by the user?
                                r.content = fmt.Sprintf("0x%08X", *uint32p)
                                enableSegment = true
                            default:
                                r.errorInfo = fmt.Sprintf("no formatter for type:%d, data size:%d bytes", keyBufType, keyBufSize)
                                failed = true
                            }
                    } else {
                        // key value query failure
                        r.errorInfo = fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data with error '%s'", regQueryErr)
                        failed = true
                    }
                } else {
                    r.errorInfo = fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data size with error '%s'", regQueryErr)
                    failed = true
                }
            } else {
                r.errorInfo = fmt.Sprintf("Error converting regsitry_key (%s) to UTF16, with error %d", registryKey, regKeyErr)
                failed = true        
            }
            
            windows.RegCloseKey(hKey);

        } else {
            r.errorInfo = fmt.Sprintf("Error calling RegOpenKeyEx to open registry (root handle: '%s', path: '%s') with error '%s'", registryRoot, registryPath, regOpenErr)
            failed = true
        }
    } else {
        r.errorInfo = fmt.Sprintf("Error converting regsitry_path (%s) to UTF16, with error %d", registryPath, regPathErr)
        failed = true
    }

    // Ok decide what to do...
    if (failed) {
        switch (failBehaviour){
            case "hide_segment":
                enableSegment = false
            case "display_fallback_string":
                r.content = fallbackString
                enableSegment = true
            case "show_debug_info":
                r.content = r.errorInfo
                enableSegment = true
        }
    } else {
        enableSegment = true
    }

    return enableSegment;
}
