package tasks

import (
	"debug/pe"
	"os"
	"path/filepath"
	"testing"
)

func TestGetPEFileVersion(t *testing.T) {
	files := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{path: "fixtures/Fake.dll", want: "", wantErr: true},
		{path: "fixtures/NewRelic.Agent.Extensions.dll", want: "6.17.387.0"},
		// Built by a current toolchain and taken from the Linux agent package, which ships the same managed PE assemblies
		{path: "fixtures/NewRelic.Api.Agent.linux-10.54.dll", want: "10.54.0.46"},
		{path: "fixtures/versionedtester.exe", want: "1.0.0.2"},
		{path: "fixtures/unversionedtester.exe", want: "", wantErr: true},
		{path: "fixtures/fixtures.md", want: "", wantErr: true},
	}

	for _, file := range files {
		t.Run(file.path, func(t *testing.T) {
			version, err := GetPEFileVersion(file.path)
			if (err != nil) != file.wantErr {
				t.Errorf("GetPEFileVersion() error = %v, wantErr %v", err, file.wantErr)
			}
			if version != file.want {
				t.Errorf("GetPEFileVersion() = %v, want %v", version, file.want)
			}
		})
	}
}

func TestGetPEFileVersion_NonPEFiles(t *testing.T) {
	// Minimal ELF header (as found on a native Linux .so) padded out so it isn't rejected only for being short
	elf := append([]byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0}, make([]byte, 120)...)

	files := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "single byte", data: []byte{0x4d}},
		{name: "dos signature only", data: []byte("MZ")},
		{name: "elf", data: elf},
	}

	for _, file := range files {
		t.Run(file.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file.bin")
			if err := os.WriteFile(path, file.data, 0o600); err != nil {
				t.Fatal(err)
			}

			version, err := GetPEFileVersion(path)
			if err == nil {
				t.Errorf("GetPEFileVersion() expected an error, got version %q", version)
			}
			if version != "" {
				t.Errorf("GetPEFileVersion() = %q, want empty version", version)
			}
		})
	}
}

// Truncating a valid file at various points must never panic or return a bogus version; a cut inside the headers or
// the resource section has to surface as an error
func TestGetPEFileVersion_TruncatedFiles(t *testing.T) {
	source := "fixtures/NewRelic.Api.Agent.linux-10.54.dll"
	full, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := pe.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	rsrc := parsed.Section(".rsrc")
	_ = parsed.Close()
	if rsrc == nil {
		t.Fatalf("%s has no .rsrc section", source)
	}
	rsrcStart := int(rsrc.Offset)
	rsrcEnd := rsrcStart + int(rsrc.Size)

	cuts := []struct {
		name string
		size int
	}{
		{name: "inside DOS header", size: 32},
		{name: "inside PE headers", size: 200},
		{name: "before .rsrc starts", size: rsrcStart},
		{name: "mid .rsrc", size: rsrcStart + (rsrcEnd-rsrcStart)/2},
		{name: "one byte short of .rsrc end", size: rsrcEnd - 1},
	}

	for _, cut := range cuts {
		t.Run(cut.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "truncated.dll")
			if err := os.WriteFile(path, full[:cut.size], 0o600); err != nil {
				t.Fatal(err)
			}

			version, err := GetPEFileVersion(path)
			if err == nil {
				t.Errorf("GetPEFileVersion() expected an error for a file cut to %d of %d bytes, got version %q", cut.size, len(full), version)
			}
			if version != "" {
				t.Errorf("GetPEFileVersion() = %q, want empty version", version)
			}
		})
	}
}

func TestFindVersionResource_MalformedData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "nil", data: nil},
		{name: "shorter than a directory header", data: make([]byte, 8)},
		{name: "declares an entry beyond the data", data: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0}},
		{name: "no version type entry", data: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 3, 0, 0, 0, 0, 0, 0, 0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := findVersionResource(test.data, 0x1000); err == nil {
				t.Error("findVersionResource() expected an error")
			}
		})
	}
}

func TestParseFixedFileInfo_MalformedData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "nil", data: nil},
		{name: "no signature", data: make([]byte, 64)},
		{name: "signature too close to the end for the version fields", data: []byte{0xBD, 0x04, 0xEF, 0xFE, 0, 0, 0, 0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if version, err := parseFixedFileInfo(test.data); err == nil {
				t.Errorf("parseFixedFileInfo() expected an error, got %q", version)
			}
		})
	}
}
