package window

import (
	"strings"
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
)

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
	//fmt.Println("Process Name:", procName)
	if procName == "snippingtool.exe" || procName == "screenclippinghost.exe" || procName == "applicationframehost.exe" || procName == "shellexperiencehost.exe" {
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

	return true
}
