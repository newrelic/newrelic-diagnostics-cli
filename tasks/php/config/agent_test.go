package config

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

func TestPHPAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PHP/Config test suite")
}

var _ = Describe("PHP/Config/Agent", func() {

	Describe("checkValidation()", func() {
		var input []config.ValidateElement
		var checkedConfigs []config.ValidateElement
		var configFound bool

		JustBeforeEach(func() {
			checkedConfigs, configFound = checkValidation(input)
		})

		Context("When given slice of one valid php config", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{
					{
						Config: config.ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
				}
			})

			It("Should return a slice of 1 validate element", func() {
				Expect(len(checkedConfigs)).To(Equal(1))
				Expect(checkedConfigs).To(Equal(input))
			})

			It("Should return true", func() {
				Expect(configFound).To(Equal(true))
			})

		})

		Context("When given slice of one valid php config containing two valid keys", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{
					{
						Config: config.ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "newrelic.license",
									RawValue: "trueskis",
								},
							},
						},
					},
				}
			})

			It("Should return a slice of 1 validate element", func() {
				Expect(len(checkedConfigs)).To(Equal(1))
				Expect(checkedConfigs).To(Equal(input))
			})

			It("Should return true", func() {
				Expect(configFound).To(Equal(true))
			})

		})
		Context("When given slice of multiple valid php configs", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{
					{
						Config: config.ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
					{
						Config: config.ConfigElement{
							FileName: "newrelic.cfg",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "utilization.detect_docker",
							RawValue: "trueskis",
						},
					},
				}
			})

			It("Should return a slice of 2 validate elements", func() {
				Expect(len(checkedConfigs)).To(Equal(2))
				Expect(checkedConfigs).To(Equal(input))
			})

			It("Should return true", func() {
				Expect(configFound).To(Equal(true))
			})
		})
		Context("When given slice of invalid php configs", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{
					{
						Config: config.ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "name",
							RawValue: "trueskis",
						},
					},
					{
						Config: config.ConfigElement{
							FileName: "newrelic.yml",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
				}
			})

			It("Should return an empty slice of validate elements", func() {
				Expect(len(checkedConfigs)).To(Equal(0))
				Expect(checkedConfigs).To(Equal([]config.ValidateElement{}))
			})

			It("Should return false", func() {
				Expect(configFound).To(Equal(false))
			})
		})
		Context("When given slice of both invalid and valid php configs", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{
					{
						Config: config.ConfigElement{ //valid
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "utilization.detect_docker",
							RawValue: "trueskis",
						},
					},
					{
						Config: config.ConfigElement{ //valid
							FileName: "newrelic.cfg",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
					{
						Config: config.ConfigElement{ //invalid
							FileName: "newrelic.yml",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
				}
			})

			It("Should return a slice of 2  valid elements", func() {

				expectedResults := []config.ValidateElement{
					{
						Config: config.ConfigElement{ //valid
							FileName: "newrelic.ini",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "utilization.detect_docker",
							RawValue: "trueskis",
						},
					},
					{
						Config: config.ConfigElement{ //valid
							FileName: "newrelic.cfg",
							FilePath: "/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.enabled",
							RawValue: "trueskis",
						},
					},
				}
				Expect(len(checkedConfigs)).To(Equal(2))
				Expect(checkedConfigs).To(Equal(expectedResults))
			})

			It("Should return false", func() {
				Expect(configFound).To(Equal(true))
			})
		})
		Context("When given empty slice", func() {
			BeforeEach(func() {
				input = []config.ValidateElement{}
			})

			It("Should return an empty slice of validate elements", func() {
				Expect(len(checkedConfigs)).To(Equal(0))
				Expect(checkedConfigs).To(Equal([]config.ValidateElement{}))
			})

			It("Should return false", func() {
				Expect(configFound).To(Equal(false))
			})
		})

	})
})
