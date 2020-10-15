package minion

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

func TestSyntheticsMinionCollectLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Synthetics/Minion/* test suite")
}

var _ = Describe("Synthetics/Minion/CollectLogs", func() {

	var p SyntheticsMinionCollectLogs

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Synthetics",
				Subcategory: "Minion",
				Name:        "CollectLogs",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Collect logs of found Containerized Private Minions"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return Synthetics/Minion/DetectCPM as a dependency", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Synthetics/Minion/DetectCPM"}))
		})
	})

	Describe("Execute()", func() {

		//Fixtures
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		fixtureContainerIds := []string{
			"b3e2f06abf13fddb3ee805d7cdd4d3b5160e014627b460f50c181267f586ba74",
			"db9438b731b265313bc9a9ddd1f6ef641dad01d2a359cf16c396e908c535c854",
			"c4da144535976bf9b7853f97276c0d36586db1216cef84d705087b5dddfd6e7c",
			"7f19f8818654cb489d17184697b649adc541b477d6020de0221c5be602627a96",
		}

		//Successful DockerContainer structs from DetectCPMResult. We only need to refer to container Id in this task.
		cpmContainers := []tasks.DockerContainer{}

		for _, containerId := range fixtureContainerIds {
			var container tasks.DockerContainer
			container.Id = containerId
			cpmContainers = append(cpmContainers, container)
		}

		successfulDetectCPMResult := tasks.Result{
			Status:  tasks.Success,
			Summary: "Found Containerized Private Minions",
			Payload: cpmContainers,
		}

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("If upstream DetectCPM finds no running CPM containers", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Synthetics/Minion/DetectCPM": tasks.Result{
						Status:  tasks.None,
						Summary: "No Containerized Private Minions found",
						Payload: nil,
					},
				}
			})

			It("Should return a task status of none with summary that there are no CPMs to collect logs from", func() {
				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "No CPMs detected to collect logs from",
					Payload: nil,
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

		Context("If error there is an error with docker log command", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Synthetics/Minion/DetectCPM": successfulDetectCPMResult,
				}
				p.executeCommand = func(limit int64, cmd string, args ...string) (*bufio.Scanner, error) {
					argContainerId := args[1]
					containerIdToError := "b3e2f06abf13fddb3ee805d7cdd4d3b5160e014627b460f50c181267f586ba74"

					if argContainerId == containerIdToError {
						return nil, errors.New("Error reading output")
					}
					logBytes := fmt.Sprintf("Logs from %s", argContainerId)
					return bufio.NewScanner(strings.NewReader(logBytes)), nil

				}
			})

			It("Should return a task status of error with error message in summary", func() {
				expectedResult := tasks.Result{
					Status:  tasks.Error,
					Summary: "Error collecting logs from containers: Error reading output\n",
					Payload: nil,
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

	})
})
