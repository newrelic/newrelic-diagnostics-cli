// +build windows

package env

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestJavaEnvProcess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Java/Env/* test suite")
}

var _ = Describe("JavaEnvProcess", func() {
	var p JavaEnvProcess

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Java",
				Subcategory: "Env",
				Name:        "Process",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})
	Describe("Explain()", func() {
		It("Should return an explanation of the task", func() {
			expectedExplanation := "Collect Java process JVM command line options"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})
})
