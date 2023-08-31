package config

// Tests for Infra/Config/IntegrationsCollect

import (
	"github.com/newrelic/newrelic-diagnostics-cli/internal/haberdasher"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base/Config/ValidateHSM", func() {
	var p BaseConfigValidateHSM

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Config",
				Name:        "ValidateHSM",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Validate High Security Mode agent configuration against account configuration"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Base/Config/ValidateLicenseKey", "Base/Config/Validate", "Base/Env/CollectEnvVars"}
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

		Context("When provided multiple license keys that are hsm true and local config matches", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.js",
									FilePath: "/app/",
								},
							},
						},
					},
					"Base/Config/ValidateLicenseKey": {
						Status: tasks.Success,
						Payload: map[string][]string{
							"Banana": {"/etc/newrelic.yml"},
							"Peel":   {"/app/newrelic.js"},
						},
					},
				}

				p.hsmService = func(licenseKeys []string) ([]haberdasher.HSMresult, *haberdasher.Response, error) {
					results := []haberdasher.HSMresult{}

					for _, lk := range licenseKeys {
						result := haberdasher.HSMresult{
							LicenseKey: lk,
							IsEnabled:  true,
						}
						results = append(results, result)
					}

					return results, &haberdasher.Response{}, nil
				}

			})

			expectedResult := tasks.Result{
				Status:  tasks.Success,
				Summary: "High Security Mode setting for accounts associated with found license keys match local configuration.",
				Payload: []HSMvalidation{
					{
						LicenseKey: "Banana",
						AccountHSM: true,
						LocalHSM: map[string]bool{
							"/etc/newrelic.yml": true,
						},
					},
					{
						LicenseKey: "Peel",
						AccountHSM: true,
						LocalHSM: map[string]bool{
							"/app/newrelic.js": true,
						},
					},
				},
			}

			It("Should return a success status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("It should return a successful summary", func() {
				Expect(result.Summary).To(Equal(expectedResult.Summary))
			})

			It("It should return a payload with an hsmValidation for each key", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})
		})

		Context("When provided multiple license keys and one has a matching account and local hsm state, and the other does not", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "false",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.js",
									FilePath: "/app/",
								},
							},
						},
					},
					"Base/Config/ValidateLicenseKey": {
						Status: tasks.Success,
						Payload: map[string][]string{
							"Banana": {"/etc/newrelic.yml"},
							"Peel":   {"/app/newrelic.js"},
						},
					},
				}

				p.hsmService = func(licenseKeys []string) ([]haberdasher.HSMresult, *haberdasher.Response, error) {
					results := []haberdasher.HSMresult{}

					for _, lk := range licenseKeys {
						result := haberdasher.HSMresult{
							LicenseKey: lk,
							IsEnabled:  true,
						}
						results = append(results, result)
					}

					return results, &haberdasher.Response{}, nil
				}

			})

			expectedResult := tasks.Result{
				Status:  tasks.Failure,
				Summary: "High Security Mode setting (true) for account with license key:\n\nBanana\n\nmismatches configuration in:\n/etc/newrelic.yml\n\n",
				Payload: []HSMvalidation{
					{
						LicenseKey: "Banana",
						AccountHSM: true,
						LocalHSM: map[string]bool{
							"/etc/newrelic.yml": false,
						},
					},
					{
						LicenseKey: "Peel",
						AccountHSM: true,
						LocalHSM: map[string]bool{
							"/app/newrelic.js": true,
						},
					},
				},
			}

			It("Should return a failure status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("It should return a failure summary", func() {
				Expect(result.Summary).To(Equal(expectedResult.Summary))
			})

			It("It should return a payload with an hsmValidation for each licenseKey", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})
		})

		Context("When provided a license key that has multiple sources but one source does not match account hsm", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "false",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.js",
									FilePath: "/app/",
								},
							},
						},
					},
					"Base/Config/ValidateLicenseKey": {
						Status: tasks.Success,
						Payload: map[string][]string{
							"Banana": {"/etc/newrelic.yml", "/app/newrelic.js", "NEW_RELIC_LICENSE_KEY"},
						},
					},
					"Base/Env/CollectEnvVars": {
						Status: tasks.Success,
						Payload: map[string]string{
							"NEW_RELIC_HIGH_SECURITY": "true",
						},
					},
				}

				p.hsmService = func(licenseKeys []string) ([]haberdasher.HSMresult, *haberdasher.Response, error) {
					results := []haberdasher.HSMresult{}

					for _, lk := range licenseKeys {
						result := haberdasher.HSMresult{
							LicenseKey: lk,
							IsEnabled:  false,
						}
						results = append(results, result)
					}

					return results, &haberdasher.Response{}, nil
				}

			})

			expectedResult := tasks.Result{
				Status:  tasks.Failure,
				Summary: "High Security Mode setting (false) for account with license key:\n\nBanana\n\nmismatches configuration in:\n/app/newrelic.js\nNEW_RELIC_HIGH_SECURITY\n\n",
				Payload: []HSMvalidation{
					{
						LicenseKey: "Banana",
						AccountHSM: false,
						LocalHSM: map[string]bool{
							"/etc/newrelic.yml": false,
							"/app/newrelic.js":  true,
						},
					},
				},
			}

			It("Should return a failure status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("It should return a failure summary", func() {
				Expect(result.Summary).To(ContainSubstring("High Security Mode setting (false) for account with license key:\n\nBanana\n\nmismatches configuration in:"))
				Expect(result.Summary).To(ContainSubstring("\n/app/newrelic.js"))
				Expect(result.Summary).To(ContainSubstring("\nNEW_RELIC_HIGH_SECURITY"))
			})

			It("It should return a payload with an hsmValidation with expected payload", func() {
				Expect(result.Payload.([]HSMvalidation)[0].LicenseKey).To(Equal(expectedResult.Payload.([]HSMvalidation)[0].LicenseKey))
				Expect(result.Payload.([]HSMvalidation)[0].AccountHSM).To(Equal(expectedResult.Payload.([]HSMvalidation)[0].AccountHSM))
				Expect(result.Payload.([]HSMvalidation)[0].LocalHSM).To(HaveKeyWithValue("/etc/newrelic.yml", false))
				Expect(result.Payload.([]HSMvalidation)[0].LocalHSM).To(HaveKeyWithValue("/app/newrelic.js", true))
				Expect(result.Payload.([]HSMvalidation)[0].LocalHSM).To(HaveKeyWithValue("NEW_RELIC_HIGH_SECURITY", true))
			})

		})

	})
})
