package tasks

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("file search Task helpers", func() {

	Describe("FindStringInFile", func() {
		var (
			search   string
			filepath string
			exists   bool
		)
		JustBeforeEach(func() {
			exists = FindStringInFile(search, filepath)
		})
		Context("When invalid regex supplied", func() {
			BeforeEach(func() {
				search = "blarh[tryfh"
				filepath = ""
				osOpen = os.Open
			})
			It("Should return false", func() {
				Expect(exists).To(BeFalse())
			})
		})
		Context("When error opening file", func() {
			BeforeEach(func() {
				search = ""
				filepath = ""
				osOpen = func(string) (*os.File, error) {
					return nil, errors.New("error opening file")
				}
			})
			It("Should return false", func() {
				Expect(exists).To(BeFalse())
			})
		})
		Context("When regex not in file ", func() {
			BeforeEach(func() {
				search = "not in the file!!!"
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				osOpen = os.Open
			})
			It("Should return false", func() {
				Expect(exists).To(BeFalse())
			})
		})
		Context("When regex not in file ", func() {
			BeforeEach(func() {
				search = "us-west-2.compute.internal"
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				osOpen = os.Open
			})
			It("Should return false", func() {
				Expect(exists).To(BeTrue())
			})
		})

	})

	Describe("ReturnStringSubmatchInFileAllMatches", func() {
		var (
			search   string
			filepath string
			match    [][]string
			err      error
		)
		JustBeforeEach(func() {
			match, err = ReturnStringSubmatchInFileAllMatches(search, filepath)
		})
		Context("when reading php log from end, find all matching strings", func() {
			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = "us-west-2.compute.internal"
				osOpen = os.Open
			})

			It("should return expected string matching search string ", func() {
				expectedMatches := [][]string{
					[]string{"us-west-2.compute.internal"},
					[]string{"us-west-2.compute.internal"},
					[]string{"us-west-2.compute.internal"},
				}
				Expect(match).To(Equal(expectedMatches))
			})
			It("should return right number of matches", func() {
				Expect(len(match)).To(Equal(3))
			})
		})
		Context("when reading php log from end, find matching capture group", func() {
			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = `Relic (\d+\.\d\.\d\.\d{3})`
				osOpen = os.Open
			})

			It("should return expected string matching search regex ", func() {
				expectedMatches := [][]string{
					[]string{"Relic 1.2.4.200", "1.2.4.200"},
					[]string{"Relic 2.3.4.500", "2.3.4.500"},
					[]string{"Relic 7.5.0.199", "7.5.0.199"},
				}
				Expect(match).To(Equal(expectedMatches))
			})
		})
		Context("when search string not in file", func() {
			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = "xia"
				osOpen = os.Open
			})

			It("should return an empty slice of slices, because no match is found ", func() {
				Expect(match).To(Equal([][]string{}))
				Expect(err).To(BeNil())
			})
		})
		Context("when opening file causes an error", func() {
			BeforeEach(func() {
				search = ""
				filepath = "no way this works"
				osOpen = os.Open
			})
			It("should err that is not nil", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		Context("when invalid regex supplied", func() {
			BeforeEach(func() {
				search = "bla[k"
				filepath = ""
				osOpen = os.Open
			})
			It("should err that is not nil", func() {
				Expect(err.Error()).To(Equal("error parsing regexp: missing closing ]: `[k`"))
			})
		})
		Context("when os.Open returns error", func() {
			BeforeEach(func() {
				search = ""
				filepath = "brokenfile"
				osOpen = func(string) (*os.File, error) {
					return nil, errors.New("error opening file")
				}
			})
			It("should err that is not nil", func() {
				Expect(err.Error()).To(Equal("error opening file"))
			})
		})
	})

	Describe("ReturnStringSubmatchInFile", func() {
		var (
			search   string
			filepath string
			matches  []string
			err      error
		)

		JustBeforeEach(func() {
			matches, err = ReturnStringSubmatchInFile(search, filepath)
		})

		Context("when multiple matches found", func() {

			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = `Relic (\d+\.\d\.\d\.\d{3})`
				osOpen = os.Open
			})

			It("should return only first result", func() {
				expectedMatch := []string{
					"Relic 1.2.4.200",
					"1.2.4.200",
				}
				Expect(matches).To(Equal(expectedMatch))
			})
			It("should return a nil error", func() {
				Expect(err).To(BeNil())
			})

		})

		Context("when match not found", func() {
			BeforeEach(func() {
				search = "invalid string"
				filepath = "fixtures/fileSearch/newrelic_agent.log"
			})

			It("should return an empty slice of strings ", func() {
				Expect(matches).To(Equal([]string{}))
			})
			It("should return nil for error", func() {
				Expect(err).To(BeNil())
			})

		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				search = ")) bad regex! #(*&$"
				filepath = "fixtures/fileSearch/newrelic_agent.log"
			})

			It("should return an empty slice of strings", func() {
				Expect(matches).To(Equal([]string{}))
			})
			It("should return an error", func() {
				Expect(err).To(Not(BeNil()))
			})

		})
	})

	Describe("ReturnLastStringSubmatchInFile", func() {
		var (
			search   string
			filepath string
			matches  []string
			err      error
		)

		JustBeforeEach(func() {
			matches, err = ReturnLastStringSubmatchInFile(search, filepath)
		})

		Context("when multiple matches found", func() {

			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = `Relic (\d+\.\d\.\d\.\d{3})`
				osOpen = os.Open
			})

			It("should return only last result", func() {
				expectedMatch := []string{
					"Relic 7.5.0.199",
					"7.5.0.199",
				}
				Expect(matches).To(Equal(expectedMatch))
			})
			It("should return a nil error", func() {
				Expect(err).To(BeNil())
			})

		})

		Context("when match with capture groups is found", func() {

			BeforeEach(func() {
				filepath = "fixtures/fileSearch/newrelic_agent.log"
				search = `info: New Relic (?P<version>(?P<major>\d)(\.)?(?P<minor>\d+)(\.)?(?P<patch>\d+)(\.)?(?P<build>\d+)) \("(?P<codename>[^"]+)"`
				osOpen = os.Open
			})

			It("should return only last result", func() {
				expecteVersionMatch := "7.5.0.199"
				Expect(matches[1]).To(Equal(expecteVersionMatch))
			})
			It("should return a nil error", func() {
				Expect(err).To(BeNil())
			})

		})

		Context("when match not found", func() {
			BeforeEach(func() {
				search = "invalid string"
				filepath = "fixtures/fileSearch/newrelic_agent.log"
			})

			It("should return an empty slice of strings ", func() {
				Expect(matches).To(Equal([]string{}))
			})
			It("should return nil for error", func() {
				Expect(err).To(BeNil())
			})

		})
	})

})
