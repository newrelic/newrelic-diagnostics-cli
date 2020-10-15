package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/suites"
)

func TestProcessTasks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "processTasks test suite")
}

var _ = Describe("processFlagsSuites()", func() {
	var (
		//inputs
		flagValue string
		osArgs    []string
		//output
		expectedSuites []suites.Suite
		expectedErr    error
	)

	BeforeEach(func() {
		osArgs = []string{}
	})

	Context("when suites arguments has a trailing space or empty argument", func() {
		BeforeEach(func() {
			flagValue = "java, "

			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
			}

		})
		It("Should ignore any arguments after the space", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return a nil error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(BeNil())
		})
	})
	Context("when an argument passed has a name that is not part of list of suites", func() {
		BeforeEach(func() {
			flagValue = "foo"
			expectedSuites = []suites.Suite{}
			expectedErr = fmt.Errorf("\n- Could not find the following task suites:\n\n  \"%s\"\n\nPlease use the `--help suites` to check for available suites and proper formatting \n", flagValue)
		})
		It("Should return an empty slice of suites", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return an error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(Equal(expectedErr))
		})
	})
	Context("when one argument matches a suite and another doesn't", func() {
		BeforeEach(func() {
			flagValue = "java,foo"
			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
			}
			expectedErr = fmt.Errorf("\n- Could not find the following task suites:\n\n  \"%s\"\n\nPlease use the `--help suites` to check for available suites and proper formatting \n", "foo")
		})
		It("Should return one matched suite", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return an error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(Equal(expectedErr))
		})
	})
	Context("when suites has out of place argument", func() {
		BeforeEach(func() {
			flagValue = "java"
			osArgs = []string{"./nrdiag", "--suites=java", "node"}

			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
			}
			expectedErr = fmt.Errorf("\n- You may have attempted to pass these arguments as suites:\n\n  \"%v\"\n\nPlease use the `--help suites` to check for available suites and proper formatting \n", "node")

		})
		It("Should return the matched suite in the correct position", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return an error about the out of place argument", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(Equal(expectedErr))
		})
	})
	Context("when suites has out of place argument that doesn't match to suite", func() {
		BeforeEach(func() {
			flagValue = "java"
			osArgs = []string{"--suites=java", "foo"}

			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
			}
		})
		It("Should return the matched suite in the correct position", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should not return an error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(BeNil())
		})
	})
	Context("when multiple arguments match to suites", func() {
		BeforeEach(func() {
			flagValue = "java,infra"
			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
				{
					Identifier:  "infra",
					DisplayName: "Infrastructure Agent",
					Description: "Infrastructure Agent installation",
					Tasks: []string{
						"Base/*",
						"Infra/*",
					},
				},
			}
			expectedErr = nil
		})
		It("Should return suites for each valid argument", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return nil for error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(BeNil())
		})
	})
	Context("if suite argument has mixed capitalization", func() {
		BeforeEach(func() {
			flagValue = "JaVa"
			expectedSuites = []suites.Suite{
				{
					Identifier:  "java",
					DisplayName: "Java Agent",
					Description: "Java Agent installation",
					Tasks: []string{
						"Base/*",
						"Java/*",
					},
				},
			}
		})
		It("Should still match to suite", func() {
			suites, _ := processFlagsSuites(flagValue, osArgs)

			Expect(suites).To(Equal(expectedSuites))

		})
		It("Should return nil for error", func() {
			_, err := processFlagsSuites(flagValue, osArgs)

			Expect(err).To(BeNil())
		})
	})

})
