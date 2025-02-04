//go:build windows
// +build windows

package tasks

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"golang.org/x/sys/windows/registry"
)

type FILEINFO struct {
	Signature           uint32
	StrucVersion        uint32
	MostSignificantBits uint32
	LessSignificantBits uint32
}

type USER_INFO_1 struct {
	Usri1_name         *uint16
	Usri1_password     *uint16
	Usri1_password_age uint32
	Usri1_priv         uint32
	Usri1_home_dir     *uint16
	Usri1_comment      *uint16
	Usri1_flags        uint32
	Usri1_script_path  *uint16
}

const USER_PRIV_ADMIN = 2

var (
	version                 = syscall.NewLazyDLL("version.dll")
	procFileVersionInfoSize = version.NewProc("GetFileVersionInfoSizeW")
	procFileVersionInfo     = version.NewProc("GetFileVersionInfoW")
	procVersionQueryValue   = version.NewProc("VerQueryValueW")
	modNetAPI32             = syscall.NewLazyDLL("netapi32.dll")
	usrNetGetAnyDCName      = modNetAPI32.NewProc("NetGetAnyDCName")
	usrNetUserGetInfo       = modNetAPI32.NewProc("NetUserGetInfo")
	usrNetAPIBufferFree     = modNetAPI32.NewProc("NetApiBufferFree")
)

// Function GetFileVersion() - Returns the internal version number from a windows dll or exe file
//
// Usage: GetFileVersion(file)
// Args: file- String with the full path of the file. i.e. "C:\Program Files\New Relic\.NET Agent\NewRelic.Agent.Core.dll".
// Returns: string - value of full internal version number.
//
//	error  - any error message encountered. `nil` if none.
type GetFileVersionFunc func(string) (string, error)

func GetFileVersion(file string) (string, error) {
	if !FileExists(file) {
		return "", errors.New("file does not exist")
	}

	size := fileVersionInfoSize(file)
	if size == 0 {
		return "", errors.New("no permissions or No version information found")
	}

	info := make([]byte, size)
	ok := fileVersionInfo(file, info)
	if !ok {
		return "", errors.New("getFileVersionInfo failed")
	}

	parameters, ok := queryInfoValue(info)
	if !ok {
		return "", errors.New("queryInfoValue failed")
	}
	version := parameters.fileVersion()

	// Bitwise rotation to return each of the octets correctly
	return fmt.Sprintf("%d.%d.%d.%d",
		version&0xFFFF000000000000>>48,
		version&0x0000FFFF00000000>>32,
		version&0x00000000FFFF0000>>16,
		version&0x000000000000FFFF>>0,
	), nil
}

// FileVersion concatenates MostSignificantBits and LessSignificantBits to a uint64 value.
func (fv FILEINFO) fileVersion() uint64 {
	return uint64(fv.MostSignificantBits)<<32 | uint64(fv.LessSignificantBits)
}

// Calls the exported method GetFileVersionInfo of version.dll
// Retrieves Version Information into the data slice of bytes
//
// see https://msdn.microsoft.com/en-us/library/windows/desktop/ms647003(v=vs.85).aspx
func fileVersionInfo(path string, data []byte) bool {
	ret, _, _ := procFileVersionInfo.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		0,
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&data[0])),
	)
	return ret != 0
}

// Calls the exported method GetFileVersionInfoSize of version.dll
// Check the existence of Version Information on the file and returns the size in bytes of it
//
// see https://msdn.microsoft.com/en-us/library/windows/desktop/ms647005(v=vs.85).aspx
func fileVersionInfoSize(path string) uint32 {
	ret, _, _ := procFileVersionInfoSize.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		0,
	)
	return uint32(ret)
}

// Calls the exported method VerQueryValue of version.dll
// Populates the FILEINFO struct from block
// Finds the start and length of the version information info and extracts the slice
//
// see https://msdn.microsoft.com/en-us/library/windows/desktop/ms647464(v=vs.85).aspx
func queryInfoValue(block []byte) (FILEINFO, bool) {
	var offset uintptr
	var length uint
	blockStart := uintptr(unsafe.Pointer(&block[0]))
	ret, _, _ := procVersionQueryValue.Call(
		blockStart,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(`\`))),
		uintptr(unsafe.Pointer(&offset)),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret == 0 {
		return FILEINFO{}, false
	}
	start := int(offset) - int(blockStart)
	end := start + int(length)
	if start < 0 || start >= len(block) || end < start || end > len(block) {
		return FILEINFO{}, false
	}
	data := block[start:end]
	info := *((*FILEINFO)(unsafe.Pointer(&data[0])))
	return info, true
}

type GetProcessorArchFunc func() (string, error)

// Checks the Windows machine's Processor Architecture. Will return x86 for 32 bit systems and AMD64 for 64 bit systems.
// Can return other types or undefined for some types of non-x86 and non-x64 processors
// best practice is to explicitly check for x86 or AMD64
func GetProcessorArch() (procType string, errorReg error) {

	var procTypeRegLoc = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	var subKeyName = `PROCESSOR_ARCHITECTURE`

	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, procTypeRegLoc, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug("Error opening Environment Reg Key. Error = ", err.Error())
		return "", err
	}

	defer regKey.Close()
	regValue, _, regErr := regKey.GetStringValue(subKeyName)

	if regErr != nil {
		log.Debug("Error opening PROCESSOR_ARCHITECTURE Reg Sub Key. Error = ", string(regErr.Error()))
		return "", regErr
	}
	return regValue, nil

}

// ProcInfoStruct - Struct to be used to hold process info
type ProcInfoStruct struct {
	Name           string
	CommandLine    string
	ExecutablePath string
	ProcessId      uint32
}

// GetProcInfoByPid - Returns the ProcInfoStruct of the process, if not found returns an error and struct with empty strings and a pid of 0
func GetProcInfoByPid(pid string) (ProcInfoStruct, error) {
	var procInfoStruct []ProcInfoStruct
	blankInfo := ProcInfoStruct{"", "", "", 0}
	query := "SELECT Name,CommandLine,ExecutablePath,ProcessId FROM Win32_Process WHERE ProcessId = " + pid
	err := wmi.Query(query, &procInfoStruct)
	if len(procInfoStruct) < 1 {
		procInfoStruct = append(procInfoStruct, blankInfo)
	}
	if err != nil {
		err = errors.New("no matching process found ")
		return procInfoStruct[0], err
	}
	return procInfoStruct[0], err
}

// GetProcInfoByName - Returns a slice of ProcInfoStructs and any errors that match the passed in name. Uses SQL for WMI LIKE Operator https://msdn.microsoft.com/en-us/library/aa392263(v=vs.85).aspx
func GetProcInfoByName(name string) ([]ProcInfoStruct, error) {
	var procInfoStruct []ProcInfoStruct
	query := "SELECT Name,CommandLine,ExecutablePath,ProcessId FROM Win32_Process WHERE name LIKE \"" + name + "\""
	err := wmi.Query(query, &procInfoStruct)
	if err == nil && len(procInfoStruct) < 1 {
		err = errors.New("no matching process found ")
		return procInfoStruct, err
	}
	return procInfoStruct, err
}

// IsUserAdmin - uses netapi32.dll to determine if user is admin (checks domain first then local)
func IsUserAdmin(username string, domain string) (bool, error) {
	var dataPointer uintptr
	var dcPointer uintptr
	uPointer, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		log.Debug("Unable to encode username to UTF16.")
		return false, err
	}
	dPointer, err := syscall.UTF16PtrFromString(domain)
	if err != nil {
		log.Debug("Unable to encode domain to UTF16.")
		return false, err
	}

	_, _, _ = usrNetGetAnyDCName.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(dPointer)),
		uintptr(unsafe.Pointer(&dcPointer)),
	)
	defer usrNetAPIBufferFree.Call(uintptr(dcPointer))

	_, _, _ = usrNetUserGetInfo.Call(
		uintptr(dcPointer),
		uintptr(unsafe.Pointer(uPointer)),
		uintptr(uint32(1)),
		uintptr(unsafe.Pointer(&dataPointer)),
	)
	defer usrNetAPIBufferFree.Call(uintptr(dataPointer))

	if dataPointer == uintptr(0) {
		log.Debug("Unable to determine domain user, retrying as local")
		return isLocalUserAdmin(username)
	}

	var data = (*USER_INFO_1)(unsafe.Pointer(dataPointer))

	if data.Usri1_priv == USER_PRIV_ADMIN {
		return true, nil
	}
	return false, nil
}

func isLocalUserAdmin(username string) (bool, error) {
	var dataPointer uintptr
	uPointer, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		log.Debug("Unable to encode username to UTF16.")
		return false, err
	}
	_, _, _ = usrNetUserGetInfo.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(uPointer)),
		uintptr(uint32(1)),
		uintptr(unsafe.Pointer(&dataPointer)),
	)
	defer usrNetAPIBufferFree.Call(uintptr(dataPointer))

	if dataPointer == uintptr(0) {
		log.Debug("Unable to determine local user.")
		return false, errors.New("unable to determine local user.")
	}

	var data = (*USER_INFO_1)(unsafe.Pointer(dataPointer))

	if data.Usri1_priv == USER_PRIV_ADMIN {
		return true, nil
	}

	return false, nil
}
