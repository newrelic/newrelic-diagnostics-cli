package containers

import (
	"testing"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os/exec"
)

func TestBaseDetectDocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Containers/DetectDocker test suite")
}

var _ = Describe("Base/Containers/DetectDocker", func() {
	var p BaseContainersDetectDocker

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Containers",
				Name:        "DetectDocker",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Detect Docker Daemon"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return a slice ", func() {
			expectedDependencies := []string{}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		dockerInfoOutput, err := ioutil.ReadFile("./fixtures/dockerInfoOutput")
		//dockerInfoOutputString := string(dockerInfoOutput);

		if err != nil {
			log.Info("Error with reading fixture: " + err.Error())
		}

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("If Docker is not running", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.executeCommand = func(name string, arg ...string) ([]byte, error) {

					return []byte("Banana could not be detected"), &exec.ExitError{}
				}
			})
			It("Should return a task status of none", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return a summary of the error", func() {
				Expect(result.Summary).To(Equal("Error retrieving Docker Daemon info: <nil> - Banana could not be detected"))
			})

		})

		Context("If there is no error with running docker info and JSON is invalid but we dont get ServerVersion", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.executeCommand = func(name string, arg ...string) ([]byte, error) {
					return []byte(`{"Foo":"Bar"}`), nil
				}
			})
			It("Should return an none task status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return a summary with the error", func() {
				Expect(result.Summary).To(Equal("Can't determine if Docker Daemon is running on host: unexpected output"))
			})

		})

		Context("If Docker is running but output can't be parsed from json", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.executeCommand = func(name string, arg ...string) ([]byte, error) {

					return []byte("Banana could not be detected"), nil
				}
			})
			It("Should return a task status of none", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return a summary of the error", func() {
				Expect(result.Summary).To(Equal("Error parsing JSON docker info invalid character 'B' looking for beginning of value"))
			})

		})

		Context("If DockerDaemon is determined to be running", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.executeCommand = func(name string, arg ...string) ([]byte, error) {

					return dockerInfoOutput, nil
				}
			})
			It("Should return a task status of info", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("Should return a summary that reports Docker Daemon is running", func() {
				Expect(result.Summary).To(Equal("Docker Daemon is Running"))
			})
			It("Should have a payload of info command output", func() {

				expectedPayload := tasks.DockerInfo{
					Driver:        "overlay2",
					ServerVersion: "19.03.4",
					MemTotal:      2095968256,
					NCPU:          4,
				}
				Expect(result.Payload).To(Equal(expectedPayload))
			})
		})
	})
})
