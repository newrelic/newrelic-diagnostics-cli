package agent

import (
	"errors"
	"strings"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mockPathFinder() (string, error) {
	return "mockTestPath", nil
}

func mockPathFinderError() (string, error) {
	return "", errors.New("error getting current working directory")
}

func mockCmdExec(name string, arg ...string) ([]byte, error) {
	return []byte("5.3.10\n"), nil
}

func mockCmdExecExtraInfo(name string, arg ...string) ([]byte, error) {
	return []byte("Jul 25, 2019 14:33:31 -0700 [65178 1] com.newrelic INFO: Agent is using Logback\n5.3.10\n"), nil
}

func mockCmdExecJavaUnsupported(name string, arg ...string) ([]byte, error) {
	lines := []string{
		"Jul 25, 2019 14:33:28 -0700 [65171 1] com.newrelic INFO: Agent is using Logback",
		"----------",
		"Java version is: 1.8.0_212.  This version of the New Relic Agent does not support Java SE 8 (1.8).",
		"----------",
	}
	return []byte(strings.Join(lines, "\n")), nil
}

func mockCmdExecFailed(name string, arg ...string) ([]byte, error) {
	return []byte(""), errors.New("failed to execute command! :(")
}

func mockFindFiles(patterns []string, paths []string) []string {
	return []string{"mock path"}
}

func TestJavaAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Java/Agent/Version test suite")
}

var _ = Describe("Java/Agent/Version", func() {
	var p JavaAgentVersion //instance of our task struct to be used in tests

	//Tests go here!

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Java",
				Subcategory: "Agent",
				Name:        "Version",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})
	Describe("Explain()", func() {
		It("Should return correct explanation", func() {
			expectedExplain := "Determine New Relic Java agent version"
			Expect(p.Explain()).To(Equal(expectedExplain))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return correct dependency", func() {
			expectedDependencies := []string{
				"Java/Config/Agent",
			}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
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
		Context("When upstream dependency does not return tasks.Success", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})
			It("Should return tasks.None if agent version is empty", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
		})
		Context("When upstream dependency does return tasks.Success", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p = JavaAgentVersion{
					wdGetter:     mockPathFinder,
					cmdExec:      mockCmdExec,
					findTheFiles: mockFindFiles,
				}
			})
			It("Should return tasks.Info if agent version is not empty", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
		})
		Context("When upstream is successful but agent version is blank", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p = JavaAgentVersion{
					wdGetter:     mockPathFinder,
					cmdExec:      mockCmdExecFailed,
					findTheFiles: mockFindFiles,
				}
			})
			It("Should return tasks.Failure cmdExec Fails", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
	})
	Describe("findJavaJar()", func() {
		Context("When no error from Getwd", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					wdGetter: mockPathFinder,
				}
			})
			It("Should return correct paths to newrelic jar", func() {
				expectedResults := []string{
					"mockTestPath",
					"mockTestPath/newrelic",
				}
				Expect(p.findJavaJar()).To(Equal(expectedResults))
			})
		})
		Context("When error is received from Getwd", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					wdGetter: mockPathFinderError,
				}
			})
			It("Should return correct paths to newrelic jar", func() {
				var expectedResults []string
				Expect(p.findJavaJar()).To(Equal(expectedResults))
			})
		})
	})
	Describe("getAgentVersion()", func() {
		var (
			jarLocation string
		)
		JustBeforeEach(func() {
			jarLocation = "mockJarLocationPath"
		})
		Context("When CmdExec is nil for error", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					cmdExec: mockCmdExec,
				}
			})
			It("Should return agent version", func() {
				expectedResults := "5.3.10"
				jarLocationPath := jarLocation
				Expect(p.getAgentVersion(jarLocationPath)).To(Equal(expectedResults))
			})
		})
		Context("When CmdExec returns additional information", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					cmdExec: mockCmdExecExtraInfo,
				}
			})
			It("Should return agent version", func() {
				expectedResults := "5.3.10"
				jarLocationPath := jarLocation
				Expect(p.getAgentVersion(jarLocationPath)).To(Equal(expectedResults))
			})
		})
		Context("When CmdExec returns information about an unsupported JVM", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					cmdExec: mockCmdExecJavaUnsupported,
				}
			})
			It("Should return an empty string and an Error", func() {
				expectedResults := ""
				jarLocationPath := jarLocation
				version, err := p.getAgentVersion(jarLocationPath)
				Expect(version).To(Equal(expectedResults))
				Expect(err).To(Not(BeNil()))
			})
		})
		Context("When Error is not nil from CmdExec", func() {
			BeforeEach(func() {
				p = JavaAgentVersion{
					cmdExec: mockCmdExecFailed,
				}
			})
			It("Should return an empty string and an Error", func() {
				expectedResults := ""
				jarLocationPath := jarLocation
				version, err := p.getAgentVersion(jarLocationPath)
				Expect(version).To(Equal(expectedResults))
				Expect(err.Error()).To(Equal("failed to execute command! :("))
			})
		})
	})
})
