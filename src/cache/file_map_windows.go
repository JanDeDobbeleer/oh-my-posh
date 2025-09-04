package cache

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Configuration constants
const (
	maxStringSize = 1024 * 1024 // 1MB maximum string size
)

// Windows API constants
const (
	fileMapAllAccess    = 0x001f001f
	pageReadwrite       = 0x04
	genericRead         = 0x80000000
	genericWrite        = 0x40000000
	createAlways        = 2
	openExisting        = 3
	fileAttributeNormal = 0x80
)

// Windows API functions
var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	createFileW        = kernel32.NewProc("CreateFileW")
	createFileMappingW = kernel32.NewProc("CreateFileMappingW")
	mapViewOfFile      = kernel32.NewProc("MapViewOfFile")
	unmapViewOfFile    = kernel32.NewProc("UnmapViewOfFile")
	closeHandle        = kernel32.NewProc("CloseHandle")
	setFilePointer     = kernel32.NewProc("SetFilePointer")
	setEndOfFile       = kernel32.NewProc("SetEndOfFile")
	getFileSizeEx      = kernel32.NewProc("GetFileSizeEx")
)

// PersistentSharedString represents a memory-mapped file for storing a single string
type PersistentSharedString struct {
	filePath   string
	fileHandle uintptr
	mapHandle  uintptr
	data       uintptr
}

func createOrOpenPersistentString(filePath string) (*PersistentSharedString, error) {
	// First, try to open existing file
	pss, err := openExistingFile(filePath)
	if err == nil {
		return pss, nil
	}

	// File doesn't exist, create new one
	return createNewFile(filePath)
}

// openExistingFile attempts to open an existing memory-mapped file
func openExistingFile(filePath string) (*PersistentSharedString, error) {
	filePathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert file path to UTF16: %v", err)
	}

	// Try to open existing file
	fileHandle, _, _ := createFileW.Call(
		uintptr(unsafe.Pointer(filePathPtr)), // lpFileName
		genericRead|genericWrite,             // dwDesiredAccess
		0,                                    // dwShareMode
		0,                                    // lpSecurityAttributes
		openExisting,                         // dwCreationDisposition
		fileAttributeNormal,                  // dwFlagsAndAttributes
		0,                                    // hTemplateFile
	)

	if fileHandle == uintptr(0xFFFFFFFFFFFFFFFF) { // INVALID_HANDLE_VALUE
		return nil, fmt.Errorf("file does not exist")
	}

	// Get file size to validate it's large enough
	var fileSize int64
	ret, _, _ := getFileSizeEx.Call(fileHandle, uintptr(unsafe.Pointer(&fileSize)))
	if ret == 0 || uintptr(fileSize) < maxStringSize+5 { // 4 bytes length + 1 null terminator
		_, _, _ = closeHandle.Call(fileHandle)
		return nil, fmt.Errorf("existing file is too small")
	}

	return createMappingFromFile(filePath, fileHandle)
}

// createNewFile creates a new memory-mapped file
func createNewFile(filePath string) (*PersistentSharedString, error) {
	filePathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert file path to UTF16: %v", err)
	}

	// Create new file
	fileHandle, _, err := createFileW.Call(
		uintptr(unsafe.Pointer(filePathPtr)), // lpFileName
		genericRead|genericWrite,             // dwDesiredAccess
		0,                                    // dwShareMode
		0,                                    // lpSecurityAttributes
		createAlways,                         // dwCreationDisposition (overwrites if exists)
		fileAttributeNormal,                  // dwFlagsAndAttributes
		0,                                    // hTemplateFile
	)

	if fileHandle == uintptr(0xFFFFFFFFFFFFFFFF) { // INVALID_HANDLE_VALUE
		return nil, fmt.Errorf("CreateFileW failed: %v", err)
	}

	// Set file size (4 bytes for length + MAX_STRING_SIZE for string + 1 for null terminator)
	totalSize := maxStringSize + 5
	_, _, _ = setFilePointer.Call(fileHandle, uintptr(totalSize), 0, 0) // FILE_BEGIN = 0
	_, _, _ = setEndOfFile.Call(fileHandle)

	pss, mapErr := createMappingFromFile(filePath, fileHandle)
	if mapErr != nil {
		_, _, _ = closeHandle.Call(fileHandle)
		return nil, mapErr
	}

	// Initialize new file with empty string
	basePtr := unsafe.Pointer(pss.data)
	lengthPtr := (*uint32)(basePtr)
	*lengthPtr = 0

	return pss, nil
}

// createMappingFromFile creates a memory mapping from an open file handle
func createMappingFromFile(filePath string, fileHandle uintptr) (*PersistentSharedString, error) {
	totalSize := maxStringSize + 5 // 4 bytes length + MAX_STRING_SIZE + 1 null terminator

	// Create file mapping
	mapHandle, _, err := createFileMappingW.Call(
		fileHandle,         // hFile
		0,                  // lpAttributes (NULL)
		pageReadwrite,      // flProtect
		0,                  // dwMaximumSizeHigh
		uintptr(totalSize), // dwMaximumSizeLow
		0,                  // lpName (NULL for unnamed mapping)
	)

	if mapHandle == 0 {
		return nil, fmt.Errorf("CreateFileMappingW failed: %v", err)
	}

	// Map view of file
	data, _, err := mapViewOfFile.Call(
		mapHandle,          // hFileMappingObject
		fileMapAllAccess,   // dwDesiredAccess
		0,                  // dwFileOffsetHigh
		0,                  // dwFileOffsetLow
		uintptr(totalSize), // dwNumberOfBytesToMap
	)

	if data == 0 {
		_, _, _ = closeHandle.Call(mapHandle)
		return nil, fmt.Errorf("MapViewOfFile failed: %v", err)
	}

	return &PersistentSharedString{
		filePath:   filePath,
		fileHandle: fileHandle,
		mapHandle:  mapHandle,
		data:       data,
	}, nil
}

// SetString stores a string in the memory-mapped file (automatically persisted)
func (pss *PersistentSharedString) SetString(value string) error {
	strBytes := []byte(value)

	if uintptr(len(strBytes)) > maxStringSize {
		return fmt.Errorf("string too large for allocated space (%d > %d)", len(strBytes), maxStringSize)
	}

	basePtr := unsafe.Pointer(pss.data)

	// Write length as first 4 bytes (little-endian)
	lengthPtr := (*uint32)(basePtr)
	*lengthPtr = uint32(len(strBytes))

	// Write string data starting at offset 4
	if len(strBytes) > 0 {
		stringPtr := unsafe.Add(basePtr, 4)
		stringSlice := unsafe.Slice((*byte)(stringPtr), len(strBytes))
		copy(stringSlice, strBytes)
	}

	// Write null terminator
	nullPtr := (*byte)(unsafe.Add(basePtr, 4+len(strBytes)))
	*nullPtr = 0

	// No need to explicitly flush - Windows handles this automatically
	return nil
}

func (pss *PersistentSharedString) bytes() []byte {
	basePtr := unsafe.Pointer(pss.data)

	// Read length from first 4 bytes
	lengthPtr := (*uint32)(basePtr)
	length := *lengthPtr

	if length == 0 {
		log.Debug("empty string")
		return []byte{0}
	}

	if length > uint32(maxStringSize) {
		log.Error(fmt.Errorf("corrupted data: length %d exceeds max size %d", length, maxStringSize))
		return []byte{0}
	}

	// Read string data starting at offset 4
	stringPtr := unsafe.Add(basePtr, 4)
	stringSlice := unsafe.Slice((*byte)(stringPtr), length)

	// Convert to string
	result := make([]byte, length)
	copy(result, stringSlice)
	return result
}

// Close closes the memory-mapped file and handles
func (pss *PersistentSharedString) close() error {
	var err error

	if pss.data != 0 {
		if ret, _, e := unmapViewOfFile.Call(pss.data); ret == 0 {
			err = fmt.Errorf("UnmapViewOfFile failed: %v", e)
		}
		pss.data = 0
	}

	if pss.mapHandle != 0 {
		if ret, _, e := closeHandle.Call(pss.mapHandle); ret == 0 {
			if err == nil {
				err = fmt.Errorf("CloseHandle (mapping) failed: %v", e)
			}
		}
		pss.mapHandle = 0
	}

	if pss.fileHandle != 0 {
		if ret, _, e := closeHandle.Call(pss.fileHandle); ret == 0 {
			if err == nil {
				err = fmt.Errorf("CloseHandle (file) failed: %v", e)
			}
		}
		pss.fileHandle = 0
	}

	return err
}
