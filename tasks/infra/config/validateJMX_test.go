package config

// Tests for Infra/Config/ValidateJMX

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("Infra/Config/ValidateJMX", func() {
	format.TruncatedDiff = false
	var p InfraConfigValidateJMX
	log.Debug("testing")

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Config",
				Name:        "ValidateJMX",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Validate New Relic Infrastructure JMX integration configuration file"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})

	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{
				"Infra/Config/IntegrationsMatch",
				"Java/Env/Version",
			}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	//Find config file in "integrations.d" foldernc

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("IntegrationsValidate was not successful", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("Should return a 'None' result status.", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
		})

		Context("upstream dependency task returned unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: "I am a string, I should be a MatchedIntegrationFiles{}",
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("no configurations were present in upstream payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status: tasks.Success,
						Payload: MatchedIntegrationFiles{
							IntegrationFilePairs: map[string]*IntegrationFilePair{},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No JMX Integration config or definition files were found. Task not executed."))
			})
		})

		Context("no JXM configurations were present in upstream payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status: tasks.Success,
						Payload: MatchedIntegrationFiles{
							Errors: []IntegrationMatchError{},
							IntegrationFilePairs: map[string]*IntegrationFilePair{
								"apache": &IntegrationFilePair{
									Configuration: config.ValidateElement{
										Config: config.ConfigElement{
											FileName: "apache-config.yml",
											FilePath: "/etc/newrelic-infra/integrations.d/",
										},
										ParsedResult: tasks.ValidateBlob{},
									},
									Definition: config.ValidateElement{
										Config: config.ConfigElement{
											FileName: "apache-definition.yml",
											FilePath: "/var/db/newrelic-infra/custom-integrations/",
										},
										ParsedResult: tasks.ValidateBlob{},
									},
								},
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No JMX Integration config or definition files were found. Task not executed."))
			})
		})
		Context("Invalid JXM configuration present in upstream payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/emptyJMX.yml", "fixtures/validateJMX/emptyJMX.yml"),
					},
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected failure result summary", func() {
				Expect(result.Summary).To(Equal("Unexpected results for jmx-config.yml: Invalid configuration found: collection_files not set"))
			})
			It("should return help URL", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection"))
			})
		})
		Context("Valid JXM configuration present in upstream payload containing auth", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-config.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}
				p.mCmdExecutor = func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error) {
					return []byte("success"), nil
				}

			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected successful result summary", func() {
				Expect(result.Summary).To(Equal("Successfully connected to configured JMX Integration config"))
			})

			It("should redact auth from payload JSON", func() {
				resultPayload := result.Payload.(JmxConfig)
				payloadJSON, _ := json.MarshalIndent(resultPayload, "", "	")
				expectedPayloadJSON := "{\n\t\"jmx_host\": \"jmx-host.localnet\",\n\t\"jmx_port\": \"9999\",\n\t\"jmx_user\": \"_REDACTED_\",\n\t\"jmx_pass\": \"_REDACTED_\",\n\t\"collection_files\": \"/etc/newrelic-infra/integrations.d/jvm-metrics.yml,/etc/newrelic-infra/integrations.d/tomcat-metrics.yml\",\n\t\"java_version\": \"Unable to find a Java path/version after running the command: java -version\",\n\t\"jmx_process_arguments\": [\n\t\t\"Unable to find a running JVM process that has JMX enabled or configured in its arguments\"\n\t]\n}"
				Expect(string(payloadJSON)).To(Equal(expectedPayloadJSON))
			})
		})
		Context("Valid partial JXM configuration present in upstream payload", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-partial.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}
				p.mCmdExecutor = func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error) {
					return []byte("success"), nil
				}
			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected success result summary", func() {
				Expect(result.Summary).To(Equal("Successfully connected to configured JMX Integration config"))
			})
		})
		Context("Invalid JXM configuration present with duplicate keys in upstream payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-duplicate-keys.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}

			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("should return help URL", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection"))
			})

			It("should return an expected Failure result summary", func() {
				Expect(result.Summary).To(Equal("Unexpected results for jmx-config.yml: Multiple key jmx_host found"))
			})
		})
		Context("Valid partial JXM configuration present in upstream payload but JMX server connection failed", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-partial.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}
				p.mCmdExecutor = func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error) {
					errorString := "Apr 25, 2019 9:47:20 PM org.newrelic.nrjmx.Application main\nSEVERE: Can't connect to JMX server: service:jmx:rmi:///jndi/rmi://localhost:999/jmxrmi"
					return []byte(errorString), errors.New("error connecting blarg")
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("should return help URL", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#troubleshoot"))
			})

			It("should return an expected Failure result summary", func() {
				var expectedSummary strings.Builder
				expectedSummary.WriteString("We tested the JMX integration connection to local JMXServer by running the command echo '*:*' | nrjmx -H localhost -P 8080 -v -d - and we found this error:\n")
				expectedSummary.WriteString("Apr 25, 2019 9:47:20 PM org.newrelic.nrjmx.Application main\n")
				expectedSummary.WriteString("SEVERE: Can't connect to JMX server: service:jmx:rmi:///jndi/rmi://localhost:999/jmxrmi")
				Expect(result.Summary).To(Equal(expectedSummary.String()))
			})
		})
		Context("Valid default params JXM configuration present in upstream payload", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-default-parms.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}
				p.mCmdExecutor = func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error) {
					return []byte("success"), nil
				}
			})

			It("should return an expected Warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("should return help URL", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection"))
			})

			It("should return an expected Warning result summary", func() {
				var expectedSummary strings.Builder
				expectedSummary.WriteString("Successfully connected to JMX server but no hostname or port defined in jmx-config.yml\n")
				expectedSummary.WriteString("We recommend configuring this instead of relying on default parameters")
				Expect(result.Summary).To(Equal(expectedSummary.String()))
			})
		})
		Context("Invalid JXM configuration present in upstream payload", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsMatch": tasks.Result{
						Status:  tasks.Success,
						Payload: matchedIntegrationFilesFromFiles("fixtures/validateJMX/jmx-no-parms.yml", "fixtures/validateJMX/jmx-definition.yml"),
					},
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("should return help URL", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection"))
			})

			It("should return an expected Failure result summary", func() {
				Expect(result.Summary).To(Equal("Unexpected results for jmx-config.yml: Invalid configuration found: collection_files not set"))
			})
		})

	})

	// wrap this all in non-windows since os specific functions are difficult to cross-test
	if runtime.GOOS != "windows" {
		Describe("tasks.MultiCmdExecutor", func() {
			Context("When testing real functions", func() {
				cmdWrapper1 := tasks.CmdWrapper{
					Cmd:  "echo",
					Args: []string{"11"},
				}
				cmdWrapper2 := tasks.CmdWrapper{
					Cmd: "base64",
				}
				output, _ := tasks.MultiCmdExecutor(cmdWrapper1, cmdWrapper2)
				It("Should return valid output", func() {
					Expect(string(output)).To(Equal("MTEK\n"))
				})
			})
			Context("When returning an error from cmd1", func() {
				cmdWrapper1 := tasks.CmdWrapper{
					Cmd: "nonsense garbage",
				}
				cmdWrapper2 := tasks.CmdWrapper{
					Cmd: "base64",
				}
				_, err := tasks.MultiCmdExecutor(cmdWrapper1, cmdWrapper2)
				It("Should return valid output", func() {
					Expect(err.Error()).To(ContainSubstring("not found"))
				})
			})

			Context("When returning an error from cmd2", func() {
				cmdWrapper1 := tasks.CmdWrapper{
					Cmd:  "echo",
					Args: []string{"11"},
				}
				cmdWrapper2 := tasks.CmdWrapper{
					Cmd: "I've got a lovely bunch of coconuts",
				}
				_, err := tasks.MultiCmdExecutor(cmdWrapper1, cmdWrapper2)
				It("Should return valid output", func() {
					Expect(err.Error()).To(ContainSubstring("not found"))
				})
			})

		})

	}

	Describe("checkJMXServer", func() {
		var p InfraConfigValidateJMX
		Context("When parsing map with all values", func() {
			jmxKeys := JmxConfig{
				Host:            "localhost",
				Port:            "9999",
				User:            "admin",
				Password:        "admin",
				CollectionFiles: "file1, file2",
			}
			p.mCmdExecutor = func(cmdWrapper1, cmdWrapper2 tasks.CmdWrapper) ([]byte, error) {
				return []byte("success"), nil
			}
			err := p.checkJMXServer(jmxKeys)
			It("Should return nil err", func() {
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("processJmxMap", func() {
		Context("When parsing map with all values", func() {
			jmxKeys := JmxConfig{
				Host:            "127.0.0.1",
				Port:            "9998",
				User:            "admin",
				Password:        "admin",
				CollectionFiles: "file1, file2",
			}

			output := buildNrjmxArgs(jmxKeys)
			It("Should return expected output", func() {
				expectedOutput := []string{"-hostname", "127.0.0.1", "-port", "9998", "-username", "admin", "-password", "admin"}
				Expect(output).To(Equal(expectedOutput))
			})
		})

		Context("When parsing map with partial values", func() {
			jmxKeys := JmxConfig{
				Host:            "localhost",
				Port:            "9999",
				CollectionFiles: "file1, file2",
			}

			output := buildNrjmxArgs(jmxKeys)
			It("Should return expected output", func() {
				expectedOutput := []string{"-hostname", "localhost", "-port", "9999"}
				Expect(output).To(Equal(expectedOutput))
			})
		})

	})

})

func matchedIntegrationFilesFromFiles(confFile, defFile string) MatchedIntegrationFiles {

	confFilehandle, _ := os.Open(confFile)
	defer confFilehandle.Close()
	parsedYmlConf, _ := config.ParseYaml(confFilehandle)
	defFilehandle, _ := os.Open(defFile)
	defer defFilehandle.Close()
	parsedYmlDef, _ := config.ParseYaml(defFilehandle)
	confDir, _ := filepath.Split(confFile)
	defDir, _ := filepath.Split(defFile)

	return MatchedIntegrationFiles{
		IntegrationFilePairs: map[string]*IntegrationFilePair{
			"jmx": &IntegrationFilePair{
				Configuration: config.ValidateElement{
					Config: config.ConfigElement{
						FileName: "jmx-config.yml",
						FilePath: confDir,
					},
					ParsedResult: parsedYmlConf,
				},
				Definition: config.ValidateElement{
					Config: config.ConfigElement{
						FileName: "jmx-definition.yml",
						FilePath: defDir,
					},
					ParsedResult: parsedYmlDef,
				},
			},
		},
	}
}
