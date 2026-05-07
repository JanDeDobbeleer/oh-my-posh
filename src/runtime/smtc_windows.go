//go:build windows

package runtime

// smtc_windows.go provides a minimal, custom WinRT binding for querying the
// Windows System Media Transport Controls (SMTC).  No third-party WinRT
// library is used; method dispatch is performed directly via COM vtable
// pointers and the combase.dll WinRT activation APIs.
//
// # COM vtable layout
//
// Every COM/WinRT interface object starts with a single pointer to its vtable.
// Indices 0-2 are IUnknown (QueryInterface, AddRef, Release).
// Indices 3-5 are IInspectable (GetIids, GetRuntimeClassName, GetTrustLevel).
// Interface-specific methods begin at index 6.
//
//	IGlobalSystemMediaTransportControlsSessionManagerStatics
//	  [6] RequestAsync → IAsyncOperation<SessionManager>
//
//	IGlobalSystemMediaTransportControlsSessionManager
//	  [6] GetCurrentSession
//	  [7] GetSessions → IVectorView<Session>
//
//	IVectorView<GlobalSystemMediaTransportControlsSession>
//	  [6] GetAt(index uint32) → Session
//	  [7] get_Size → uint32
//
//	IGlobalSystemMediaTransportControlsSession
//	  [6] get_SourceAppUserModelId → HSTRING
//	  [7] TryGetMediaPropertiesAsync → IAsyncOperation<MediaProperties>
//	  [8] GetTimelineProperties (unused)
//	  [9] GetPlaybackInfo → PlaybackInfo
//
//	IGlobalSystemMediaTransportControlsSessionMediaProperties
//	  [6] get_Title → HSTRING
//	  [7] get_AlbumTitle (unused)
//	  [8] get_Artist → HSTRING
//
//	IGlobalSystemMediaTransportControlsSessionPlaybackInfo
//	  [6]  get_Controls (unused)
//	  [7]  get_IsShuffleActive (unused)
//	  [8]  get_AutoRepeatMode (unused)
//	  [9]  get_PlaybackRate (unused)
//	  [10] get_PlaybackStatus → int32 (enum)
//
//	IAsyncOperation<T>  (after IInspectable)
//	  [6] put_Completed
//	  [7] get_Completed
//	  [8] GetResults → T
//
//	IAsyncInfo  (after IInspectable)
//	  [6] get_Id
//	  [7] get_Status → uint32
//	  [8] get_ErrorCode
//	  [9] Cancel
//	  [10] Close

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// comVTable represents the memory layout of a COM vtable: an array of
// function pointers.  64 slots is more than sufficient for all interfaces
// used here.
type comVTable struct {
	methods [64]uintptr
}

// comObj represents the memory layout of any COM interface object: a pointer
// to its vtable followed by implementation-specific data (opaque to callers).
type comObj struct {
	vtable *comVTable
}

// smtcPlaybackStatus mirrors GlobalSystemMediaTransportControlsSessionPlaybackStatus.
const (
	smtcStatusPlaying = 4
	smtcStatusPaused  = 5
)

// asyncStatus mirrors Windows.Foundation.AsyncStatus.
const (
	asyncCompleted = 1
	asyncCanceled  = 2
	asyncError     = 3
)

// asyncTimeout is the maximum time to wait for a WinRT async operation
// to reach a terminal state before giving up.
const asyncTimeout = 5 * time.Second

var (
	combase = syscall.NewLazyDLL("combase.dll")

	procRoInitialize              = combase.NewProc("RoInitialize")
	procRoUninitialize            = combase.NewProc("RoUninitialize")
	procRoGetActivationFactory    = combase.NewProc("RoGetActivationFactory")
	procWindowsCreateString       = combase.NewProc("WindowsCreateString")
	procWindowsDeleteString       = combase.NewProc("WindowsDeleteString")
	procWindowsGetStringRawBuffer = combase.NewProc("WindowsGetStringRawBuffer")

	// procRtlMoveMemory is used to copy HSTRING buffer contents into a
	// Go-owned slice without requiring uintptr→unsafe.Pointer conversion.
	// kernel32 is declared in win32_windows.go.
	procRtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

// IIDs for the interfaces we need.
var (
	// IGlobalSystemMediaTransportControlsSessionManagerStatics
	// Source: Windows SDK, Windows.Media.Control namespace, introduced in
	// Windows 10 version 1903 (UniversalApiContract version 7).
	iidSessionManagerStatics = windows.GUID{
		Data1: 0x87f8a6a9,
		Data2: 0x10ca,
		Data3: 0x5a82,
		Data4: [8]byte{0xbe, 0x6a, 0x2c, 0xf6, 0x5e, 0x89, 0xc0, 0x4f},
	}

	// IAsyncInfo — well-known WinRT base interface IID.
	iidAsyncInfo = windows.GUID{
		Data1: 0x00000036,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
)

// comCall invokes the COM method at vtable index idx on the object at ptr,
// forwarding additional args after the implicit `this` pointer.
// ptr must be a non-nil unsafe.Pointer to a COM interface object.
// Returns the raw HRESULT.
//
// The vtable is accessed through Go struct field reads (comObj → comVTable),
// which are fully managed by the type system and do not require unsafe
// uintptr arithmetic.
func comCall(ptr unsafe.Pointer, idx uintptr, args ...uintptr) uintptr {
	obj := (*comObj)(ptr)
	fn := obj.vtable.methods[idx]
	all := make([]uintptr, 0, 1+len(args))
	all = append(all, uintptr(ptr)) // this (unsafe.Pointer → uintptr for syscall, rule 4)
	all = append(all, args...)
	ret, _, _ := syscall.SyscallN(fn, all...)
	return ret
}

// comRelease calls IUnknown::Release (vtable[2]) on ptr.
func comRelease(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	comCall(ptr, 2)
}

// hstring is a WinRT HSTRING handle (opaque pointer-sized value).
type hstring uintptr

// newHString creates a WinRT HSTRING from a Go string.
func newHString(s string) (hstring, error) {
	utf16, err := syscall.UTF16FromString(s)
	if err != nil {
		return 0, err
	}
	var hs hstring
	r1, _, _ := procWindowsCreateString.Call(
		uintptr(unsafe.Pointer(&utf16[0])),
		uintptr(len(utf16)-1), // exclude null terminator
		uintptr(unsafe.Pointer(&hs)),
	)
	if r1 != 0 {
		return 0, windows.Errno(r1)
	}
	return hs, nil
}

// deleteHString destroys a WinRT HSTRING, releasing its backing buffer.
func deleteHString(hs hstring) {
	if hs == 0 {
		return
	}
	procWindowsDeleteString.Call(uintptr(hs))
}

// hstringToString converts a WinRT HSTRING to a Go string.
// The HSTRING buffer is copied into a Go-owned []uint16 via RtlMoveMemory,
// avoiding any uintptr→unsafe.Pointer conversion.
func hstringToString(hs hstring) string {
	if hs == 0 {
		return ""
	}
	var length uint32
	ptr, _, _ := procWindowsGetStringRawBuffer.Call(
		uintptr(hs),
		uintptr(unsafe.Pointer(&length)),
	)
	if ptr == 0 || length == 0 {
		return ""
	}
	// Allocate a Go-owned buffer and copy via RtlMoveMemory so that no
	// uintptr→unsafe.Pointer conversion is required.
	buf := make([]uint16, length)
	procRtlMoveMemory.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		ptr,
		uintptr(length)*2, // size in bytes (UTF-16 = 2 bytes/char)
	)
	return syscall.UTF16ToString(buf)
}

// waitForAsync blocks until the IAsyncOperation at asyncOp reaches a terminal
// state, polling IAsyncInfo::get_Status at 1 ms intervals.
// Times out after 5 seconds.
func waitForAsync(asyncOp unsafe.Pointer) error {
	var asyncInfo unsafe.Pointer
	hr := comCall(asyncOp, 0, // IUnknown::QueryInterface
		uintptr(unsafe.Pointer(&iidAsyncInfo)),
		uintptr(unsafe.Pointer(&asyncInfo)),
	)
	if hr != 0 {
		return fmt.Errorf("QueryInterface IAsyncInfo: 0x%08x", hr)
	}
	defer comRelease(asyncInfo)

	deadline := time.Now().Add(asyncTimeout)
	for time.Now().Before(deadline) {
		var status uint32
		comCall(asyncInfo, 7, uintptr(unsafe.Pointer(&status))) // get_Status
		switch status {
		case asyncCompleted:
			return nil
		case asyncCanceled:
			return errors.New("async operation was cancelled")
		case asyncError:
			return errors.New("async operation failed")
		}
		time.Sleep(time.Millisecond)
	}
	return errors.New("async operation timed out")
}

// smtcQuery uses WinRT COM interop to query the System Media Transport Controls
// for the active Spotify media session and returns a newline-separated string:
//
//	"Artist\nTitle\nPlaybackStatus"
//
// where PlaybackStatus is one of "Playing", "Paused", or "Stopped".
// An error is returned when no Spotify SMTC session is active or when any
// WinRT call fails.
func smtcQuery() (string, error) {
	// Initialise WinRT on the current OS thread.
	// RO_INIT_MULTITHREADED = 1.  S_FALSE (1) means already initialised.
	hr, _, _ := procRoInitialize.Call(1)
	if hr != 0 && hr != 1 {
		return "", fmt.Errorf("RoInitialize: 0x%08x", hr)
	}
	defer procRoUninitialize.Call()

	// Create class-name HSTRING.
	classID, err := newHString("Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager")
	if err != nil {
		return "", fmt.Errorf("WindowsCreateString: %w", err)
	}
	defer deleteHString(classID)

	// Obtain the statics activation factory.
	var factory unsafe.Pointer
	hr, _, _ = procRoGetActivationFactory.Call(
		uintptr(classID),
		uintptr(unsafe.Pointer(&iidSessionManagerStatics)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if hr != 0 {
		return "", fmt.Errorf("RoGetActivationFactory: 0x%08x", hr)
	}
	defer comRelease(factory)

	// factory::RequestAsync() → IAsyncOperation<SessionManager>  [vtable 6]
	var asyncOp unsafe.Pointer
	if hr = comCall(factory, 6, uintptr(unsafe.Pointer(&asyncOp))); hr != 0 {
		return "", fmt.Errorf("RequestAsync: 0x%08x", hr)
	}
	defer comRelease(asyncOp)

	if err = waitForAsync(asyncOp); err != nil {
		return "", fmt.Errorf("await RequestAsync: %w", err)
	}

	// asyncOp::GetResults() → SessionManager  [vtable 8]
	var manager unsafe.Pointer
	if hr = comCall(asyncOp, 8, uintptr(unsafe.Pointer(&manager))); hr != 0 {
		return "", fmt.Errorf("GetResults (manager): 0x%08x", hr)
	}
	defer comRelease(manager)

	// manager::GetSessions() → IVectorView<Session>  [vtable 7]
	var sessions unsafe.Pointer
	if hr = comCall(manager, 7, uintptr(unsafe.Pointer(&sessions))); hr != 0 {
		return "", fmt.Errorf("GetSessions: 0x%08x", hr)
	}
	defer comRelease(sessions)

	// sessions::get_Size() → uint32  [vtable 7]
	var size uint32
	if hr = comCall(sessions, 7, uintptr(unsafe.Pointer(&size))); hr != 0 {
		return "", fmt.Errorf("get_Size: 0x%08x", hr)
	}

	for i := uint32(0); i < size; i++ {
		// sessions::GetAt(i) → Session  [vtable 6]
		var session unsafe.Pointer
		if hr = comCall(sessions, 6, uintptr(i), uintptr(unsafe.Pointer(&session))); hr != 0 {
			continue
		}

		result, err := trySpotifySession(session)
		comRelease(session)

		if err == nil {
			return result, nil
		}
		// Not a Spotify session or failed: try the next one.
	}

	return "", errors.New("no active Spotify SMTC session found")
}

// trySpotifySession returns "Artist\nTitle\nStatus" if session belongs to
// Spotify.  It returns an error when the session does not match Spotify or
// when reading its properties fails.
func trySpotifySession(session unsafe.Pointer) (string, error) {
	// session::get_SourceAppUserModelId() → HSTRING  [vtable 6]
	var appIDHS hstring
	if hr := comCall(session, 6, uintptr(unsafe.Pointer(&appIDHS))); hr != 0 {
		return "", fmt.Errorf("get_SourceAppUserModelId: 0x%08x", hr)
	}
	appID := hstringToString(appIDHS)
	deleteHString(appIDHS)

	if !strings.HasPrefix(strings.ToLower(appID), "spotify") {
		return "", errors.New("not a Spotify session")
	}

	// session::TryGetMediaPropertiesAsync()  [vtable 7]
	var propsAsyncOp unsafe.Pointer
	if hr := comCall(session, 7, uintptr(unsafe.Pointer(&propsAsyncOp))); hr != 0 {
		return "", fmt.Errorf("TryGetMediaPropertiesAsync: 0x%08x", hr)
	}
	defer comRelease(propsAsyncOp)

	if err := waitForAsync(propsAsyncOp); err != nil {
		return "", fmt.Errorf("await TryGetMediaPropertiesAsync: %w", err)
	}

	// propsAsyncOp::GetResults() → MediaProperties  [vtable 8]
	var props unsafe.Pointer
	if hr := comCall(propsAsyncOp, 8, uintptr(unsafe.Pointer(&props))); hr != 0 {
		return "", fmt.Errorf("GetResults (props): 0x%08x", hr)
	}
	defer comRelease(props)

	// props::get_Title()  [vtable 6]
	var titleHS hstring
	if hr := comCall(props, 6, uintptr(unsafe.Pointer(&titleHS))); hr != 0 {
		return "", fmt.Errorf("get_Title: 0x%08x", hr)
	}
	title := hstringToString(titleHS)
	deleteHString(titleHS)

	// props::get_Artist()  [vtable 8]
	var artistHS hstring
	if hr := comCall(props, 8, uintptr(unsafe.Pointer(&artistHS))); hr != 0 {
		return "", fmt.Errorf("get_Artist: 0x%08x", hr)
	}
	artist := hstringToString(artistHS)
	deleteHString(artistHS)

	// session::GetPlaybackInfo()  [vtable 9]
	var playbackInfo unsafe.Pointer
	if hr := comCall(session, 9, uintptr(unsafe.Pointer(&playbackInfo))); hr != 0 {
		return "", fmt.Errorf("GetPlaybackInfo: 0x%08x", hr)
	}
	defer comRelease(playbackInfo)

	// playbackInfo::get_PlaybackStatus()  [vtable 10]
	var playbackStatus int32
	if hr := comCall(playbackInfo, 10, uintptr(unsafe.Pointer(&playbackStatus))); hr != 0 {
		return "", fmt.Errorf("get_PlaybackStatus: 0x%08x", hr)
	}

	var statusStr string
	switch playbackStatus {
	case smtcStatusPlaying:
		statusStr = "Playing"
	case smtcStatusPaused:
		statusStr = "Paused"
	default:
		statusStr = "Stopped"
	}

	return fmt.Sprintf("%s\n%s\n%s", artist, title, statusStr), nil
}
