package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node/Env/NpmPackage", func() {

	var p NodeEnvNpmPackage

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "NpmPackage",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	// the explanation of the task should match this
	Describe("Explain()", func() {
		It("Should return an explanation of the task", func() {
			expectedExplanation := "Collect package.json and package-lock.json if they exist"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	// depends on NPM being installed, but not necessarily the Node Agent
	Describe("Dependencies()", func() {
		It("Should test if NPM is installed or not", func() {
			expectedDependencies := []string{"Node/Config/Agent", "Node/Env/NpmVersion"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})
	Describe("Execute()", func() {
		var (
			//inputs
			options  tasks.Options
			upstream map[string]tasks.Result
			//output
			result tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When NPM is not installed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("Should return a None result status and summary", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("NPM is not installed. This task did not run"))
			})
		})
		Context("When no files found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.Getwd = func() (string, error) {
					return "customerDir/", nil
				}
				p.fileFinder = func(patterns []string, paths []string) []string {
					return []string{}
				}
			})

			It("Should return a failure result status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("The package.json and package-lock.json files were not found where the " + tasks.ThisProgramFullName + " was executed. Please ensure the " + tasks.ThisProgramFullName + " executable is within your application's directory alongside your package.json file"))
			})
		})

		Context("When 1 file is found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.Getwd = func() (string, error) {
					return "myDir/", nil
				}
				p.fileFinder = func(patterns []string, paths []string) []string {
					return []string{"myDir/package.json"}
				}
			})

			It("Should return a success result status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("We have succesfully retrieved the following file(s):\npackage.json"))
			})
		})

		Context("When 2 files are found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.Getwd = func() (string, error) {
					return "myDir/", nil
				}
				p.fileFinder = func(patterns []string, paths []string) []string {
					return []string{"myDir/package.json", "myDir/package-lock.json"}
				}
			})

			It("Should return a success result status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("We have succesfully retrieved the following file(s):\npackage.json\npackage-lock.json"))
			})
		})
		Context("When 3 files are found, but one is in node_modules", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
				}
				p.Getwd = func() (string, error) {
					return "myDir/", nil
				}
				p.fileFinder = func(patterns []string, paths []string) []string {
					return []string{"myDir/package.json", "myDir/package-lock.json", "myDir/node_modules/package.json"}
				}
			})

			It("Should return a success result status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("We have succesfully retrieved the following file(s):\npackage.json\npackage-lock.json"))
			})

			It("Should should ignore files found in node_modules", func() {
				expectedPayload := []PackageJsonElement{
					{
						FileName: "package.json",
						FilePath: "myDir/",
					},
					{
						FileName: "package-lock.json",
						FilePath: "myDir/",
					}}
				expectedFilesToCopy := []tasks.FileCopyEnvelope{
					{
						Path: "myDir/package.json",
					},
					{
						Path: "myDir/package-lock.json",
					},
				}
				Expect(result.Payload).To(Equal(expectedPayload))
				Expect(result.FilesToCopy).To(Equal(expectedFilesToCopy))
			})
		})
	})

})
