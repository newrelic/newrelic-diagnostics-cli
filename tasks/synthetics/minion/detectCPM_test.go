package minion

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Synthetics/Minion/DetectCPM", func() {

	var p SyntheticsMinionDetectCPM

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Synthetics",
				Subcategory: "Minion",
				Name:        "DetectCPM",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Collect docker inspect of New Relic Synthetics containerized private minions"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an empty slice of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Base/Containers/DetectDocker"}))
		})
	})

	Describe("Execute()", func() {

		//Fixtures
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		successfulDetectDockerResult := tasks.Result{
			Status: tasks.Info,
			Payload: tasks.DockerInfo{
				ServerVersion: "18.09.0",
			},
		}

		fixtureContainerIds := []string{
			"b3e2f06abf13fddb3ee805d7cdd4d3b5160e014627b460f50c181267f586ba74",
			"db9438b731b265313bc9a9ddd1f6ef641dad01d2a359cf16c396e908c535c854",
			"c4da144535976bf9b7853f97276c0d36586db1216cef84d705087b5dddfd6e7c",
			"7f19f8818654cb489d17184697b649adc541b477d6020de0221c5be602627a96",
		}

		dockerMultiContainerInspectBytes, readErr := ioutil.ReadFile("./fixtures/multi-inspect-fixture.json")

		if readErr != nil {
			fmt.Printf("Error reading 'multi-inspect-fixture.json': %s", readErr.Error())
		}

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("If upstream DetectDocker reports docker daemon is not running", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": tasks.Result{
						Status: tasks.None,
						Payload: tasks.DockerInfo{
							ServerVersion: "",
						},
					},
				}
			})

			It("Should return a task status of none", func() {
				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "Docker Daemon not available to detect CPM",
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

		Context("If there are no CPMs running but there are four exited ones", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": successfulDetectDockerResult,
				}
				p.executeCommand = func(name string, args ...string) ([]byte, error) {
					argAction := args[0]

					if argAction == "ps" {
						return []byte("b3e2f06abf13\ndb9438b731b2\nc4da14453597\n7f19f8818654"), nil
					} else if argAction == "inspect" {
						return dockerMultiContainerInspectBytes, nil
					}
					return []byte{}, errors.New("unknown command")
				}
			})

			It("Should return a JSON payload of four inspected containers", func() {

				resultContainerIds := []string{}
				containerSlice, ok := result.Payload.([]tasks.DockerContainer)

				Expect(ok).To(Equal(true))

				//collect containerIds from payloads
				for _, container := range containerSlice {
					containerId := container.Id
					resultContainerIds = append(resultContainerIds, containerId)
				}
				Expect(resultContainerIds).To(ConsistOf(fixtureContainerIds))

			})

			It("Should redact un-whitelisted env variables from inspected containers", func() {
				containerSlice := result.Payload.([]tasks.DockerContainer)
				expectedRedactedVar := "PLEASE_REDACT=_REDACTED_"
				expectedRedactionCount := 4
				redactionCount := 0

				for _, container := range containerSlice {
					envVars := container.Config.Env

					for _, envVar := range envVars {
						if strings.ToUpper(envVar) == expectedRedactedVar {
							redactionCount++
						}
					}
				}

				Expect(redactionCount).To(Equal(expectedRedactionCount))
			})

		})
		Context("If there are no CPMs running or exited", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": successfulDetectDockerResult,
				}
				p.executeCommand = func(name string, args ...string) ([]byte, error) {
					argAction := args[0]

					if argAction == "ps" {
						return []byte{}, nil
					} else if argAction == "inspect" {
						return []byte{}, nil
					}
					return []byte{}, errors.New("unknown command")
				}
			})

			It("Should return a task result of none, and summary that No CPMs were found", func() {
				expectedResult := tasks.Result{}
				expectedResult.Payload = nil
				expectedResult.Status = tasks.None
				expectedResult.Summary = "No Containerized Private Minions found"

				Expect(result).To(Equal(expectedResult))

			})
		})
		Context("If there is an error with inspecting containers", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": successfulDetectDockerResult,
				}
				p.executeCommand = func(name string, args ...string) ([]byte, error) {
					argAction := args[0]

					if argAction == "ps" {
						return []byte("b3e2f06abf13\ndb9438b731b2\nc4da14453597\n7f19f8818654"), nil
					} else if argAction == "inspect" {
						return []byte("Foo!"), errors.New("inspect error")
					}
					return []byte{}, errors.New("unknown command")
				}
			})

			It("Should return a task result of Error, and summary that of error and output bytes", func() {
				expectedResult := tasks.Result{}
				expectedResult.Payload = nil
				expectedResult.Status = tasks.Error
				expectedResult.Summary = "inspect error Foo!"

				Expect(result).To(Equal(expectedResult))

			})
		})
		Context("If there is an error with redacting containers", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": successfulDetectDockerResult,
				}
				p.executeCommand = func(name string, args ...string) ([]byte, error) {
					argAction := args[0]

					if argAction == "ps" {
						return []byte("b3e2f06abf13\ndb9438b731b2\nc4da14453597\n7f19f8818654"), nil
					} else if argAction == "inspect" {
						return []byte(`[{"foo":"bar"}]`), nil
					}
					return []byte{}, errors.New("unknown command")
				}
			})

			It("Should return a task result of Error, and summary with error", func() {
				expectedResult := tasks.Result{}
				expectedResult.Payload = nil
				expectedResult.Status = tasks.Error
				expectedResult.Summary = "could not find Env variables in container inspect blob"

				Expect(result).To(Equal(expectedResult))

			})
		})
		Context("If there is an error getting containers ids", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Containers/DetectDocker": successfulDetectDockerResult,
				}
				p.executeCommand = func(name string, args ...string) ([]byte, error) {
					argAction := args[0]

					if argAction == "ps" {
						return []byte("Bad query format"), errors.New("invalid query")
					}
					return []byte{}, errors.New("unknown command")
				}
			})

			It("Should return a task result of Error, and summary with error", func() {
				expectedResult := tasks.Result{}
				expectedResult.Payload = nil
				expectedResult.Status = tasks.Error
				expectedResult.Summary = "error querying for container: invalid query: Bad query format"

				Expect(result).To(Equal(expectedResult))

			})
		})
	})
})
