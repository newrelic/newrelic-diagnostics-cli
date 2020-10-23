package template

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Example/Template/MinimalTask", func() {

	var p ExampleTemplateMinimalTask

	Context("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Example",
				Subcategory: "Template",
				Name:        "MinimalTask",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Context("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("This task doesn't do anything."))
		})
	})

	Context("Dependencies()", func() {
		It("Should return an empty slice of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{}))
		})
	})

	Context("Execute()", func() {
		It("Should return expected result", func() {
			executeOptions := tasks.Options{}
			executeUpstream := map[string]tasks.Result{}

			expectedResult := tasks.Result{
				Status:  tasks.None,
				Summary: "I succeeded in doing nothing.",
			}

			Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
		})
	})
})
