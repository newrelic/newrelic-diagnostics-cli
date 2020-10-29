package env

import (
	"testing"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Base/Env/InitSystem", func() {
	var p BaseEnvInitSystem //instance of our task struct to be used in tests

	Describe("Identifier()", func() {
		It("Should return identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Category: "Base", Subcategory: "Env", Name: "InitSystem"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return help text", func() {
			Expect(p.Explain()).To(Equal("Determine Linux init system"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return the empty list of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{}))
		})
	})

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when running on Windows", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				p.runtimeOs = "windows"
				upstream = map[string]tasks.Result{}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task does not apply to Windows"))
			})
		})

		Context("when running on Darwin", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				p.runtimeOs = "darwin"
				upstream = map[string]tasks.Result{}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task does not apply to Mac OS"))
			})
		})

		Context("when there is an error evaluating /sbin/init symlink", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}

				p.runtimeOs = "linux"
				p.evalSymlink = func(string) (string, error) {
					return "", errors.New("Could not resolve symlink!")
				}
			})

			It("should return an expected error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected error result summary", func() {
				Expect(result.Summary).To(Equal("Unable to read symbolic link for /sbin/init: Could not resolve symlink!"))
			})
		})


		Context("when /sbin/init symlink points to unrecognized init system path", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}

				p.runtimeOs = "linux"
				p.evalSymlink = func(string) (string, error) {
					return "/usr/gibson/garbage_file", nil
				}
			})

			It("should return an expected error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected error result summary", func() {
				Expect(result.Summary).To(Equal("Unable to parse init system from: /usr/gibson/garbage_file"))
			})
		})

		Context("when /sbin/init symlink points to a recognized init system path", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}

				p.runtimeOs = "linux"
				p.evalSymlink = func(string) (string, error) {
					return "/usr/gibson/systemd", nil
				}
			})

			It("should return an expected info result ", func() {
				expectedResult := tasks.Result{
					Status: tasks.Info,
					Summary: "Systemd detected",
					Payload: "Systemd",
				}
				Expect(result).To(Equal(expectedResult))
			})
		})
	})

})

func Test_parseInitSystem(t *testing.T) {

	tests := []struct {
		initPath string
		want string
	}{
		{initPath: "/sbin/init", want: "SysV"},
		{initPath: "/lib/systemd/systemd", want: "Systemd"},
		{initPath: "/usr/lib/systemd/systemd", want: "Systemd"},
		{initPath: "/test/location/systemd", want: "Systemd"},
		{initPath: "/lib/systemd/upstart", want: "Upstart"},
		{initPath: "/usr/lib/systemd/upstart", want: "Upstart"},
		{initPath: "upstart", want: "Upstart"},
		{initPath: "", want: ""},
		{initPath: "/lib/systemd/", want: ""},
		{initPath: "/usr/lib/systemd/fgdfg", want: ""},
		{initPath: "/test/location/systemd/", want: ""},		
	}
	for _, tt := range tests {
		t.Run("parseInitSystem() unit test", func(t *testing.T) {
			if got := parseInitSystem(tt.initPath); got != tt.want {
				t.Errorf("parseInitSystem('%v') = %v, want %v", tt.initPath, got, tt.want)
			}
		})
	}
}
