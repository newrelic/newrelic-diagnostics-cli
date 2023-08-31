package main

import (
	"io/ioutil"
	"runtime"
	"sort"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dependency grapher test suite test suite")
}

func getFileContentString(path string) string {
	fileContentBytes, _ := ioutil.ReadFile(path)
	return string(fileContentBytes)
}

var _ = Describe("Dependency Grapher", func() {
	//Fixtures
	fileContent_badTaskFile := getFileContentString("fixtures/badTaskFile.go")
	fileContent_noDependencies := getFileContentString("fixtures/goodTaskFile_noDependencies.go")
	fileContent_twoDependencies := getFileContentString("fixtures/goodTaskFile_twoDependencies.go")

	Describe("isTaskFile()", func() {
		var filePath string
		var result bool

		JustBeforeEach(func() {
			result = isTaskFile(filePath)
		})

		Context("when given a file with requisite method signatures", func() {
			BeforeEach(func() {
				filePath = "fixtures/goodTaskFile_twoDependencies.go"
			})

			It("should return true", func() {
				Expect(result).To(Equal(true))
			})
		})

		Context("when given a file that doesn't exist", func() {
			BeforeEach(func() {
				filePath = "fixtures/i_dont_exist.go"
			})

			It("should return false", func() {
				Expect(result).To(Equal(false))
			})
		})

		Context("when given a file that doesn't contain requisite method signatures", func() {
			BeforeEach(func() {
				filePath = "fixtures/badTaskFile.go"
			})

			It("should return false", func() {
				Expect(result).To(Equal(false))
			})
		})
	})

	Describe("getGoFiles()", func() {
		var expath string
		var result []string
		var err error

		BeforeEach(func() {
			expath = "fixtures/"
		})
		JustBeforeEach(func() {
			result, err = getGoFiles(expath)
		})

		Context("when given folder containing multiple file types", func() {

			It("should only return .go files", func() {
				var expectedResult []string
				if runtime.GOOS == "windows" {
					expectedResult = []string{
						"fixtures\\badTaskFile.go",
						"fixtures\\goodTaskFile_noDependencies.go",
						"fixtures\\goodTaskFile_twoDependencies.go",
					}
				} else {
					expectedResult = []string{
						"fixtures/badTaskFile.go",
						"fixtures/goodTaskFile_noDependencies.go",
						"fixtures/goodTaskFile_twoDependencies.go",
					}
				}
				Expect(expectedResult).To(Equal(result))
			})

			It("should return a nil error", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("parseIdentifier()", func() {
		var fileContent string
		var result string

		JustBeforeEach(func() {
			result = parseIdentifier(fileContent)
		})

		Context("when given a string containing expected sub string", func() {
			BeforeEach(func() {
				fileContent = fileContent_twoDependencies
			})
			It("should return identifier", func() {
				Expect(result).To(Equal("Good/TaskFile/TwoDependencies"))
			})
		})

		Context("when given a string not containing expected sub string", func() {
			BeforeEach(func() {
				fileContent = fileContent_badTaskFile
			})
			It("should return empty string", func() {
				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("parseDependencies()", func() {
		var fileContent string
		var result []string

		JustBeforeEach(func() {
			result = parseDependencies(fileContent)
		})

		Context("when given valid file containing two dependencies", func() {
			BeforeEach(func() {
				fileContent = fileContent_twoDependencies
			})
			It("should return the two dependencies in a slice of strings", func() {
				expectedResult := []string{
					"I/Am/Dependency1",
					"I/Am/Dependency2",
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

		Context("when given valid file containing no dependencies", func() {
			BeforeEach(func() {
				fileContent = fileContent_noDependencies
			})
			It("should return an empty slice of strings", func() {
				expectedResult := []string{}
				Expect(result).To(Equal(expectedResult))
			})
		})
	})

	Describe("parseTaskInfo()", func() {
		var filePath string
		var result taskInfo
		var err error

		JustBeforeEach(func() {
			result, err = parseTaskInfo(filePath)
		})

		Context("when given valid task file", func() {
			BeforeEach(func() {
				filePath = "fixtures/goodTaskFile_twoDependencies.go"
			})

			It("should return taskInfo with expected identifier", func() {
				Expect(result.Identifier).To(Equal("Good/TaskFile/TwoDependencies"))
			})
			It("should return taskInfo with expected dependencies", func() {
				expectedDependencies := []string{
					"I/Am/Dependency1",
					"I/Am/Dependency2",
				}
				Expect(result.Dependencies).To(Equal(expectedDependencies))
			})

			It("should not return an error", func() {
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("parseTasks()", func() {
		var goFiles []string
		var result map[string]map[string]bool
		var err error

		JustBeforeEach(func() {
			result, err = parseTasks(goFiles)
		})

		Context("when given a list of .go files", func() {
			BeforeEach(func() {
				if runtime.GOOS == "windows" {
					goFiles = []string{
						"fixtures\\badTaskFile.go",
						"fixtures\\goodTaskFile_noDependencies.go",
						"fixtures\\goodTaskFile_twoDependencies.go",
					}
				} else {
					goFiles = []string{
						"fixtures/badTaskFile.go",
						"fixtures/goodTaskFile_noDependencies.go",
						"fixtures/goodTaskFile_twoDependencies.go",
					}
				}

			})

			It("should return map of only task file info", func() {
				expectedTwoDependencyMap := map[string]bool{
					"I/Am/Dependency1": true,
					"I/Am/Dependency2": true,
				}
				expectedNoDependencyMap := map[string]bool{}
				Expect(result["Good/TaskFile/TwoDependencies"]).To(Equal(expectedTwoDependencyMap))
				Expect(result["Good/TaskFile/NoDependencies"]).To(Equal(expectedNoDependencyMap))
			})

			It("should not return an error", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("taskMapToCSV()", func() {
		var taskMap map[string]map[string]bool
		var result []string

		JustBeforeEach(func() {
			result = taskMapToCSV(taskMap)
			sort.Strings(result)
		})

		Context("when given a valid task map", func() {
			BeforeEach(func() {
				taskMap = map[string]map[string]bool{
					"Good/TaskFile/TwoDependencies": {
						"I/Am/Dependency1": true,
						"I/Am/Dependency2": true,
					},
					"Good/TaskFile/NoDependencies": {},
					"Example/Task/IgnoreMe":        {},
				}

			})

			It("should output dependencies as comma separated slice of strings without example task", func() {
				var expectedResults = []string{
					"I/Am/Dependency1,Good/TaskFile/TwoDependencies",
					"I/Am/Dependency2,Good/TaskFile/TwoDependencies",
					",Good/TaskFile/NoDependencies",
				}
				sort.Strings(expectedResults)
				Expect(result).To(Equal(expectedResults))
			})

		})
	})

})
