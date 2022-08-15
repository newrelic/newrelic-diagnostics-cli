package env

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node/Env/NpmVersion", func() {
	var p NodeEnvNpmVersion

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "NpmVersion",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Determine NPM version"

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct array", func() {
			expectedArray := []string{"Node/Env/Version"}

			Expect(p.Dependencies()).To(Equal(expectedArray))
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

		Context("when Node isn't present", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Node.js was not detected. This task didn't run."))
			})
		})

		Context("Node.js is installed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.npmVersionGetter = func(tasks.CmdExecFunc) (string, error) {
					return "5.6.0", nil
				}
			})

			It("should return an expected info status result", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("should return an npm version", func() {
				Expect(result.Summary).To(Equal("5.6.0"))
			})
		})

		Context("when Node.js is installed and error is returned from npmVersionGetter", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.npmVersionGetter = func(tasks.CmdExecFunc) (string, error) {
					return "", errors.New("burrito error")
				}
			})

			It("should return an expected info status result", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an error message", func() {
				Expect(result.Summary).To(Equal("Unable to execute command: $ npm -v. Error: burrito error"))
			})
		})

		Context("when Node.js is installed and empty string is returned from npmVersionGetter", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.npmVersionGetter = func(tasks.CmdExecFunc) (string, error) {
					return "", nil
				}
			})

			It("should return an expected info status result", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an error message", func() {
				Expect(result.Summary).To(Equal("Unable to get npm version."))
			})
		})
	})

	Describe("getNpmVersion()", func() {

		var (
			result              string
			resultError         error
			mockCommandExecutor tasks.CmdExecFunc
		)

		JustBeforeEach(func() {
			result, resultError = getNpmVersion(mockCommandExecutor)
		})

		Context("when getNpmVersion() returns a valid string", func() {

			BeforeEach(func() {
				mockCommandExecutor = func(name string, arg ...string) ([]byte, error) {
					// do stuff
					return []byte("things"), nil
				}
			})

			It("Should return expected result", func() {
				Expect(result).To(Equal("things"))
			})

			It("Should return expected error", func() {
				Expect(resultError).To(BeNil())
			})
		})

		Context("when getNpmVersion() returns an error", func() {

			BeforeEach(func() {
				mockCommandExecutor = func(name string, arg ...string) ([]byte, error) {
					return []byte(""), errors.New("fancy Burrito error")
				}
			})

			It("Should return expected result", func() {
				Expect(result).To(Equal(""))
			})

			It("Should return expected error", func() {
				Expect(resultError.Error()).To(Equal("fancy Burrito error"))
			})
		})

		Context("getNpmVersion() trims extra space", func() {

			BeforeEach(func() {
				mockCommandExecutor = func(name string, arg ...string) ([]byte, error) {
					return []byte("5.6.0 "), nil
				}
			})

			It("Should return expected result", func() {
				Expect(result).To(Equal("5.6.0"))
			})

			It("Should return expected error", func() {
				Expect(resultError).To(BeNil())
			})
		})
	})
})
