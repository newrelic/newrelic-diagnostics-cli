package config

import "testing"

func TestIsAppSettingsJSON(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "appsettings.json", want: true},
		{filename: "AppSettings.json", want: true},
		{filename: "appsettings.Production.json", want: true},
		{filename: "appsettings.Development.json", want: true},
		{filename: "appsettings.json.bak", want: false},
		{filename: "myappsettings.json", want: false},
		{filename: "appsettings.Production.json.old", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Act
			got := IsAppSettingsJSON(tt.filename)

			// Assert
			if got != tt.want {
				t.Errorf("IsAppSettingsJSON(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestEnvironmentSpecificAppSettingsIsCollected(t *testing.T) {
	// Act
	got := IsNewRelicConfigFile("C:/inetpub/myapp/appsettings.Production.json")

	// Assert
	if !got {
		t.Error("expected appsettings.Production.json to match the config collection patterns")
	}
}
