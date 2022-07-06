package agent

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	BrowserFixtures "github.com/newrelic/newrelic-diagnostics-cli/tasks/fixtures/Browser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browser/Agent/GetSource", func() {
	var p BrowserAgentGetSource

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Browser",
				Subcategory: "Agent",
				Name:        "GetSource",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Determine New Relic Browser agent loader script from provided URL"
			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
			expectedDependencies := []string{}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("getLoaderScript()", func() {
		var (
			data                    string
			goodScripts, badScripts []string
		)

		JustBeforeEach(func() {
			goodScripts, badScripts = getLoaderScript(data)
		})

		Context("when source loader of browser agent is not in the head tag", func() {

			BeforeEach(func() {
				data = BrowserFixtures.HTMLWithBadLoader
			})

			It("should return only badScripts ", func() {
				Expect(len(goodScripts)).To(Equal(0))
				Expect(len(badScripts)).To(Equal(1))
			})
		})
	})

})
