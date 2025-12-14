package cache

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Configuration constants
const (
	minStringSize = 50 * 1024  // 50KB minimum string size
	maxStringSize = 10 * 1024 * 1024 // 10MB maximum string size
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
	size       int // Current allocated size
}

func createOrOpenPersistentString(filePath string) (*PersistentSharedString, error) {
	return createOrOpenPersistentStringWithSize(filePath, minStringSize)
}

func createOrOpenPersistentStringWithSize(filePath string, requiredSize int) (*PersistentSharedString, error) {
	// Ensure size is within bounds
	if requiredSize < minStringSize {
		requiredSize = minStringSize
	}
	if requiredSize > maxStringSize {
		return nil, fmt.Errorf("required size %d exceeds maximum %d", requiredSize, maxStringSize)
	}

	// First, try to open existing file
	pss, err := openExistingFileWithSize(filePath, requiredSize)
	if err == nil {
		return pss, nil
	}

	// File doesn't exist or too small, create new one with required size
	return createNewFileWithSize(filePath, requiredSize)
}

// openExistingFileWithSize attempts to open an existing memory-mapped file
// openExistingFileWithSize attempts to open an existing memory-mapped file
func openExistingFileWithSize(filePath string, requiredSize int) (*PersistentSharedString, error) {
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

	// Get file size to check if it's large enough
	var fileSize int64
	ret, _, _ := getFileSizeEx.Call(fileHandle, uintptr(unsafe.Pointer(&fileSize)))
	if ret == 0 {
		_, _, _ = closeHandle.Call(fileHandle)
		return nil, fmt.Errorf("failed to get file size")
	}

	actualSize := int(fileSize) - 5 // Subtract header (4 bytes length + 1 null terminator)
	if actualSize < requiredSize {
		// Existing file is too small, close and recreate
		_, _, _ = closeHandle.Call(fileHandle)
		return nil, fmt.Errorf("existing file is too small (%d < %d)", actualSize, requiredSize)
	}

	return createMappingFromFileWithSize(filePath, fileHandle, actualSize)
}

// createNewFileWithSize creates a new memory-mapped file with the specified size
func createNewFileWithSize(filePath string, size int) (*PersistentSharedString, error) {
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

	// Set file size (4 bytes for length + size for string + 1 for null terminator)
	totalSize := size + 5
	_, _, _ = setFilePointer.Call(fileHandle, uintptr(totalSize), 0, 0) // FILE_BEGIN = 0
	_, _, _ = setEndOfFile.Call(fileHandle)

	pss, mapErr := createMappingFromFileWithSize(filePath, fileHandle, size)
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

// Deprecated: Use createNewFileWithSize instead
func createNewFile(filePath string) (*PersistentSharedString, error) {
	return createNewFileWithSize(filePath, minStringSize)
}

// createMappingFromFileWithSize creates a memory mapping from an open file handle with specified size
func createMappingFromFileWithSize(filePath string, fileHandle uintptr, size int) (*PersistentSharedString, error) {
	totalSize := size + 5 // 4 bytes length + size + 1 null terminator

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
		size:       size,
	}, nil
}

// Deprecated: Use createMappingFromFileWithSize instead
func createMappingFromFile(filePath string, fileHandle uintptr) (*PersistentSharedString, error) {
	return createMappingFromFileWithSize(filePath, fileHandle, minStringSize)
}
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

	if len(strBytes) > pss.size {
		return fmt.Errorf("string too large for allocated space (%d > %d)", len(strBytes), pss.size)
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

	if length > uint32(pss.size) {
		log.Error(fmt.Errorf("corrupted data: length %d exceeds allocated size %d", length, pss.size))
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
