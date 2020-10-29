package env

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func mockFailure(string, ...string) ([]byte, error) {
	return []byte{}, errors.New("an error message")
}
func mockSuccess(string, ...string) ([]byte, error) {
	return []byte("a long list of Node.js modules"), nil
}

var _ = Describe("Node/Env/Dependencies", func() {

	var p NodeEnvDependencies

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "Dependencies",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})
	Describe("Explain()", func() {
		It("Should return an explanation of the task", func() {
			expectedExplanation := "Collect Nodejs application dependencies"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return an slice required to run this task", func() {
			expectedDependencies := []string{"Node/Env/NpmVersion", "Node/Config/Agent"}
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
		Context("When node agent is not installed", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})
			It("Should return a None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return a None result summary", func() {
				Expect(result.Summary).To(Equal("Node agent not detected. This task did not run"))
			})
		})

		Context("When NPM is not installed", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.None,
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})
			It("Should return a None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return a None result summary", func() {
				Expect(result.Summary).To(Equal("NPM is not installed. This task did not run"))
			})
		})
		Context("When getModulesList gets an error", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte{}, errors.New("an error message")
				}
			})
			It("Should return a Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return a Failure result summary", func() {
				Expect(result.Summary).To(Equal("an error message: npm throwed an error while running the command npm ls --depth=0 --parseable=true --long=true. Please verify that NR Diagnostics is running in your Node application directory. Possible causes for npm errors: https://docs.npmjs.com/common-errors"))
			})
		})
		Context("When NodeModulesVersions length is zero", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				//Adjusted the real output to make the regex fail
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte(`/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/babel-jest@23.6.0undefined\n/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/bcryptjs@2.4.3undefined\n`), nil
				}

			})
			It("Should return a Error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return a Error result summary", func() {
				Expect(result.Summary).To(Equal("We failed to parse the output of npm ls, but have included it in nrdiag-output.zip"))
			})
		})
		Context("When getModulesList returns a succesful output that we pass to tasks.FileCopyEnvelope", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/NpmVersion": tasks.Result{
						Status: tasks.Info,
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte(`/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/babel-jest:babel-jest@23.6.0:undefined\n/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/bcryptjs:bcryptjs@2.4.3:undefined\n`), nil
				}
			})
			It("Should return a Info result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("Should return a Succesful result summary", func() {
				Expect(result.Summary).To(Equal("We have successfully retrieved a list of dependencies from your node_modules folder"))
			})
			It("Should return a Succesful result FilesToCopy", func() {
				Expect(len(result.FilesToCopy)).To(Equal(1))
				Expect(result.FilesToCopy[0].Path).To(Equal("npm_ls_output.txt"))
			})
		})
	})

	Describe("streamSource()", func() {
		var (
			//input
			stream      chan string
			modulesList string
			//output
			// streamOut chan string
		)
		Context("When given a valid zero length channel and modulesList multiline string", func() {
			BeforeEach(func() {
				stream = make(chan string)
				//to replicate the real ouput of getModulesListStr(a multiline string), init a slice of strings and join them with newline
				npmLsLines := []string{
					"/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/babel-jest:babel-jest@23.6.0:undefined",
					"/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/bcryptjs:bcryptjs@2.4.3:undefined",
					"/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/express:express@4.16.4:undefined",
				}
				modulesList = strings.Join(npmLsLines, "\n")

			})
			JustBeforeEach(func() {
				go streamSource(modulesList, stream)
			})
			It("should return a channel that receives one line at a time when reading from a scanner", func() {
				Expect(<-stream).To(Equal("/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/babel-jest:babel-jest@23.6.0:undefined\n"))
				Expect(<-stream).To(Equal("/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/bcryptjs:bcryptjs@2.4.3:undefined\n"))
				Expect(<-stream).To(Equal("/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/express:express@4.16.4:undefined\n"))
			})
			It("should close the channel after its output is exhausted", func() {
				for range stream {
					//it would range over channel to consume it
				}
				_, ok := <-stream //should return false if channel it's been closed
				Expect(ok).To(Equal(false))
			})

		})
	})
	Describe("getNodeModulesVersions()", func() {
		var (
			//input
			modulesList string
			//output
			modulesVersions []NodeModuleVersion
		)
		dependencyInfo := []NodeModuleVersion{
			NodeModuleVersion{
				Module:  "express",
				Version: "4.16.4",
			},
			NodeModuleVersion{
				Module:  "mongoose",
				Version: "5.4.0",
			},
			NodeModuleVersion{
				Module:  "babel-jest",
				Version: "23.6.0",
			},
		}
		JustBeforeEach(func() {
			modulesVersions = p.getNodeModulesVersions(modulesList)
		})

		Context("When given a multiline string(modulesList)", func() {
			BeforeEach(func() {
				npmLsLines := []string{
					"/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/express:express@4.16.4:undefined", "/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/mongoose:mongoose@5.4.0:undefined", "/Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/babel-jest:babel-jest@23.6.0:undefined",
				}
				modulesList = strings.Join(npmLsLines, "\n")
			})
			It("should return an slice of modulesVersions struct", func() {
				Expect(modulesVersions).To(Equal(dependencyInfo))
			})
		})
	})

})
