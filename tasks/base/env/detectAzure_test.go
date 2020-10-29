package env

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Base/Env/DetectAzure", func() {
	var p BaseEnvDetectAzure //instance of our task struct to be used in tests

	//Tests go here!
	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Env",
				Name:        "DetectAzure",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explanation of task", func() {
			Expect(p.Explain()).To(Equal("Detect if running in Azure environment"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return list of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Base/Env/CollectEnvVars"}))
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

		Context("When collectEnvVars task returns tasks.Warning", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Warning,
						Summary: "Unable to gather any Environment Variables from the current shell.",
					},
				}
			})

			It("Should return a status of None and a summary.", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("Unable to gather environment variables, this task did not run"))
			})
		})

		Context("When we did not detected an Azure environment variable", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Summary: "Gathered Environment variables of current shell.",
						Payload: map[string]string{
							"COR_ENABLE_PROFILING": "1",
						},
					},
				}
			})

			It("Should return task.Status of None", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("Should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Detected that this is not an Azure environment."))
			})
		})

		Context("When we detect the Azure environment variable", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Summary: "Gathered Environment variables of current shell.",
						Payload: map[string]string{
							"COR_ENABLE_PROFILING": "1",
							"WEBSITE_SITE_NAME":    "My Application Name",
						},
					},
				}
			})

			It("Should return task.Status of Info", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("Should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Identified this as an Azure environment."))
			})
		})
	})

})
