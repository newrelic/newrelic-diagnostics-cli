package tasks

import (
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	rtVersion              = 16         // RT_VERSION resource type ID
	vsFixedFileInfoSig     = 0xFEEF04BD // VS_FIXEDFILEINFO.dwSignature
	resourceSubdirFlag     = 0x80000000
	resourceDirHeaderSize  = 16
	resourceDirEntrySize   = 8
	resourceDataEntrySize  = 16
	maxResourceDirectories = 3 // type -> name -> language
)

// GetFileVersionFunc - Signature shared by GetFileVersion and GetPEFileVersion, so tasks can swap in a fake for tests
type GetFileVersionFunc func(string) (string, error)

// GetPEFileVersion - Returns the file version (e.g. "10.45.0.0") from the VS_VERSIONINFO resource of a PE (dll or exe) file.
// Unlike GetFileVersion, this reads the file directly and works on any OS.
func GetPEFileVersion(file string) (string, error) {
	f, err := pe.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	rsrc := f.Section(".rsrc")
	if rsrc == nil {
		return "", errors.New("no resource section found")
	}
	data, err := rsrc.Data()
	if err != nil {
		return "", err
	}

	versionInfo, err := findVersionResource(data, rsrc.VirtualAddress)
	if err != nil {
		return "", err
	}
	return parseFixedFileInfo(versionInfo)
}

// findVersionResource walks the resource directory tree and returns the raw VS_VERSIONINFO bytes
func findVersionResource(rsrc []byte, rsrcVirtualAddress uint32) ([]byte, error) {
	offset := uint32(0)
	for level := 0; level < maxResourceDirectories; level++ {
		entry, err := findResourceDirEntry(rsrc, offset, level == 0)
		if err != nil {
			return nil, err
		}
		if entry&resourceSubdirFlag == 0 {
			if level != maxResourceDirectories-1 {
				return nil, errors.New("unexpected resource data entry")
			}
			offset = entry
			break
		}
		offset = entry &^ resourceSubdirFlag
		if level == maxResourceDirectories-1 {
			return nil, errors.New("unexpected resource subdirectory")
		}
	}

	if uint64(offset)+resourceDataEntrySize > uint64(len(rsrc)) {
		return nil, errors.New("resource data entry out of range")
	}
	dataRVA := binary.LittleEndian.Uint32(rsrc[offset:])
	dataSize := binary.LittleEndian.Uint32(rsrc[offset+4:])
	if dataRVA < rsrcVirtualAddress {
		return nil, errors.New("version resource out of range")
	}
	start := uint64(dataRVA - rsrcVirtualAddress)
	end := start + uint64(dataSize)
	if end > uint64(len(rsrc)) {
		return nil, errors.New("version resource out of range")
	}
	return rsrc[start:end], nil
}

// findResourceDirEntry returns the OffsetToData of the RT_VERSION entry (at the type level) or the first entry (at lower levels)
func findResourceDirEntry(rsrc []byte, dirOffset uint32, isTypeLevel bool) (uint32, error) {
	if uint64(dirOffset)+resourceDirHeaderSize > uint64(len(rsrc)) {
		return 0, errors.New("resource directory out of range")
	}
	namedEntries := uint32(binary.LittleEndian.Uint16(rsrc[dirOffset+12:]))
	idEntries := uint32(binary.LittleEndian.Uint16(rsrc[dirOffset+14:]))
	entriesStart := uint64(dirOffset) + resourceDirHeaderSize

	for i := uint32(0); i < namedEntries+idEntries; i++ {
		entryOffset := entriesStart + uint64(i)*resourceDirEntrySize
		if entryOffset+resourceDirEntrySize > uint64(len(rsrc)) {
			return 0, errors.New("resource directory entry out of range")
		}
		nameOrID := binary.LittleEndian.Uint32(rsrc[entryOffset:])
		if !isTypeLevel || nameOrID == rtVersion {
			return binary.LittleEndian.Uint32(rsrc[entryOffset+4:]), nil
		}
	}
	return 0, errors.New("no version information found")
}

// parseFixedFileInfo finds the VS_FIXEDFILEINFO structure in a VS_VERSIONINFO resource and formats its file version
func parseFixedFileInfo(versionInfo []byte) (string, error) {
	// VS_FIXEDFILEINFO is DWORD-aligned and starts with its signature, followed by dwStrucVersion, dwFileVersionMS and dwFileVersionLS
	for i := 0; i+16 <= len(versionInfo); i += 4 {
		if binary.LittleEndian.Uint32(versionInfo[i:]) != vsFixedFileInfoSig {
			continue
		}
		versionMS := binary.LittleEndian.Uint32(versionInfo[i+8:])
		versionLS := binary.LittleEndian.Uint32(versionInfo[i+12:])
		return fmt.Sprintf("%d.%d.%d.%d", versionMS>>16, versionMS&0xFFFF, versionLS>>16, versionLS&0xFFFF), nil
	}
	return "", errors.New("no fixed file version information found")
}
