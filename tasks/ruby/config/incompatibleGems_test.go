package config

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestRubyConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ruby/Config test suite")
}

var _ = Describe("Ruby/Config/IncompatibleGems", func() {
	Describe("checkGems", func() {
		var (
			gemFiles []string
			output   []BadGemAndPath
		)

		JustBeforeEach(func() {
			output = checkGems(gemFiles)
		})

		Context("When given a list with one bad gem", func() {
			BeforeEach(func() {

				gemFiles = []string{
					"../../fixtures/ruby/config/badgems_one_Gemfile",
				}
			})
			It("Should return a slice containing the bad gemfile and path", func() {
				expectedOutput := []BadGemAndPath{
					BadGemAndPath{
						GemName:     "db-charmer",
						GemfilePath: "../../fixtures/ruby/config/badgems_one_Gemfile",
					},
				}
				Expect(output).To(Equal(expectedOutput))

			})
		})
	})

	Describe("displayResults()", func() {
		var (
			//input
			incompatibleGems []BadGemAndPath

			//output
			summary string
		)

		JustBeforeEach(func() {
			summary = displayResults(incompatibleGems)
		})

		Context("When given an empty list of incompatible gems", func() {
			BeforeEach(func() {
				incompatibleGems = []BadGemAndPath{}
			})
			It("Should return expected summary", func() {
				Expect(summary).To(Equal("There were no incompatible gems found."))
			})
		})

		Context("When given a single incompatible gem", func() {
			format.TruncatedDiff = false
			BeforeEach(func() {
				incompatibleGems = []BadGemAndPath{{
					GemName:     "ar-octopus",
					GemfilePath: "/earth/americas/north/usa/illinois/chicago/shedd_aquarium/gemfile.lock",
				}}
			})
			It("Should return expected summary identifying the incompatible gem", func() {
				expectedSummaryLines := []string{
					"We detected 1 Ruby gem(s) incompatible with the New Relic Ruby agent:",
					"ar-octopus - /earth/americas/north/usa/illinois/chicago/shedd_aquarium/gemfile.lock",
				}
				Expect(summary).To(Equal(strings.Join(expectedSummaryLines, "\n")))
			})
		})

		Context("When given multiple incompatible gems", func() {
			format.TruncatedDiff = false
			BeforeEach(func() {
				incompatibleGems = []BadGemAndPath{{
					GemName:     "ar-octopus",
					GemfilePath: "/path1/gemfile.lock",
				},
					{
						GemName:     "db-charmer",
						GemfilePath: "/path2/gemfile.lock",
					},
					{
						GemName:     "escape_utils",
						GemfilePath: "/path3/gemfile.lock",
					},
					{
						GemName:     "escape_utils",
						GemfilePath: "/this/is/another/folder/gemfile.lock",
					},
					{
						GemName:     "right_http_connection",
						GemfilePath: "/path4/gemfile.lock",
					},
				}
			})
			It("Should return expected summary identifying the incompatible gems", func() {
				expectedSummaryLines := []string{
					"We detected 5 Ruby gem(s) incompatible with the New Relic Ruby agent:",
					"ar-octopus - /path1/gemfile.lock",
					"db-charmer - /path2/gemfile.lock",
					"escape_utils - /path3/gemfile.lock",
					"escape_utils - /this/is/another/folder/gemfile.lock",
					"right_http_connection - /path4/gemfile.lock",
				}
				Expect(summary).To(Equal(strings.Join(expectedSummaryLines, "\n")))
			})
		})
	})
})
