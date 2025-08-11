package window

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	procGetWindowLongPtrW   = user32.NewProc("GetWindowLongPtrW")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procGetWindow           = user32.NewProc("GetWindow")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	procGetClassNameW       = user32.NewProc("GetClassNameW")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	dwmapi                  = syscall.NewLazyDLL("dwmapi.dll")
	procDwmGetWindowAttr    = dwmapi.NewProc("DwmGetWindowAttribute")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImg = kernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	advapi32                = syscall.NewLazyDLL("advapi32.dll")
	procOpenProcessToken    = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation = advapi32.NewProc("GetTokenInformation")
)

const (
	GWL_STYLE                  = -16
	GWL_EXSTYLE                = -20
	WS_CHILD                   = 0x40000000
	WS_EX_TOOLWINDOW           = 0x00000080
	WS_EX_APPWINDOW            = 0x00040000
	WS_EX_TOPMOST              = 0x00000008
	WS_EX_LAYERED              = 0x00080000
	GW_OWNER                   = 4
	DWMWA_CLOAKED              = 14
	PROCESS_QUERY_LIMITED_INFO = 0x1000
	TOKEN_QUERY                = 0x0008
	TokenIntegrityLevel        = 25

	SECURITY_MANDATORY_LOW_RID    = 0x1000
	SECURITY_MANDATORY_MEDIUM_RID = 0x2000
	SECURITY_MANDATORY_HIGH_RID   = 0x3000
	SECURITY_MANDATORY_SYSTEM_RID = 0x4000
)

var currentIntegrityOnce sync.Once
var currentIntegrityLevel uint32

func getCurrentIntegrity() uint32 {
	currentIntegrityOnce.Do(func() {
		hProc, _ := syscall.GetCurrentProcess()
		if hProc == 0 {
			currentIntegrityLevel = SECURITY_MANDATORY_MEDIUM_RID
			return
		}
		var token syscall.Handle
		r, _, _ := procOpenProcessToken.Call(uintptr(hProc), uintptr(TOKEN_QUERY), uintptr(unsafe.Pointer(&token)))
		if r == 0 || token == 0 {
			currentIntegrityLevel = SECURITY_MANDATORY_MEDIUM_RID
			return
		}
		defer procCloseHandle.Call(uintptr(token))
		lvl, ok := readIntegrityLevel(token)
		if ok {
			currentIntegrityLevel = lvl
		} else {
			currentIntegrityLevel = SECURITY_MANDATORY_MEDIUM_RID
		}
	})
	return currentIntegrityLevel
}

func readIntegrityLevel(token syscall.Handle) (uint32, bool) {
	var retLen uint32

	procGetTokenInformation.Call(uintptr(token), uintptr(TokenIntegrityLevel), 0, 0, uintptr(unsafe.Pointer(&retLen)))
	if retLen == 0 {
		return 0, false
	}
	buf := make([]byte, retLen)
	r, _, _ := procGetTokenInformation.Call(uintptr(token), uintptr(TokenIntegrityLevel), uintptr(unsafe.Pointer(&buf[0])), uintptr(retLen), uintptr(unsafe.Pointer(&retLen)))
	if r == 0 {
		return 0, false
	}

	sidPtr := *(*uintptr)(unsafe.Pointer(&buf[0]))
	if sidPtr == 0 {
		return 0, false
	}
	maxSIDLen := 68
	type sliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}

	hdr := sliceHeader{Data: sidPtr, Len: maxSIDLen, Cap: maxSIDLen}
	sidBytes := *(*[]byte)(unsafe.Pointer(&hdr))
	subAuthCount := sidBytes[1]
	if subAuthCount == 0 {
		return 0, false
	}
	if subAuthCount > 15 { // sanity
		return 0, false
	}
	lastSubAuthOffset := 8 + int(subAuthCount-1)*4
	if lastSubAuthOffset+4 > len(sidBytes) {
		return 0, false
	}
	level := *(*uint32)(unsafe.Pointer(&sidBytes[lastSubAuthOffset]))
	return level, true
}

func getWindowLongPtr(hwnd uintptr, index int) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(hwnd, uintptr(index))
	return r
}

func isWindowVisible(hwnd uintptr) bool {
	r, _, _ := procIsWindowVisible.Call(hwnd)
	return r != 0
}

func getOwner(hwnd uintptr) uintptr {
	r, _, _ := procGetWindow.Call(hwnd, GW_OWNER)
	return r
}

func isCloaked(hwnd uintptr) bool {
	if procDwmGetWindowAttr.Find() != nil {
		return false
	}
	var cloaked uint32
	procDwmGetWindowAttr.Call(hwnd, DWMWA_CLOAKED, uintptr(unsafe.Pointer(&cloaked)), unsafe.Sizeof(cloaked))
	return cloaked != 0
}

func windowTitleLength(hwnd uintptr) int {
	r, _, _ := procGetWindowTextLength.Call(hwnd)
	return int(r)
}

func getClassName(hwnd uintptr) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

func getProcessName(hwnd uintptr) string {
	var pid uint32
	procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}

	h, _, _ := procOpenProcess.Call(PROCESS_QUERY_LIMITED_INFO, 0, uintptr(pid))
	if h == 0 {
		return ""
	}

	defer procCloseHandle.Call(h)
	buf := make([]uint16, 260)
	size := uint32(len(buf))

	procQueryFullProcessImg.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	full := syscall.UTF16ToString(buf)

	lastSlash := strings.LastIndexAny(full, "\\/")

	if lastSlash >= 0 && lastSlash+1 < len(full) {
		return strings.ToLower(full[lastSlash+1:])
	}
	return strings.ToLower(full)
}

func getWindowSize(hwnd uintptr) (int, int) {
	var r struct{ Left, Top, Right, Bottom int32 }

	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	return int(r.Right - r.Left), int(r.Bottom - r.Top)
}

func IsAppWindow(hwnd uintptr) bool {
	if !isWindowVisible(hwnd) {
		return false
	}
	if isCloaked(hwnd) {
		return false
	}

	style := getWindowLongPtr(hwnd, GWL_STYLE)
	exStyle := getWindowLongPtr(hwnd, GWL_EXSTYLE)
	if style&WS_CHILD != 0 { // child window
		return false
	}

	owner := getOwner(hwnd)
	isTool := exStyle&WS_EX_TOOLWINDOW != 0
	isApp := exStyle&WS_EX_APPWINDOW != 0

	if !(isApp || (owner == 0 && !isTool)) {
		return false
	}

	if windowTitleLength(hwnd) == 0 {
		return false
	}

	cls := getClassName(hwnd)
	if cls == "Progman" || cls == "Button" {
		return false
	}

	procName := getProcessName(hwnd)

	// exclude system stuff
	if procName == "snippingtool.exe" || procName == "searchhost.exe" || procName == "screenclippinghost.exe" || procName == "applicationframehost.exe" || procName == "shellexperiencehost.exe" {
		return false
	}

	if strings.Contains(strings.ToLower(cls), "snip") || strings.Contains(strings.ToLower(cls), "clipping") {
		return false
	}

	if (exStyle&WS_EX_TOPMOST) != 0 && (exStyle&WS_EX_LAYERED) != 0 {
		w, h := getWindowSize(hwnd)
		if h <= 100 && w < 800 {
			return false
		}
	}

	if procName == "explorer.exe" {
		lc := strings.ToLower(cls)
		if lc != "cabinetwclass" { // allow only main explorer windows and not stuff like the run window
			return false
		}
	}

	if isHigherLevelProcess(hwnd) {
		return false
	}

	// fmt.Println(procName)

	return true
}

func isHigherLevelProcess(hwnd uintptr) bool {
	var pid uint32
	procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return false
	}
	h, _, _ := procOpenProcess.Call(PROCESS_QUERY_LIMITED_INFO, 0, uintptr(pid))
	if h == 0 {
		return false // can't open, assume not higher (or we lack rights)
	}
	defer procCloseHandle.Call(h)
	var token syscall.Handle
	r, _, _ := procOpenProcessToken.Call(h, uintptr(TOKEN_QUERY), uintptr(unsafe.Pointer(&token)))
	if r == 0 || token == 0 {
		return false
	}
	defer procCloseHandle.Call(uintptr(token))
	lvl, ok := readIntegrityLevel(token)
	if !ok {
		return false
	}
	myLvl := getCurrentIntegrity()
	return lvl > myLvl
}
