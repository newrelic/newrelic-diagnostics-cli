package env

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

var _ = Describe("Node/Env/VersionCompatibility", func() {
	var p NodeEnvVersionCompatibility

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "VersionCompatibility",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Check Nodejs version compatibility with New Relic Nodejs agent"

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
			expectedDependencies := []string{"Node/Env/Version"}

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

		Context("when Node.js version does not return an string", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: 4,
					},
				}
				p = NodeEnvVersionCompatibility{
					supportedNodeVersions: []string{"6.0+"},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: type assertion failure"))
			})
		})

		Context("when upstream dependency does not return successful status", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Error,
						Payload: "4",
					},
				}
				p = NodeEnvVersionCompatibility{
					supportedNodeVersions: []string{"6.0+"},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Node is not installed"))
			})
		})

		Context("when running Ver.IsCompatible returns an error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				parsedVer, _ := tasks.ParseVersion("6.0.1.2")
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: parsedVer,
					},
				}
				p = NodeEnvVersionCompatibility{
					supportedNodeVersions: []string{"I Am Not AVersionRequirement"},
				}

			})

			It("should return an expected error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected error result summary", func() {
				Expect(result.Summary).To(Equal("There was an issue when checking for Node.js Version compatibility: Unable to parse version: I Am Not AVersionRequirement"))
			})
		})

		Context("When running versionIsCompatible returns false", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				parsedVer, _ := tasks.ParseVersion("5.0.7")
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: parsedVer,
					},
				}
				p = NodeEnvVersionCompatibility{
					supportedNodeVersions: []string{"6.0+"},
				}
			})

			It("should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("should return an expected failure result summary", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 5.0.7.0, is not compatible with New Relic's Node.js agent"))
			})
		})
		Context("When running versionIsCompatible returns true", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				parsedVer, _ := tasks.ParseVersion("10.0.7")
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: parsedVer,
					},
				}
			})
			p = NodeEnvVersionCompatibility{
				supportedNodeVersions: []string{"6.0+"},
			}

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("should return an expected succes result summary", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 10.0.7.0, is compatible with New Relic's Node.js agent"))
			})
		})
	})

})
