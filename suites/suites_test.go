package suites

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSuites(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "suites test suite")
}

var _ = Describe("FindTasksBySuites()", func() {
	var (
		//inputs
		suites []Suite
		sm     *SuiteManager
		//output
		expectedTasks []string
	)
	Context("when given suites", func() {
		BeforeEach(func() {
			suites = []Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Diagnose Java Agent installation",
					Tasks: []string{
						"Java/*",
					},
				},
				{
					Identifier:  "infra",
					DisplayName: "Infrastructure Agent",
					Description: "Diagnose Infrastructure Agent installation",
					Tasks: []string{
						"Infra/*",
					},
				},
			}
			sm = NewSuiteManager(suiteDefinitions)
			expectedTasks = []string{
				"Java/*",
				"Infra/*",
			}

		})
		It("Should return the collected task identifier patterns associated with them", func() {
			tasks := sm.FindTasksBySuites(suites)

			Expect(tasks).Should(ConsistOf(expectedTasks))
		})
	})
})
