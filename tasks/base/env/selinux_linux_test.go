package env

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Base/Env/SELinux", func() {
	var p BaseEnvCheckSELinux

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Env",
				Name:        "SELinux",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Check for SELinux presence."

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
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

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When sestatus command will return an unexpected error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return []byte{}, errors.New("execution error")
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return in the summary the execution error", func() {
				Expect(result.Summary).To(Equal("Unable to execute command: sestatus Error: " + "execution error"))
			})
			
			It("should return a payload with the SEUnknown SEMode", func() {
				Expect(result.Payload).To(Equal(SEUnknown))
			})
		})
		Context("When sestatus command returns an expected, 'healthy', error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return []byte{}, errors.New("Command 'sestatus' not found")
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return a summary that reports SELinux is not installed", func() {
				Expect(result.Summary).To(Equal("SELinux does seem to be installed in this environment: Command 'sestatus' not found"))
			})
			It("should return an payload of SEMode NotInstalled", func() {
				Expect(result.Payload).To(Equal(NotInstalled))
			})
		})

		Context("When SELinux is set to enforcing mode", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					output :=
						`SELinux status:                 enabled
					SELinuxfs mount:                /sys/fs/selinux
					SELinux root directory:         /etc/selinux
					Loaded policy name:             targeted
					Current mode:                   enforcing
					Mode from config file:          enforcing
					Policy MLS status:              enabled
					Policy deny_unknown status:     allowed
					Max kernel policy version:      28`
					return []byte(output), nil
				}
			})

			It("should return an expected warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return a summary that SELinux is enabled", func() {
				Expect(result.Summary).To(Equal("We have detected that SELinux is enabled with enforcing mode in your environment. If you are having issues installing a New Relic product or you have no data reporting, temporarily disable SELinux, to verify this resolves the issue."))
			})

			It("should return an payload of SEMode Enforced", func() {
				Expect(result.Payload).To(Equal(Enforced))
			})
		})

		Context("When SELinux is disabled", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return []byte(`SELinux status:                 disabled`), nil
				}
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected success result summary", func() {
				Expect(result.Summary).To(Equal("Verified SELinux is not enforcing."))
			})

			It("should return an payload of SEMode NotEnforced", func() {
				Expect(result.Payload).To(Equal(NotEnforced))
			})
		})
	})
})
