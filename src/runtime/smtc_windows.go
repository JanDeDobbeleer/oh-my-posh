//go:build windows

package runtime

import (
	"errors"
	"fmt"
	goruntime "runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// querySpotifySMTC reads the active Spotify session from the Windows System
// Media Transport Controls (SMTC) using a minimal in-tree WinRT binding —
// `combase.dll` exports + COM vtable dispatch. It returns the line
//
//	"<status>|<title>|<artist>|<album>|<trackNumber>"
//
// which the segment-side parseSMTCOutput already understands, including the
// `playing && album=="" && trackNumber=="0"` ad-detection heuristic.
//
// Vtable indices and IIDs were verified by .NET reflection against
// Windows.Media.Control.winmd on Windows 11 (see the table in the const
// blocks below). All WinRT interfaces inherit from IInspectable, which sits
// on top of IUnknown:
//
//	IUnknown      [0] QueryInterface  [1] AddRef  [2] Release
//	IInspectable  [3] GetIids  [4] GetRuntimeClassName  [5] GetTrustLevel
//
// so interface-specific methods always start at slot 6.

// COM vtable layout. The actual object in memory starts with a pointer to
// its vtable; the vtable starts with N function pointers. Modelling both as
// Go structs keeps `go vet`'s unsafe rules satisfied — we navigate by struct
// field access rather than uintptr arithmetic.
type comVTable struct {
	methods [64]uintptr
}

type comObj struct {
	vtable *comVTable
}

const (
	iunknownQueryInterface = 0
	iunknownRelease        = 2

	staticsRequestAsync = 6 // IGlobalSystemMediaTransportControlsSessionManagerStatics

	managerGetSessions = 7 // IGlobalSystemMediaTransportControlsSessionManager

	vectorViewGetAt   = 6 // IVectorView<T>
	vectorViewGetSize = 7

	sessionGetSourceAppUserModelID    = 6 // IGlobalSystemMediaTransportControlsSession
	sessionTryGetMediaPropertiesAsync = 7
	sessionGetPlaybackInfo            = 9

	playbackInfoGetPlaybackStatus = 7 // IGlobalSystemMediaTransportControlsSessionPlaybackInfo

	mediaPropsGetTitle       = 6 // IGlobalSystemMediaTransportControlsSessionMediaProperties
	mediaPropsGetArtist      = 9
	mediaPropsGetAlbumTitle  = 10
	mediaPropsGetTrackNumber = 11

	asyncOpGetResults  = 8 // IAsyncOperation<T>
	asyncInfoGetStatus = 7 // IAsyncInfo
)

// GlobalSystemMediaTransportControlsSessionPlaybackStatus enum values.
const (
	smtcStatusClosed   = 0
	smtcStatusOpened   = 1
	smtcStatusChanging = 2
	smtcStatusStopped  = 3
	smtcStatusPlaying  = 4
	smtcStatusPaused   = 5
)

// Windows.Foundation.AsyncStatus enum values.
const (
	asyncStatusCompleted = 1
	asyncStatusCanceled  = 2
	asyncStatusError     = 3
)

const (
	smtcAsyncTimeout      = 5 * time.Second
	smtcAsyncPollInterval = 5 * time.Millisecond
)

var (
	// IGlobalSystemMediaTransportControlsSessionManagerStatics
	iidSpotifySessionManagerStatics = windows.GUID{
		Data1: 0x2050c4ee,
		Data2: 0x11a0,
		Data3: 0x57de,
		Data4: [8]byte{0xae, 0xd7, 0xc9, 0x7c, 0x70, 0x33, 0x82, 0x45},
	}
	// IAsyncInfo
	iidSpotifyAsyncInfo = windows.GUID{
		Data1: 0x00000036,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
)

var (
	combase                       = windows.NewLazySystemDLL("combase.dll")
	procRoInitialize              = combase.NewProc("RoInitialize")
	procRoUninitialize            = combase.NewProc("RoUninitialize")
	procRoGetActivationFactory    = combase.NewProc("RoGetActivationFactory")
	procWindowsCreateString       = combase.NewProc("WindowsCreateString")
	procWindowsDeleteString       = combase.NewProc("WindowsDeleteString")
	procWindowsGetStringRawBuffer = combase.NewProc("WindowsGetStringRawBuffer")

	// kernel32 is declared in win32_windows.go.
	procRtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

// hstring is a WinRT HSTRING handle. Pointer-sized opaque value — must be
// freed with WindowsDeleteString once the consumer is done with it.
type hstring uintptr

func newHString(s string) (hstring, error) {
	utf16, err := syscall.UTF16FromString(s)
	if err != nil {
		return 0, err
	}
	var hs hstring
	r1, _, _ := procWindowsCreateString.Call(
		uintptr(unsafe.Pointer(&utf16[0])),
		uintptr(len(utf16)-1),
		uintptr(unsafe.Pointer(&hs)),
	)
	if r1 != 0 {
		return 0, fmt.Errorf("WindowsCreateString: 0x%08x", r1)
	}
	return hs, nil
}

func (hs hstring) String() string {
	if hs == 0 {
		return ""
	}
	var length uint32
	ptr, _, _ := procWindowsGetStringRawBuffer.Call(uintptr(hs), uintptr(unsafe.Pointer(&length)))
	if ptr == 0 || length == 0 {
		return ""
	}
	// Copy WinRT-owned UTF-16 into a Go-owned slice via RtlMoveMemory so we
	// never construct an unsafe.Pointer from the raw buffer's uintptr.
	out := make([]uint16, length)
	_, _, _ = procRtlMoveMemory.Call(
		uintptr(unsafe.Pointer(&out[0])),
		ptr,
		uintptr(length)*2, // UTF-16 = 2 bytes per code unit
	)
	return syscall.UTF16ToString(out)
}

func (hs hstring) Close() {
	if hs == 0 {
		return
	}
	_, _, _ = procWindowsDeleteString.Call(uintptr(hs))
}

// comCall invokes vtable[idx] on the COM object at ptr, with `this` and the
// given arguments.
func comCall(ptr unsafe.Pointer, idx uintptr, args ...uintptr) uintptr {
	fn := (*comObj)(ptr).vtable.methods[idx]
	all := make([]uintptr, 0, 1+len(args))
	all = append(all, uintptr(ptr))
	all = append(all, args...)
	ret, _, _ := syscall.SyscallN(fn, all...)
	return ret
}

func comRelease(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	comCall(ptr, iunknownRelease)
}

// awaitAsync polls IAsyncInfo::Status until the async operation reaches a
// terminal state. We poll instead of registering a Completed handler to
// avoid implementing a COM callback object in Go.
func awaitAsync(asyncOp unsafe.Pointer) error {
	var asyncInfo unsafe.Pointer
	hr := comCall(asyncOp, iunknownQueryInterface,
		uintptr(unsafe.Pointer(&iidSpotifyAsyncInfo)),
		uintptr(unsafe.Pointer(&asyncInfo)),
	)
	if hr != 0 {
		return fmt.Errorf("QueryInterface IAsyncInfo: 0x%08x", hr)
	}
	defer comRelease(asyncInfo)

	deadline := time.Now().Add(smtcAsyncTimeout)
	for {
		var status uint32
		comCall(asyncInfo, asyncInfoGetStatus, uintptr(unsafe.Pointer(&status)))
		switch status {
		case asyncStatusCompleted:
			return nil
		case asyncStatusCanceled:
			return errors.New("async cancelled")
		case asyncStatusError:
			return errors.New("async failed")
		}
		if time.Now().After(deadline) {
			return errors.New("async timed out")
		}
		time.Sleep(smtcAsyncPollInterval)
	}
}

func querySpotifySMTC() (string, error) {
	// RoInitialize / RoUninitialize update thread-local state, so we have to
	// pin the goroutine to one OS thread for the duration of this call.
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	// RO_INIT_MULTITHREADED = 1. S_FALSE (1) means already initialised on
	// this thread, which is fine — we still pair with RoUninitialize.
	hr, _, _ := procRoInitialize.Call(1)
	if hr != 0 && hr != 1 {
		return "", fmt.Errorf("RoInitialize: 0x%08x", hr)
	}
	defer func() { _, _, _ = procRoUninitialize.Call() }()

	classID, err := newHString("Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager")
	if err != nil {
		return "", err
	}
	defer classID.Close()

	var factory unsafe.Pointer
	hr, _, _ = procRoGetActivationFactory.Call(
		uintptr(classID),
		uintptr(unsafe.Pointer(&iidSpotifySessionManagerStatics)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if hr != 0 {
		return "", fmt.Errorf("RoGetActivationFactory: 0x%08x", hr)
	}
	defer comRelease(factory)

	var asyncOp unsafe.Pointer
	if hr := comCall(factory, staticsRequestAsync, uintptr(unsafe.Pointer(&asyncOp))); hr != 0 {
		return "", fmt.Errorf("RequestAsync: 0x%08x", hr)
	}
	defer comRelease(asyncOp)

	if err := awaitAsync(asyncOp); err != nil {
		return "", fmt.Errorf("await RequestAsync: %w", err)
	}

	var manager unsafe.Pointer
	if hr := comCall(asyncOp, asyncOpGetResults, uintptr(unsafe.Pointer(&manager))); hr != 0 {
		return "", fmt.Errorf("GetResults manager: 0x%08x", hr)
	}
	defer comRelease(manager)

	var sessions unsafe.Pointer
	if hr := comCall(manager, managerGetSessions, uintptr(unsafe.Pointer(&sessions))); hr != 0 {
		return "", fmt.Errorf("GetSessions: 0x%08x", hr)
	}
	defer comRelease(sessions)

	var size uint32
	if hr := comCall(sessions, vectorViewGetSize, uintptr(unsafe.Pointer(&size))); hr != 0 {
		return "", fmt.Errorf("get_Size: 0x%08x", hr)
	}

	for i := uint32(0); i < size; i++ {
		var session unsafe.Pointer
		if hr := comCall(sessions, vectorViewGetAt, uintptr(i), uintptr(unsafe.Pointer(&session))); hr != 0 {
			continue
		}
		result, ok := readSpotifySession(session)
		comRelease(session)
		if ok {
			return result, nil
		}
	}

	// No Spotify session in the SMTC list — match the contract the existing
	// PowerShell script returns when the manager has nothing matching.
	return "closed||||0", nil
}

func readSpotifySession(session unsafe.Pointer) (string, bool) {
	var appHS hstring
	if hr := comCall(session, sessionGetSourceAppUserModelID, uintptr(unsafe.Pointer(&appHS))); hr != 0 {
		return "", false
	}
	appID := appHS.String()
	appHS.Close()

	if !strings.Contains(strings.ToLower(appID), "spotify") {
		return "", false
	}

	var playbackInfo unsafe.Pointer
	if hr := comCall(session, sessionGetPlaybackInfo, uintptr(unsafe.Pointer(&playbackInfo))); hr != 0 {
		return "", false
	}
	defer comRelease(playbackInfo)

	var status int32
	if hr := comCall(playbackInfo, playbackInfoGetPlaybackStatus, uintptr(unsafe.Pointer(&status))); hr != 0 {
		return "", false
	}

	statusStr := smtcStatusString(status)

	// Title/artist/album/trackNumber are best-effort; if the async op fails
	// we still return the status so the segment can hide cleanly.
	title, artist, album, trackNumber := readMediaProperties(session)
	return fmt.Sprintf("%s|%s|%s|%s|%d", statusStr, title, artist, album, trackNumber), true
}

func smtcStatusString(s int32) string {
	switch s {
	case smtcStatusPlaying:
		return "playing"
	case smtcStatusPaused:
		return "paused"
	case smtcStatusStopped:
		return "stopped"
	case smtcStatusOpened:
		return "opened"
	case smtcStatusChanging:
		return "changing"
	default:
		return "closed"
	}
}

func readMediaProperties(session unsafe.Pointer) (title, artist, album string, trackNumber int32) {
	var asyncOp unsafe.Pointer
	if hr := comCall(session, sessionTryGetMediaPropertiesAsync, uintptr(unsafe.Pointer(&asyncOp))); hr != 0 {
		return
	}
	defer comRelease(asyncOp)

	if err := awaitAsync(asyncOp); err != nil {
		return
	}

	var props unsafe.Pointer
	if hr := comCall(asyncOp, asyncOpGetResults, uintptr(unsafe.Pointer(&props))); hr != 0 {
		return
	}
	defer comRelease(props)

	title = readHStringMethod(props, mediaPropsGetTitle)
	artist = readHStringMethod(props, mediaPropsGetArtist)
	album = readHStringMethod(props, mediaPropsGetAlbumTitle)
	comCall(props, mediaPropsGetTrackNumber, uintptr(unsafe.Pointer(&trackNumber)))
	return
}

func readHStringMethod(obj unsafe.Pointer, slot uintptr) string {
	var hs hstring
	if hr := comCall(obj, slot, uintptr(unsafe.Pointer(&hs))); hr != 0 {
		return ""
	}
	s := hs.String()
	hs.Close()
	return s
}
