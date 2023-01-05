package agent

// Tests for Infra/Config/IntegrationsCollect

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Infra/Agent/Version", func() {
	var p InfraAgentVersion

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Agent",
				Name:        "Version",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExaplanation := "Determine version of New Relic Infrastructure agent"
			Expect(p.Explain()).To(Equal(expectedExaplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Infra/Config/Agent", "Base/Env/CollectEnvVars"}
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

		Context("upstream dependency for retrieving environment variables failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.None,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Upstream dependency failed"))
			})
		})

		Context("upstream dependency for retrieving environment variables returns unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: "I should be a map[string]string{}, but I'm a string",
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("upstream dependency for detecting presence of infrastructure agent failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Infrastructure Agent not detected on system"))
			})
		})

		Context("Infrastructure agent is present and returns version in the expected format", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "darwin"
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"published_at": "2022-11-17T10:50:43Z"}`))),
					}, nil
				}
				p.now = func() time.Time { return time.Date(2022, time.Month(11), 21, 1, 10, 30, 0, time.UTC) }
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("New Relic Infrastructure Agent version: 1.14.0"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("1.14.0"))
			})
		})

		Context("Infrastructure agent is present and with unsupported version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "darwin"
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"published_at": "2020-11-17T10:50:43Z"}`))),
					}, nil
				}
				p.now = func() time.Time { return time.Date(2022, time.Month(11), 21, 1, 10, 30, 0, time.UTC) }
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("New Relic Infrastructure Agent version: 1.13.0"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(errUnsupportedVersion.Error()))
			})
		})

		Context("Infrastructure agent is present on windows but ProgramFiles env variable is not set", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "windows"
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("New Relic Infrastructure Agent version: 1.13.0"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to determine New Relic Infrastructure binary path: environment variable not set: ProgramFiles"))
			})
		})

		Context("Infrastructure agent is present but -version returns unparseable result", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "linux"
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("newrelic-infra: Permission denied"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to parse New Relic Infrastructure Agent version from: newrelic-infra: Permission denied"))
			})

			Context("Infrastructure agent is present but version check returns an error", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Env/CollectEnvVars": tasks.Result{
							Status:  tasks.Info,
							Payload: map[string]string{},
						},
						"Infra/Config/Agent": tasks.Result{
							Status: tasks.Success,
						},
					}
					p.runtimeOS = "linux"
					p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
						return []byte(""), errors.New("fromlet was defrobozticated")
					}

				})

				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("New Relic Infrastructure Agent version could not be determined because Diagnostics CLI encountered this issue when running the command 'newrelic-infra -version': fromlet was defrobozticated"))
				})
			})
			Context("Infrastructure agent is present but version check returns unparseable version string", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Env/CollectEnvVars": tasks.Result{
							Status:  tasks.Info,
							Payload: map[string]string{},
						},
						"Infra/Config/Agent": tasks.Result{
							Status: tasks.Success,
						},
					}
					p.runtimeOS = "linux"
					p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
						return []byte("New Relic Infrastructure Agent version: .19.1."), nil
					}

				})

				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Error))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Unable to parse New Relic Infrastructure Agent version from: New Relic Infrastructure Agent version: .19.1."))
				})
			})
		})

		Describe("getBinaryPath()", func() {

			var (
				envVars map[string]string
				result  string
				err     error
			)

			JustBeforeEach(func() {
				result, err = p.getBinaryPath(envVars)
			})
			Context("On Windows systems", func() {
				BeforeEach(func() {
					envVars = map[string]string{
						"ProgramFiles": `C:\Program Files`,
					}
					p.runtimeOS = "windows"
				})

				It("should reference the ProgramFiles ENV var when constructing the path", func() {
					Expect(result).To(Equal(`C:\Program Files\New Relic\newrelic-infra\newrelic-infra.exe`))
					Expect(err).To(BeNil())
				})
			})

			Context("On non-windows systems", func() {
				BeforeEach(func() {
					p.runtimeOS = "darwin"
				})

				It("should expect the newrelic-infra binary to exist in the PATH", func() {
					Expect(result).To(Equal(`newrelic-infra`))
				})
			})
		})

		Describe("validatePublishDate()", func() {

			var (
				version tasks.Ver
				err     error
			)

			JustBeforeEach(func() {
				err = p.validatePublishDate(version)
			})
			Context("Version prior to 1.12.0", func() {
				BeforeEach(func() {
					version = tasks.Ver{Major: 1, Minor: 11, Patch: 0, Build: 0}
				})

				It("should return unsupported version error", func() {
					Expect(err).To(Equal(errUnsupportedVersion))
				})
			})

			Context("Up-to-date version", func() {
				BeforeEach(func() {
					version = tasks.Ver{Major: 1, Minor: 33, Patch: 0, Build: 0}
					p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
						return &http.Response{
							Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"published_at": "2021-11-22T10:50:43Z"}`))),
						}, nil
					}
					p.now = func() time.Time { return time.Date(2022, time.Month(11), 21, 1, 10, 30, 0, time.UTC) }
				})

				It("should not return an error", func() {
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
