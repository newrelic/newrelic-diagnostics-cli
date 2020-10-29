package main

import (
	"testing"

	tasks "github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func Test_parseOverrides(t *testing.T) {
	type args struct {
		overrides string
	}
	tests := []struct {
		name string
		args args
		want []override
	}{
		{"singleInput", args{"Base/Config/Validate.agentLanguage=Java"}, []override{{Identifier: tasks.IdentifierFromString("Base/Config/Validate"), key: "agentLanguage", value: "Java"}}},
		{"doubleInput", args{"Base/Config/Validate.agentLanguage=Java,Base/Config/Validate.agentLanguage=Go"}, []override{{Identifier: tasks.IdentifierFromString("Base/Config/Validate"), key: "agentLanguage", value: "Java"}, {Identifier: tasks.IdentifierFromString("Base/Config/Validate"), key: "agentLanguage", value: "Go"}}},
		{"invalidInput", args{"Base/Config/Valdate.agentLanguage=Java"}, []override{{Identifier: tasks.IdentifierFromString("Base/Config/Valdate"), key: "agentLanguage", value: "Java"}}},
		{"noInput", args{}, []override(nil)},
	}
	for _, tt := range tests {
		got := parseOverrides(tt.args.overrides)
		if !testOverridesEqual(got, tt.want) {
			t.Errorf("\nTest %v failed\nparseOverrides() returned: %#v;\nWe wanted                : %#v", tt.name, got, tt.want)
		}

	}
}

func testOverridesEqual(a, b []override) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Identifier != b[i].Identifier {
			return false
		}
		if a[i].key != b[i].key {
			return false
		}

		if a[i].value != b[i].value {
			return false
		}

	}

	return true
}
