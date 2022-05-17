package env

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node/Env/Version", func() {
	var p NodeEnvVersion //instance of our task struct that tests will be using

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "Version",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})
	Describe("Explain()", func() {
		It("Should return an explanation of the task", func() {
			expectedExplanation := "Determine Nodejs version"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return slice of expected dependencies", func() {
			expectedArray := []string{"Node/Config/Agent"}
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

		Context("When node agent is not detected", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})
			It("Should return a none status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Node agent config file not detected. This task did not run"))
			})
		})

		Context("When node -v returns an error", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {

					return []byte(""), errors.New("an error message")
				}
			})
			It("Should return a error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return a bad summary", func() {
				Expect(result.Summary).To(Equal("Unable to execute command:$ node -v. Error: an error message"))
			})
		})

		Context("When node -v returns a non version string", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {

					return []byte("node is the best!"), nil
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return a bad summary", func() {
				Expect(result.Summary).To(Equal("Unexpected output from received node -v: node is the best!"))
			})
		})
		Context("With error parsing version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {

					return []byte("v.10.1.9"), nil
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("An issue occur while parsing node version: unable to convert  to an integer"))
			})
		})
		Context("When node -v returns the expected string", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {

					return []byte("v10.7.0"), nil
				}
			})
			It("Should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("Should return a succesful summary", func() {
				Expect(result.Summary).To(Equal("v10.7.0"))
			})

		})
	})
})
