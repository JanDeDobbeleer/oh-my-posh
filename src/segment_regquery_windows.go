//go:build windows
package main

import (
	"fmt"
    "strings"
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
        case "HKCR", "HKEY_CLASSES_ROOT":
            ret = windows.HKEY_CLASSES_ROOT
        case "HKCC", "HKEY_CURRENT_CONFIG":
            ret = windows.HKEY_CURRENT_CONFIG
        case "HKCU", "HKEY_CURRENT_USER":
            ret = windows.HKEY_CURRENT_USER
        case "HKLM", "HKEY_LOCAL_MACHINE":
            ret = windows.HKEY_LOCAL_MACHINE
        case "HKU", "HKEY_USERS":
            ret = windows.HKEY_USERS
    }

    return ret
}


//
// Takes a registry path like "HKLM\Software\Microsoft\Windows NT\CurrentVersion" and a key under that path like "CurrentVersion" (or "" if the (Default) key is required).
// Returns a bool and string:
//
//   true and the retrieved value formatted into a string if successful.
//   false and the string will be the error
//
func (r *regquery) getRegistryKeyValue(regPath string, regKey string) (bool, string) {

    // using short-circuit logic as previous nested if version was horrible to understand.

    // Extract root HK value and turn it into a windows.Handle to open the key.
    regPathParts := strings.SplitN(regPath, "\\", 2)

    regRootHKeyHandle := r.getHKEYValueFromAbbrString(regPathParts[0])
    if (regRootHKeyHandle == 0) {
        return false, fmt.Sprintf("Supplied root HKEY value not valid: '%s'", regPathParts[0])
    }

    // Second part of split is registry path after HK part - needs to be UTF16 to pass to the windows. API
    regPathUTF16, regPathUTF16ConversionErr :=  windows.UTF16FromString(regPathParts[1])
    if (regPathUTF16ConversionErr != nil) {
        return false, fmt.Sprintf("Could not convert supplied path '%s' to UTF16, error: '%s'", regPathParts[1], regPathUTF16ConversionErr)
    }

    // Ok - open it..
    var hKeyHandle windows.Handle;
    regOpenErr := windows.RegOpenKeyEx(regRootHKeyHandle, &regPathUTF16[0], 0, windows.KEY_READ, &hKeyHandle)
    if (regOpenErr != nil) {
        return false, fmt.Sprintf("Error RegOpenKeyEx opening registry path to '%s', error: '%s'", regPath, regOpenErr)
    }
    // Success - from here on out, when returning make sure to close that reg key with a deferred call to close:
    defer windows.RegCloseKey(hKeyHandle)

    // Again - need UTF16 of the key for the API:
    regKeyUTF16, regKeyUTF16ConversionErr := windows.UTF16FromString(regKey)
    if (regKeyUTF16ConversionErr != nil) {
        return false, fmt.Sprintf("Could not convert supplied key '%s' to UTF16, error: '%s'", regKey, regKeyUTF16ConversionErr)
    }

    // Two stage way to get the key value - query once to get size - then allocate and query again to fill it. 
    var keyBufType uint32
    var keyBufSize uint32

    regQueryErr := windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, nil, &keyBufSize);
    if (regQueryErr != nil) {
        return false, fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data size with error '%s'", regQueryErr)
    }

    // Alloc and fill...
    var keyBuf = make([]byte, keyBufSize, keyBufSize);

    regQueryErr = windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, &keyBuf[0], &keyBufSize);
    if (regQueryErr != nil) {
        return false, fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data with error '%s'", regQueryErr)
    }

    // Format result into a string, depending on type.  (future refactor - move this out into it's own function)
    switch (keyBufType) {
        case windows.REG_SZ:
            var uint16p *uint16
            uint16p = (*uint16)(unsafe.Pointer(&keyBuf[0]))  // nasty casty
            
            return true, windows.UTF16PtrToString(uint16p)
        case windows.REG_DWORD:
            var uint32p *uint32
            uint32p = (*uint32)(unsafe.Pointer(&keyBuf[0])) // more casting goodness
            
            return true, fmt.Sprintf("0x%08X", *uint32p)
        default:
            return false, fmt.Sprintf("Error, no formatter for REG_? type:%d, data size:%d bytes", keyBufType, keyBufSize)
    }    
}

func (r *regquery) enabled() bool {

    var enableSegment bool = false

    registryPath := r.props.getString(RegistryPath, "")
    registryKey := r.props.getString(RegistryKey, "")

    b, s := r.getRegistryKeyValue(registryPath, registryKey)

    // Fallback behaviour
    failBehaviour := r.props.getString(QueryFailBehaviour, "hide_segment")
    fallbackString := r.props.getString(QueryFailFallbackString, "")

    if (!b) {
        switch (failBehaviour){
            case "hide_segment":
                enableSegment = false
            case "display_fallback_string":
                r.content = fallbackString
                enableSegment = true
            case "show_debug_info":
                r.content = s//r.errorInfo
                enableSegment = true
        }
    } else {
        r.content = s
        enableSegment = true
    }

    return enableSegment;
}
