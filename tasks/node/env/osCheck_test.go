package env

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	baseEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodeEnvOsCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node/Env/* test suite")
}

var _ = Describe("Node/Env/OsCheck", func() {
	var p NodeEnvOsCheck

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "OsCheck",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain", func() {
			Expect(p.Explain()).To(Equal("This task checks the OS for compatibility with the Node Agent."))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct dependencies", func() {
			expectedDependencies := []string{"Node/Config/Agent", "Base/Env/HostInfo"}
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

		Context("when Node agent not detected", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Node agent config file not detected"))
			})
		})

		Context("when OS was detected without a version and OS is not linux", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "OS only",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "windows",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Could not determine OS version."))
			})
		})

		Context("when OS was detected without a version and OS is not known OS", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "OS only",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "martha",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})

		Context("when OS was detected without a version and OS is linux", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "OS only",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "linux",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This OS is compatible with the Node Agent."))
			})
		})
	})

	Describe("checkOs()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when windows OS is detected as not compatible type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "windows",
							Platform:        "Microsoft Windows 10 Pro",
							PlatformFamily:  "Standalone Workstation",
							PlatformVersion: "10.0.16299 Build 16299",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})

		Context("when windows OS is detected as not compatible version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "windows",
							Platform:        "Microsoft Windows 10 Pro",
							PlatformFamily:  "server",
							PlatformVersion: "5.0.16299 Build 16299",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})

		Context("when windows OS is detected as compatible version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "windows",
							Platform:        "Microsoft Windows 10 Pro",
							PlatformFamily:  "server",
							PlatformVersion: "7.0.16299 Build 16299",
						},
					},
				}
			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				Expect(result.Summary).To(Equal("This OS is compatible with the Node Agent."))
			})
		})

		Context("when darwin OS is detected as compatible with major version above 10", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "darwin",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "11.7",
						},
					},
				}
			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				Expect(result.Summary).To(Equal("This OS is compatible with the Node Agent."))
			})
		})

		Context("when darwin OS is detected as compatible with major version equal to 10 and minor version above 6", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "darwin",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "10.7",
						},
					},
				}
			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				Expect(result.Summary).To(Equal("This OS is compatible with the Node Agent."))
			})
		})

		Context("when darwin OS is detected as not compatible with major version equal to 10 and minor version below 7", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "darwin",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "10.6",
						},
					},
				}
			})

			It("should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected failure result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})

		Context("when darwin OS is detected as not compatible with major version below 10", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "darwin",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "9.6",
						},
					},
				}
			})

			It("should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected failure result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})

		Context("when a non-compatible OS was detected", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					}, "Base/Env/HostInfo": tasks.Result{
						Status:  tasks.Info,
						Summary: "frank",
						Payload: baseEnv.HostInfo{
							Hostname:        "frank",
							OS:              "martha",
							Platform:        "",
							PlatformFamily:  "",
							PlatformVersion: "",
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This OS is not compatible with the Node Agent."))
			})
		})
	})

})
