package tasks

import (
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTaskHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Task Helpers test suite")
}

func TestValidateBlob_FindKeyByPath(t *testing.T) {
	tests := []struct {
		name       string
		fields     ValidateBlob
		searchPath string
		want       ValidateBlob
	}{
		{"Find key by exact path match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"/Production/key2",
			ValidateBlob{Key: "key2", Path: "/Production"},
		},
		{"Find key by deep path match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
			"/Production/agent_enabled/enabled",
			ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
		},
		{"Fail to find key by partial path match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"/Production/",
			ValidateBlob{},
		},
		{"Fail to find key by key only ",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"key2",
			ValidateBlob{},
		},
		{"Empty result from invalid path",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"/Production/my_awesome_key",
			ValidateBlob{},
		},

		//{"Find key by exact path match xml"},
		//{"Empty result from partial path"},
		//{"Empty result from invalid path"}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidateBlob{
				Key:      tt.fields.Key,
				Path:     tt.fields.Path,
				RawValue: tt.fields.RawValue,
				Children: tt.fields.Children,
			}
			if got := v.FindKeyByPath(tt.searchPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateBlob.FindKeyByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBlob_FindKey(t *testing.T) {
	tests := []struct {
		name        string
		fields      ValidateBlob
		searchKey   string
		wantResults []ValidateBlob
	}{
		{"Find key by exact key match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"key2",
			[]ValidateBlob{
				ValidateBlob{Key: "key2", Path: "/Production"},
			},
		},
		{"Find key by deep path match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
			"enabled",
			[]ValidateBlob{
				ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
			},
		},
		{"Find key match with nested object",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{
						Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "key1", Path: "/Production"},
							ValidateBlob{Key: "key2", Path: "/Production"},
						},
					},
				}},
			"Production",
			[]ValidateBlob{
				ValidateBlob{
					Key: "Production", Path: "/",
					Children: []ValidateBlob{
						ValidateBlob{Key: "key1", Path: "/Production"},
						ValidateBlob{Key: "key2", Path: "/Production"},
					},
				},
			},
		},
		{"Empty result from invalid key",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			"my_awesome_key",
			nil, //This is nil because of how DeepEqual interacts on empty slices
		},
		{"Find key by deep path match from map object",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
			"enabled",
			[]ValidateBlob{
				ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidateBlob{
				Key:      tt.fields.Key,
				Path:     tt.fields.Path,
				RawValue: tt.fields.RawValue,
				Children: tt.fields.Children,
			}
			if gotResults := v.FindKey(tt.searchKey); !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("ValidateBlob.FindKey() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}

func TestValidateBlob_UpdateKey(t *testing.T) {
	type args struct {
		searchKey        string
		replacementValue interface{}
	}
	tests := []struct {
		name   string
		fields ValidateBlob
		args   args
		want   ValidateBlob
	}{
		{"Update key by exact key match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production", RawValue: "teststring1"},
					ValidateBlob{Key: "key2", Path: "/Production", RawValue: "teststring2"},
				}},
			args{searchKey: "/Production/key2", replacementValue: "teststringAWESOME"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production", RawValue: "teststring1"},
					ValidateBlob{Key: "key2", Path: "/Production", RawValue: "teststringAWESOME"},
				}},
		},
		{"Return original ValidateBlob by partial key match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
			args{searchKey: "enabled", replacementValue: "stuff"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
		},
		{"Update key with deep nested object",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled", RawValue: "initial"},
								}},
						}},
				}},
			args{searchKey: "/Production/agent_enabled/enabled", replacementValue: "value1"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled", RawValue: "value1", Children: nil},
								}},
						}},
				}},
		},
		{"Same ValidateBlob from invalid key",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
			args{searchKey: "my_awesome_key", replacementValue: "value1"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "key2", Path: "/Production"},
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidateBlob{
				Key:      tt.fields.Key,
				Path:     tt.fields.Path,
				RawValue: tt.fields.RawValue,
				Children: tt.fields.Children,
			}
			got := v.UpdateKey(tt.args.searchKey, tt.args.replacementValue)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateBlob.UpdateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBlob_InsertKey(t *testing.T) {
	type args struct {
		insertKey string
		value     interface{}
	}
	tests := []struct {
		name   string
		fields ValidateBlob
		args   args
		want   ValidateBlob
	}{
		{"insert key at top level",
			ValidateBlob{},
			args{insertKey: "newKey", value: "Value1"},
			ValidateBlob{Children: []ValidateBlob{
				ValidateBlob{
					Key: "newKey", RawValue: "Value1",
				},
			},
			},
		},
		{"insert key at 1st level",
			ValidateBlob{Key: "Production"},
			args{insertKey: "newKey", value: "Value1"},
			ValidateBlob{
				Key: "Production",
				Children: []ValidateBlob{
					ValidateBlob{
						Key: "newKey", RawValue: "Value1",
					},
				},
			},
		},
		{"insert key at deep nested level",
			ValidateBlob{Key: "Production"},
			args{insertKey: "/transaction_tracer/slow_sql/enabled", value: true},
			ValidateBlob{
				Key: "Production",
				Children: []ValidateBlob{
					ValidateBlob{
						Key: "transaction_tracer", Path: "/Production",
						Children: []ValidateBlob{
							ValidateBlob{
								Key: "slow_sql", Path: "/Production/transaction_tracer",
								Children: []ValidateBlob{
									ValidateBlob{
										Key: "enabled", RawValue: true, Path: "/Production/transaction_tracer/slow_sql",
									},
								},
							},
						},
					},
				},
			},
		},
		{"insert key at non-existent node",
			ValidateBlob{Key: "Production"},
			args{insertKey: "/transaction_tracer/newKey", value: "Value1"},
			ValidateBlob{
				Key: "Production", Children: []ValidateBlob{
					ValidateBlob{Key: "transaction_tracer", Path: "/Production", Children: []ValidateBlob{
						ValidateBlob{Key: "newKey", Path: "/Production/transaction_tracer", RawValue: "Value1"},
					}},
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidateBlob{
				Key:      tt.fields.Key,
				Path:     tt.fields.Path,
				RawValue: tt.fields.RawValue,
				Children: tt.fields.Children,
			}
			if got := v.InsertKey(tt.args.insertKey, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateBlob.InsertKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBlob_UpdateOrInsertKey(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name   string
		fields ValidateBlob
		args   args
		want   ValidateBlob
	}{
		{"insert key at top level",
			ValidateBlob{},
			args{key: "newKey", value: "Value1"},
			ValidateBlob{Children: []ValidateBlob{
				ValidateBlob{
					Key: "newKey", RawValue: "Value1",
				},
			},
			},
		},
		{"insert key at 1st level",
			ValidateBlob{Key: "Production"},
			args{key: "newKey", value: "Value1"},
			ValidateBlob{
				Key: "Production",
				Children: []ValidateBlob{
					ValidateBlob{
						Key: "newKey", RawValue: "Value1",
					},
				},
			},
		},
		{"Update key by exact key match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production", RawValue: "teststring1"},
					ValidateBlob{Key: "key2", Path: "/Production", RawValue: "teststring2"},
				}},
			args{key: "/Production/key2", value: "teststringAWESOME"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production", RawValue: "teststring1"},
					ValidateBlob{Key: "key2", Path: "/Production", RawValue: "teststringAWESOME"},
				}},
		},
		{"Return original ValidateBlob by partial key match",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled"},
								}},
						}},
				}},
			args{key: "enabled", value: "stuff"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled", RawValue: "stuff"},
								}},
						}},
				}},
		},
		{"Update key with deep nested object",
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled", RawValue: "initial"},
								}},
						}},
				}},
			args{key: "/Production/agent_enabled/enabled", value: "value1"},
			ValidateBlob{
				Children: []ValidateBlob{
					ValidateBlob{Key: "key1", Path: "/Production"},
					ValidateBlob{Key: "Production", Path: "/",
						Children: []ValidateBlob{
							ValidateBlob{Key: "agent_enabled", Path: "/Production",
								Children: []ValidateBlob{
									ValidateBlob{Key: "enabled", Path: "/Production/agent_enabled", RawValue: "value1", Children: nil},
								}},
						}},
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidateBlob{
				Key:      tt.fields.Key,
				Path:     tt.fields.Path,
				RawValue: tt.fields.RawValue,
				Children: tt.fields.Children,
			}
			if got := v.UpdateOrInsertKey(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateBlob.UpdateOrInsertKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBlob_String(t *testing.T) {
	tests := []struct {
		name       string
		input      ValidateBlob
		wantOutput string
	}{
		{name: "Simple blob", input: ValidateBlob{Key: "key1", RawValue: "value1"}, wantOutput: "/key1: value1\n"},
		{name: "Simple blob with Child", input: ValidateBlob{Key: "key1", Children: []ValidateBlob{
			ValidateBlob{Key: "key2", RawValue: "value2", Path: "/key1"},
		}}, wantOutput: "/key1/key2: value2\n"},
		{name: "Simple blob with multiple Children", input: ValidateBlob{Key: "key1", Children: []ValidateBlob{
			ValidateBlob{Key: "key2", RawValue: "value2", Path: "/key1"},
			ValidateBlob{Key: "key3", RawValue: "value3", Path: "/key1"},
		}}, wantOutput: "/key1/key2: value2\n/key1/key3: value3\n"},
		{name: "blob with multiple nested Children", input: ValidateBlob{Key: "key1", Children: []ValidateBlob{
			ValidateBlob{Key: "key2", RawValue: "value2", Path: "/key1"},
			ValidateBlob{Key: "key3", Path: "/key1", Children: []ValidateBlob{
				ValidateBlob{Key: "key4", RawValue: "value4", Path: "/key1/key3"},
				ValidateBlob{Key: "key5", RawValue: "value5", Path: "/key1/key3"},
			}},
		}}, wantOutput: "/key1/key2: value2\n/key1/key3/key4: value4\n/key1/key3/key5: value5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutput := tt.input.String(); gotOutput != tt.wantOutput {
				t.Errorf("ValidateBlob.String() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestDedupeStringSlice(t *testing.T) {
	dupedSlice := []string{"llamas", "llamas1", "llamas"}
	dedupedSlice := []string{"llamas", "llamas1"}
	noDupeSlice := []string{"llamas", "llamas1", "llamas2", "llamas3"}

	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "It should dedupe a slice with duplicates",
			args: args{s: dupedSlice},
			want: dedupedSlice,
		},
		{
			name: "It should return a slice with identical elements if presented a list with no duplicates",
			args: args{s: noDupeSlice},
			want: noDupeSlice,
		},
		{
			name: "It return an empty slice when given an empty slice",
			args: args{s: []string{}},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DedupeStringSlice(tt.args.s)
			// Ordering of return will be randomized, so sort before comparing
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DedupeStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	type args struct {
		s    string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "It returns true if the supplied string is in the slice",
			args: args{
				s:    "llamas",
				list: []string{"llamas", "yaks", "emus"},
			},
			want: true,
		},
		{
			name: "It returns false if the supplied string is not in the slice",
			args: args{
				s:    "kudus",
				list: []string{"llamas", "yaks", "emus"},
			},
			want: false,
		},
		{
			name: "It handles an empty slice",
			args: args{
				s:    "kudus",
				list: []string{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringInSlice(tt.args.s, tt.args.list); got != tt.want {
				t.Errorf("StringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = Describe("Task Helpers", func() {

	Describe("GetWorkingDirectories", func() {
		var (
			directories []string
		)
		JustBeforeEach(func() {
			directories = GetWorkingDirectories()
		})

		Context("when os.Getwd and os.Executable return expected result", func() {
			BeforeEach(func() {
				osGetwd = func() (string, error) {
					return "/foo", nil
				}
				osExecutable = func() (string, error) {
					return "/bar", nil
				}
			})
			It("should return expected slice", func() {
				Expect(directories).To(Equal([]string{"/foo", "/bar"}))
			})
		})

		Context("when os.Getwd returns an error and os.Executable return expected result", func() {
			BeforeEach(func() {
				osGetwd = func() (string, error) {
					return "", errors.New("i am a robot error")
				}
				osExecutable = func() (string, error) {
					return "/bar", nil
				}
			})
			It("should return expected slice", func() {
				Expect(directories).To(Equal([]string{"/bar"}))
			})
		})

		Context("when os.Getwd return expected result and os.Executable returns an error", func() {
			BeforeEach(func() {
				osGetwd = func() (string, error) {
					return "/foo", nil
				}
				osExecutable = func() (string, error) {
					return "", errors.New("this is a meatbag error")
				}
			})
			It("should return expected slice", func() {
				Expect(directories).To(Equal([]string{"/foo"}))
			})
		})

		Context("when os.Getwd and os.Executable both return an error", func() {
			BeforeEach(func() {
				osGetwd = func() (string, error) {
					return "", errors.New("i like pandas")
				}
				osExecutable = func() (string, error) {
					return "", errors.New("i also like frogs")
				}
			})
			It("should return expected slice", func() {
				Expect(directories).To(BeNil())
			})
		})

	})

	Describe("FindFiles", func() {
		var (
			patterns []string
			paths    []string
			files    []string
		)

		JustBeforeEach(func() {
			files = FindFiles(patterns, paths)
		})

		Context("With no input", func() {
			BeforeEach(func() {
			})
			It("Should return nil", func() {
				Expect(files).To(BeNil())
			})
		})
		Context("With standard path", func() {
			BeforeEach(func() {
				patterns = []string{"newrelic.yml"}
				paths = []string{"fixtures/ruby/config/"}
			})
			It("Should return found file slice", func() {
				Expect(files).To(Equal([]string{filepath.FromSlash("fixtures/ruby/config/newrelic.yml")}))
			})
		})
		Context("With pattern regex that ends in a $", func() {
			BeforeEach(func() {
				patterns = []string{".+[.]y(a)?ml$"}
				paths = []string{"fixtures/FindFiles/sampleYml"}
			})
			It("Should return found file slice", func() {
				Expect(files).To(BeNil())
			})
		})

		Context("With non-existing path", func() {
			BeforeEach(func() {
				patterns = []string{"newrelic.config"}
				paths = []string{"stuff/non"}
			})
			It("Should return nil", func() {
				Expect(files).To(BeNil())
			})
		})
		Context("With existing and non-existing path", func() {
			BeforeEach(func() {
				patterns = []string{"newrelic.yml"}
				paths = []string{"stuff/non", "fixtures/ruby/config/"}
			})
			It("Should return the expected filepath", func() {
				Expect(files).To(Equal([]string{filepath.FromSlash("fixtures/ruby/config/newrelic.yml")}))
			})
		})
		Context("With nested directory adjacent to symbolic link directory", func() {
			BeforeEach(func() {
				patterns = []string{"newrelic_agent.log"}
				paths = []string{"fixtures/FindFiles"}
			})
			It("Should return found file slice", func() {
				Expect(files).To(Equal([]string{filepath.FromSlash("fixtures/FindFiles/nestedFolder/moreNested/newrelic_agent.log")}))
			})
		})
		//This test doesn't need to run on windows since it deals with symbolic links and github doesn't deal with sym links on windows
		if runtime.GOOS != "windows" {
			Context("With symbolic link directory", func() {
				BeforeEach(func() {
					patterns = []string{"sample-fortest.yml"}
					paths = []string{"fixtures/FindFiles/symlinktodir"}
				})
				It("Should return found file slice", func() {
					Expect(files).To(Equal([]string{filepath.FromSlash("../output/fixtures/symbolicsrc/sample-fortest.yml")}))
				})
			})

		}
	})

	Describe("ValidateBlob", func() {
		Describe("Sort", func() {
			Context("When sorting", func() {
				validateBlob := ValidateBlob{Key: "key1", Children: []ValidateBlob{
					ValidateBlob{Key: "bkey2", RawValue: "value2", Path: "/key1"},
					ValidateBlob{Key: "akey3", RawValue: "value3", Path: "/key1"},
				}}
				expectedValidateBlob := ValidateBlob{Key: "key1", Children: []ValidateBlob{
					ValidateBlob{Key: "akey3", RawValue: "value3", Path: "/key1"},
					ValidateBlob{Key: "bkey2", RawValue: "value2", Path: "/key1"},
				}}
				It("Should organize by children", func() {
					validateBlob.Sort()
					Expect(validateBlob).To(Equal(expectedValidateBlob))
				})
			})

			Context("When sorting deeply nested", func() {
				validateBlob := ValidateBlob{Key: "key1", Children: []ValidateBlob{
					ValidateBlob{Key: "bkey2", RawValue: "value2", Path: "/key1"},
					ValidateBlob{Key: "akey3", RawValue: "value3", Path: "/key1"},
					ValidateBlob{Key: "ckey4", Path: "/key1", Children: []ValidateBlob{
						ValidateBlob{Key: "zkey6", RawValue: "value6", Path: "/key1/ckey4"},
						ValidateBlob{Key: "fkey5", RawValue: "value5", Path: "/key1/ckey4"},
					},
					},
				}}
				expectedValidateBlob := ValidateBlob{
					Key: "key1", Children: []ValidateBlob{
						ValidateBlob{Key: "akey3", RawValue: "value3", Path: "/key1"},
						ValidateBlob{Key: "bkey2", RawValue: "value2", Path: "/key1"},
						ValidateBlob{Key: "ckey4", Path: "/key1", Children: []ValidateBlob{
							ValidateBlob{Key: "fkey5", RawValue: "value5", Path: "/key1/ckey4"},
							ValidateBlob{Key: "zkey6", RawValue: "value6", Path: "/key1/ckey4"},
						},
						},
					},
				}
				It("Should organize by children", func() {
					validateBlob.Sort()
					Expect(validateBlob).To(Equal(expectedValidateBlob))
				})
			})
		})
	})

})
