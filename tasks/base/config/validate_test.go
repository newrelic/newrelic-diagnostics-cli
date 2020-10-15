package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func goldenFileName(testName string) string {
	return strings.Replace("fixtures/"+testName+".golden", " ", "_", -1)
}

type mockReader struct {
	Reader func([]byte) (int, error)
}

func (m mockReader) Read(b []byte) (int, error) {
	return m.Reader(b)
}

func TestBaseConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Config/* test suite")
}

var updateGoldenFiles bool

func init() {
	// can update with the command `ginkgo ./tasks/base/config/ -- --updateGoldenFiles true`
	flag.BoolVar(&updateGoldenFiles, "updateGoldenFiles", false, "updateGoldenFiles is used trigger a package level update of golden files")
}

var _ = Describe("Base/Config/Validate", func() {
	var p BaseConfigValidate
	Describe("Execute()", func() {
		Context("When upstream not successful", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status: tasks.Failure,
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return none result", func() {
				Expect(result.Summary).To(Equal("Config file collection was not successful, skipping validation step."))
				Expect(result.Status).To(Equal(tasks.None))
			})
		})
		Context("When upstream is successful but provides unexpected payload type", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status:  tasks.Success,
					Payload: ConfigElement{},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return error result", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: type assertion failure"))
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
		Context("When upstream is successful but provides no config files", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status:  tasks.Success,
					Payload: []ConfigElement{},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return none result", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: no configs"))
				Expect(result.Status).To(Equal(tasks.None))
			})
		})
		Context("When upstream returned unreadable file", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status: tasks.Success,
					Payload: ConfigElement{
						FileName: "blab",
						FilePath: "fixtures/",
					},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return error result", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: type assertion failure"))
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})

		Context("When parsing a standard XML .config file", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status: tasks.Success,
					Payload: []ConfigElement{
						ConfigElement{
							FileName: "validate_basexml.config",
							FilePath: "fixtures/",
						},
					},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return parsed xml", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)
				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result.Payload)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedPayload := readFile(goldenFile)
				Expect(result.Summary).To(Equal("Successfully parsed config file(s) - See json for full detail"))
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(fmt.Sprintf("#%v", result.Payload)).To(Equal(expectedPayload))
			})
		})
		Context("When parsing two files, one with errors", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status: tasks.Success,
					Payload: []ConfigElement{
						ConfigElement{
							FileName: "validate_basexml.config",
							FilePath: "fixtures/",
						},
						ConfigElement{
							FileName: "blah",
							FilePath: "fixtures/",
						},
					},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return parsed xml and error", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				//Windows and Linux have different filesystem error messages in this particular context.
				//We overwrite the error message here to be OS neutral so the same test works across both OSes
				normalizedPayload := result.Payload.([]ValidateElement)
				normalizedPayload[len(normalizedPayload)-1].Error = "normalized file error"

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result.Payload)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedPayload := readFile(goldenFile)
				expectedSummary := []string{
					"We were able to parse 1 of 2 configuration file(s).",
					"Errors parsing the following configuration file(s):",
					"fixtures/blah",
					"	Error: open fixtures/blah:",
				}
				Expect(result.Summary).To(ContainSubstring(strings.Join(expectedSummary, "\n")))
				Expect(result.Status).To(Equal(tasks.Warning))

				Expect(fmt.Sprintf("#%v", result.Payload)).To(Equal(expectedPayload))
			})
		})
		Context("When parsing two files, both failed", func() {
			upstream := map[string]tasks.Result{
				"Base/Config/Collect": tasks.Result{
					Status: tasks.Success,
					Payload: []ConfigElement{
						ConfigElement{
							FileName: "validate_badxml.config",
							FilePath: "fixtures/",
						},
						ConfigElement{
							FileName: "blah",
							FilePath: "fixtures/",
						},
					},
				},
			}
			options := tasks.Options{}
			result := p.Execute(options, upstream)
			It("Should return two errors", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result.Payload)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedPayload := readFile(goldenFile)
				expectedSummary := []string{
					"Errors parsing the following 2 configuration file(s):",
					"fixtures/validate_badxml.config",
					"	Error: xml.Decoder.Token() - XML syntax error on line 32: element <log> closed by </configuration>",
					"fixtures/blah",
					"	Error: open fixtures/blah:",
				}
				Expect(result.Summary).To(ContainSubstring(strings.Join(expectedSummary, "\n")))
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(fmt.Sprintf("#%v", result.Payload)).To(Equal(expectedPayload))
			})
		})

	})

	Describe("parseXML", func() {
		Context("When parsing standard XML", func() {
			file, _ := os.Open("fixtures/validate_basexml.config")
			defer file.Close()
			result, _ := parseXML(file)
			result.Sort()
			It("Should return matching basic xml ValidateBlob", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})
		Context("When parsing invalid XML", func() {
			file, err := os.Open("fixtures/validate_badxml.config")
			defer file.Close()
			result, err := parseXML(file)
			It("Should return error parsing xml", func() {
				Expect(err.Error()).To(Equal("xml.Decoder.Token() - XML syntax error on line 32: element <log> closed by </configuration>"))
				Expect(result).To(Equal(tasks.ValidateBlob{}))
			})
		})
		Context("When parsing deeply nested XML", func() {
			file, _ := os.Open("fixtures/validate_basexml.config")
			defer file.Close()
			result, _ := parseXML(file)
			result.Sort()
			It("Should return matching basic xml ValidateBlob", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})
		Context("When error reading XML file", func() {
			reader := mockReader{}
			reader.Reader = func([]byte) (int, error) {
				return 0, errReaderMock
			}
			result, err := parseXML(reader)
			result.Sort()
			It("Should return error reading file", func() {
				Expect(result).To(Equal(tasks.ValidateBlob{}))
				Expect(err).To(Equal(fmt.Errorf("%v : %v", errConfigFileNotRead, errReaderMock)))
			})
		})
		Context("When parsing an xml array", func() {
			reader := strings.NewReader(`<ignoreStatusCodes>
			<code>401</code>
			<code>404</code>
				</ignoreStatusCodes>`)
			result, err := parseXML(reader)
			result.Sort()
			It("Should return expected nested array of ValidateBlob", func() {
				expectedResult := tasks.ValidateBlob{
					Children: []tasks.ValidateBlob{
						tasks.ValidateBlob{
							Key: "ignoreStatusCodes",
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:  "code",
									Path: "/ignoreStatusCodes",
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "0",
											Path:     "/ignoreStatusCodes/code",
											RawValue: "401",
										},
										tasks.ValidateBlob{
											Key:      "1",
											Path:     "/ignoreStatusCodes/code",
											RawValue: "404",
											Children: nil,
										},
									},
								},
							},
						},
					},
				}

				Expect(result).To(Equal(expectedResult))
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("parseJS", func() {
		format.TruncatedDiff = false
		Context("When parsing standard js", func() {
			file, _ := os.Open("fixtures/validate_testdata_js_config.js")
			defer file.Close()
			result, err := parseJs(file)
			result.Sort()
			It("Should return matching basic js ValidateBlob", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
				Expect(err).To(BeNil())
			})
		})
		Context("When parsing invalid js config", func() {
			file, _ := os.Open("fixtures/validate_testdata_js_comment.js")
			defer file.Close()
			result, _ := parseJs(file)
			It("Should return special character key in ValidateBlob", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})

		Context("When parsing Typescript Config", func() {
			file, _ := os.Open("fixtures/validate_testdata_type_script.js")
			defer file.Close()
			result, _ := parseJs(file)
			result.Sort()
			It("Should return matching typescript config", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})
		Context("When parsing ES6 Config", func() {
			file, _ := os.Open("fixtures/validate_testdata_es6.js")
			defer file.Close()
			result, _ := parseJs(file)
			result.Sort()
			It("Should return matching es6 config", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})
		Context("When parsing js config with arrays", func() {
			file, _ := os.Open("../../fixtures/node/newrelicWithArrays.js")
			defer file.Close()
			result, _ := parseJs(file)
			result.Sort()
			It("Should return matching array elements", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
			})
		})

		Context("When parsing js array", func() {
			reader := strings.NewReader(`exports.config = { 
				app_name: ['My Node App', 'my appname 2'],
				}`)
			result, _ := parseJs(reader)
			result.Sort()
			It("Should return matching array config", func() {
				expectedResult := tasks.ValidateBlob{
					Children: []tasks.ValidateBlob{
						tasks.ValidateBlob{
							Key: "app_name",
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "0",
									Path:     "/app_name",
									RawValue: "My Node App",
								},
								tasks.ValidateBlob{
									Key:      "1",
									Path:     "/app_name",
									RawValue: "my appname 2",
								},
							},
						},
					},
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

	})

	Describe("parseYML", func() {
		Context("When parsing standard yml", func() {
			reader, _ := os.Open("../../fixtures/java/newrelic/newrelic.yml")
			result, err := ParseYaml(reader)
			result.Sort()
			It("Should return matching basic yml ValidateBlob", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(result.String()), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.String()).To(Equal(expectedResult))
				Expect(err).To(BeNil())
			})
		})
		Context("When error reading yml file", func() {
			reader := mockReader{}
			reader.Reader = func([]byte) (int, error) {
				return 0, errReaderMock
			}
			result, err := ParseYaml(reader)
			result.Sort()
			It("Should return error reading file", func() {
				Expect(result).To(Equal(tasks.ValidateBlob{}))
				Expect(err).To(Equal(fmt.Errorf("%v : %v", errConfigFileNotRead, errReaderMock)))
			})
		})
		Context("When error unmarshalling yml file", func() {
			reader := strings.NewReader(`invalidyaml:
			  [not valid`)
			result, err := ParseYaml(reader)
			result.Sort()
			It("Should return error parsing yml", func() {
				Expect(result).To(Equal(tasks.ValidateBlob{}))
				Expect(err.Error()).To(Equal("yaml: line 2: found character that cannot start any token.\nThis can mean that you either have incorrect spacing/indentation around this line or that you have a syntax error, such as a missing/invalid character"))
			})
		})

	})

	Describe("ParseJSON", func() {
		Context("When error reading JSON file", func() {
			reader := mockReader{}
			reader.Reader = func([]byte) (int, error) {
				return 0, errReaderMock
			}
			result, err := ParseJSON(reader)
			result.Sort()
			It("Should return error reading file", func() {
				Expect(result).To(Equal(tasks.ValidateBlob{}))
				Expect(err).To(Equal(fmt.Errorf("%v : %v", errConfigFileNotRead, errReaderMock)))
			})
		})
		Context("When error parsing JSON file", func() {
			reader := strings.NewReader(`[invalid:			{json::here}]`)
			result, err := ParseJSON(reader)
			result.Sort()
			It("Should return error parsing json", func() {
				Expect(result).To(Equal(tasks.ValidateBlob{}))
				Expect(err.Error()).To(Equal("invalid character 'i' looking for beginning of value"))
			})
		})
	})

	Describe("convertToValidateBlob", func() {
		Context("When parsing key/value pair", func() {
			input := map[string]interface{}{"key": "value"}
			result := convertToValidateBlob(input)
			result.Sort()
			It("Should return matching key-value ValidateBlob", func() {
				expectedResult := tasks.ValidateBlob{
					Children: []tasks.ValidateBlob{
						tasks.ValidateBlob{Key: "key", RawValue: "value"},
					},
				}.String()
				Expect(result.String()).To(Equal(expectedResult))
			})
		})
		Context("When parsing map[string]interface{}", func() {
			input := map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			result := convertToValidateBlob(input)
			result.Sort()
			It("Should return matching map[string]interface ValidateBlob", func() {
				expectedResult := tasks.ValidateBlob{Children: []tasks.ValidateBlob{
					tasks.ValidateBlob{Key: "key1", RawValue: "value1"},
					tasks.ValidateBlob{Key: "key2", RawValue: "value2"},
					tasks.ValidateBlob{Key: "key3", RawValue: "value3"},
				}}.String()
				Expect(result.String()).To(Equal(expectedResult))
			})
			Context("When parsing nested map[interface{}]interface with a key which can't be converted to a string", func() {
				input := map[interface{}]interface{}{
					"": map[string]interface{}{"nestedkey": "nestedvalue"},
					42: "the answer",
				}

				result := convertToValidateBlob(input)
				result.Sort()
				It("Should skip the invalid map element and return matching nested map[interface{}]interface ValidateBlob", func() {
					expectedResult := tasks.ValidateBlob{Children: []tasks.ValidateBlob{
						tasks.ValidateBlob{Children: []tasks.ValidateBlob{
							tasks.ValidateBlob{Key: "nestedkey", Path: "/", RawValue: "nestedvalue"},
						}},
					}}.String()
					Expect(result.String()).To(Equal(expectedResult))
				})
			})

			Context("When parsing nested map[interface{}]interface", func() {
				input := map[string]interface{}{"": map[interface{}]interface{}{"nestedkey": "nestedvalue"}}

				result := convertToValidateBlob(input)
				result.Sort()
				It("Should return matching nested map[interface{}]interface ValidateBlob", func() {
					expectedResult := tasks.ValidateBlob{Children: []tasks.ValidateBlob{
						tasks.ValidateBlob{Children: []tasks.ValidateBlob{
							tasks.ValidateBlob{Key: "nestedkey", Path: "/", RawValue: "nestedvalue"},
						}},
					}}.String()
					Expect(result.String()).To(Equal(expectedResult))
				})
			})
		})

	})

	Describe("processConfig", func() {
		Context("When parsing a standard .yml file", func() {
			input := ConfigElement{
				FileName: "newrelic.yml",
				FilePath: "../../fixtures/java/newrelic/",
			}
			result, processErr := processConfig(input)
			It("Should return parsed yml", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}

				expectedResult := readFile(goldenFile)
				Expect(fmt.Sprintf("#%v", result)).To(Equal(expectedResult))
				Expect(processErr).To(BeNil())
			})
		})
		Context("When parsing a standard .json file", func() {
			input := ConfigElement{
				FileName: "private-location-settings-full.json",
				FilePath: "../../fixtures/Synthetics/root/opt/newrelic/synthetics/.newrelic/synthetics/minion/",
			}
			result, processErr := processConfig(input)
			It("Should return parsed json", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(result.Config).To(Equal(input))
				Expect(fmt.Sprintf("#%v", result)).To(Equal(expectedResult))
				Expect(processErr).To(BeNil())
			})
		})
		Context("When parsing a standard .js file", func() {
			input := ConfigElement{
				FileName: "newrelic.js",
				FilePath: "../../fixtures/node/",
			}
			result, processErr := processConfig(input)
			It("Should return parsed js", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(fmt.Sprintf("#%v", result)).To(Equal(expectedResult))
				Expect(processErr).To(BeNil())
			})
		})
		Context("When parsing a standard .cfg file", func() {
			input := ConfigElement{
				FileName: "newrelic.cfg",
				FilePath: "../../fixtures/php/root/etc/newrelic/",
			}
			result, processErr := processConfig(input)
			It("Should return parsed cfg", func() {
				goldenFile := goldenFileName(CurrentGinkgoTestDescription().TestText)

				if updateGoldenFiles {
					if err := ioutil.WriteFile(goldenFile, []byte(fmt.Sprintf("#%v", result)), 0644); err != nil {
						Expect(err).To(BeNil())
					}
				}
				expectedResult := readFile(goldenFile)
				Expect(fmt.Sprintf("#%v", result)).To(Equal(expectedResult))
				Expect(processErr).To(BeNil())
			})
		})
		Context("When attempting to parse a .gradle file", func() {
			input := ConfigElement{
				FileName: "build.gradle",
				FilePath: "../../fixtures/Android/Android-Sample/",
			}
			result, processErr := processConfig(input)
			It("Should return parsed gradle", func() {
				expectedResult := ValidateElement{}
				Expect(result).To(Equal(expectedResult))
				Expect(processErr).To(Not(BeNil()))
			})
		})
		Context("When parsing an unknown file", func() {
			input := ConfigElement{
				FileName: "validation_blah.stuff",
				FilePath: "fixtures/",
			}
			result, processErr := processConfig(input)
			It("Should return default ValidateElement", func() {
				expectedResult := ValidateElement{}
				Expect(result).To(Equal(expectedResult))
				Expect(processErr).To(Not(BeNil()))
			})
		})
	})

})

func readFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Info("error reading file", err)
	}
	// remove windows file endings
	return strings.Replace(string(content), "\r", "", -1)
}
