package config

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BaseConfigValidateLicenseKey", func() {

	var p BaseConfigValidateLicenseKey

	Describe("Execute()", func() {

		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when LicenseKey upstream task did not find any license keys", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status:  tasks.Success,
						Payload: []LicenseKey{},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("No New Relic licenses keys were found. Task to validate license key did not run"))
			})
		})

		Context("when 1 license key is found and is the sample license key set in the PHP config file", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"REPLACE_WITH_REAL_KEY"`,
								Source: "/usr/local/etc/php/conf.d/newrelic.ini",
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal(`We validated 1 license key(s):` + "\n" + `The license key found in /usr/local/etc/php/conf.d/newrelic.ini does not have a valid format: REPLACE_WITH_REAL_KEY. ` + "\n" + `The NR license key is 40 alphanumeric characters. ` + "\n" + `Review this documentation to make sure that you have the proper format of a New Relic Personal API key: ` + "\n" + `https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys` + "\n\n"))

			})
		})

		Context("when 1 license key is found and is set in the ruby config file and is using the NRAL suffix", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"08a2ad66c637a29c3982469a3fe8d1982d00NRAL"`,
								Source: "/data/www/myappname-production/releases/20191009171653/config/newrelic.yml",
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					validLKToSources := make(map[string][]string)
					validLKToSources["08a2ad66c637a29c3982469a3fe8d1982d00NRAL"] = []string{"/data/www/myappname-production/releases/20191009171653/config/newrelic.yml"}

					return validLKToSources, map[string][]string{}, nil
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal(`We validated 1 license key(s):` + "\n" + `The license key found in /data/www/myappname-production/releases/20191009171653/config/newrelic.yml passed our validation check when verifying against your account:` + "\n" + ` 08a2ad66c637a29c3982469a3fe8d1982d00NRAL` + "\nNote: If your agent is reporting an 'Invalid license key' log entry for this valid License key, reach out to New Relic support to verify any issues in our end.\n\n"))

			})
		})

		Context("when 1 license key is found and is set in the python config file and is valid but we ran into an error when checking against account", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"08a2ad66c637a29c3982469a3fe8d1982d00NRAL"`,
								Source: "/app/myappname/newrelic.ini",
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {

					return map[string][]string{}, map[string][]string{}, errors.New(`Expected StatusCode < 300 got 500: {"success":false,"error":"Could not resolve license keys"`)
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\nThe license key found in\n/app/myappname/newrelic.ini\nhas a valid New Relic format: " + `"08a2ad66c637a29c3982469a3fe8d1982d00NRAL"` + "\n" + `Though we ran into an error (Expected StatusCode < 300 got 500: {"success":false,"error":"Could not resolve license keys") while trying to validate against your account. Only if your agent is reporting an 'Invalid license key' log entry, reach out to New Relic Support.` + "\n"))

			})
		})

		Context("when 2 license keys are found and is set in the dotnet/infra config file, it has valid format, but is invalid when checking against account(the account's owner rotated keys and the one being used is an old one)", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"eu01xx66c637a29c3982469a3fe8d1982d00NRAL"`,
								Source: `C:\ProgramData\New Relic\.NET Agent\newrelic.config`,
							},
							LicenseKey{
								Value:  `"eu01xx66c637a29c3982469a3fe8d1982d00NRAL"`,
								Source: `C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml`,
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					invalidLKToSources := make(map[string][]string)
					invalidLKToSources["eu01xx66c637a29c3982469a3fe8d1982d00NRAL"] = []string{`C:\ProgramData\New Relic\.NET Agent\newrelic.config`, `C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml`}
					return map[string][]string{}, invalidLKToSources, nil
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
				Expect(result.Summary).To(Equal(`We validated 1 license key(s):` + "\n" + `The license key found in C:\ProgramData\New Relic\.NET Agent\newrelic.config,` + "\n " + `C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml` + " did not match the one assigned to your account:\neu01xx66c637a29c3982469a3fe8d1982d00NRAL\nIf you are using an 'ingest key', ignore this warning. Ingest keys are secondary license keys manage by their own users that we do not validate for. Read more about ingest keys - https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-manage-license-keys-user-keys\n\n"))

			})
		})

		Context("when 2 unique license keys where found and one is valid and the other has a format issue", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  "08a2ad66c637a29c3982469a3fe8d1982d002c4a",
								Source: "/newrelic/newrelic.yml",
							},
							LicenseKey{
								Value:  "08a2ad66c637a29c3982469a3fe8d1982d002c4",
								Source: `C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml`,
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					validLKToSources := make(map[string][]string)
					validLKToSources["08a2ad66c637a29c3982469a3fe8d1982d002c4a"] = []string{"/newrelic/newrelic.yml"}
					return validLKToSources, map[string][]string{}, nil
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("We validated 2 license key(s):\n" + `The license key found in /newrelic/newrelic.yml passed our validation check when verifying against your account:` + "\n" + ` 08a2ad66c637a29c3982469a3fe8d1982d002c4a` + "\n" + "Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, reach out to New Relic support to verify any issues in our end.\n" + "\n" + `The license key found in C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml does not have a valid format: 08a2ad66c637a29c3982469a3fe8d1982d002c4. ` + "\n" + `The NR license key is 40 alphanumeric characters. ` + "\n" + `Review this documentation to make sure that you have the proper format of a New Relic Personal API key: ` + "\n" + `https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys` + "\n\n"))
			})
		})

		Context("when 2 license keys are found and one is an env var with quotes around it, and the other is a sample license key coming from a config file", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"08a2ad66c637a29c3982469a3fe8d1982d002c4a"`,
								Source: "NEW_RELIC_LICENSE_KEY",
							},
							LicenseKey{
								Value:  `"<%=license_key%>"`,
								Source: "/newrelic/newrelic.yml",
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal(`We validated 1 license key(s):` + "\nUsing quotes around NEW_RELIC_LICENSE_KEY may cause inconsistent behavior. We highly recommend removing those quotes, and running the " + tasks.ThisProgramFullName + " again.\n\n"))

			})
		})

		Context("when 4 license keys are found and 2 of them are valid env vars, and the other 2 are sample license key coming from a config file", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `08a2ad66c637a29c3982469a3fe8d1982d002c4f`,
								Source: "NEW_RELIC_LICENSE_KEY",
							},
							LicenseKey{
								Value:  `08a2ad66c637a29c3982469a3fe8d1982d002c4f`,
								Source: "NRIA_LICENSE_KEY",
							},
							LicenseKey{
								Value:  `"<%=license_key%>"`,
								Source: "/newrelic/newrelic.yml",
							},
							LicenseKey{
								Value:  "your_license_key",
								Source: `/var/log/newrelic-infra.yml`,
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					validLKToSources := make(map[string][]string)
					validLKToSources["08a2ad66c637a29c3982469a3fe8d1982d002c4a"] = []string{"NEW_RELIC_LICENSE_KEY", "NRIA_LICENSE_KEY"}
					return validLKToSources, map[string][]string{}, nil
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal(`We validated 2 license key(s):` + "\nThe license key found in NEW_RELIC_LICENSE_KEY,\n NRIA_LICENSE_KEY passed our validation check when verifying against your account:\n 08a2ad66c637a29c3982469a3fe8d1982d002c4a\n" + `Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, reach out to New Relic support to verify any issues in our end.` + "\n\n"))

			})
		})

		Context("when 1 license key is found and is a NR REST API key", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `"aaaaaa1a1a11a111111aa1111a11aa1a11aa111aaaa1a11"`,
								Source: "mynodeapp/newrelic.js",
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + `The license key found in mynodeapp/newrelic.js does not have a valid format: aaaaaa1a1a11a111111aa1111a11aa1a11aa111aaaa1a11. ` + "\n" + `The NR license key is 40 alphanumeric characters. ` + "\n" + `Review this documentation to make sure that you have the proper format of a New Relic Personal API key: ` + "\n" + `https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys` + "\n\n"))

			})
		})

		Context("when 1 license key is found and is not removing the sample syntax provided by the Java agent", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `'<%=c8f8ff84ed677d5791eeefb672a69447fb788486 %>'`,
								Source: "newrelic/newrelic.yml",
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + `The license key found in newrelic/newrelic.yml does not have a valid format: <%!=(MISSING)c8f8ff84ed677d5791eeefb672a69447fb788486%!>(MISSING). ` + "\n" + `The NR license key is 40 alphanumeric characters. ` + "\n" + `Review this documentation to make sure that you have the proper format of a New Relic Personal API key: ` + "\n" + `https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys` + "\n\n"))

			})
		})

		Context("when 1 license key is found and customer copy pasted the infra license key leaving one digit behind", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{

								Value:  `5306276ad40fb0c3caccba85f869dcadc018e54`,
								Source: `C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml`,
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + `The license key found in C:\Program Files\New Relic\newrelic-infra\newrelic-infra.yml does not have a valid format: 5306276ad40fb0c3caccba85f869dcadc018e54. ` + "\n" + `The NR license key is 40 alphanumeric characters. ` + "\n" + `Review this documentation to make sure that you have the proper format of a New Relic Personal API key: ` + "\n" + `https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys` + "\n\n"))

			})
		})

		Context("when 2 license keys are found and one of them is an env var with valid license key, and the other is a sample license key coming from a config file", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{

								Value:  `'<%=license_key%>'`,
								Source: "newrelic/newrelic.yml",
							},
							LicenseKey{
								Value:  `eu01xx66c637a29c3982469a3fe8d1982d00NRAL`,
								Source: "NEW_RELIC_LICENSE_KEY",
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					validLKToSources := make(map[string][]string)
					validLKToSources["eu01xx66c637a29c3982469a3fe8d1982d00NRAL"] = []string{"NEW_RELIC_LICENSE_KEY"}
					return validLKToSources, map[string][]string{}, nil
				}
			})
			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + `The license key found in NEW_RELIC_LICENSE_KEY passed our validation check when verifying against your account:` + "\n" + ` eu01xx66c637a29c3982469a3fe8d1982d00NRAL` + "\n" + `Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, please verify that your agent version is compatible with New Relic license keys that are 'region aware': https://docs.newrelic.com/docs/using-new-relic/welcome-new-relic/get-started/our-eu-us-region-data-centers. Reach out to Support if this is not the issue.` + "\n"))

			})

		})
		Context("when 2 license keys are found and it is an env var with a valid region license key and the other is the sample config license key", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{
								Value:  `eu01xx66c637a29c3982469a3fe8d1982d00NRAL`,
								Source: "NRIA_LICENSE_KEY",
							},
							LicenseKey{
								Value:  "your_license_key",
								Source: `/var/log/newrelic-infra.yml`,
							},
						},
					},
				}
				p.validateAgainstAccount = func(map[string][]string) (map[string][]string, map[string][]string, error) {
					validLKToSources := make(map[string][]string)
					validLKToSources["eu01xx66c637a29c3982469a3fe8d1982d00NRAL"] = []string{"NRIA_LICENSE_KEY"}
					return validLKToSources, map[string][]string{}, nil
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + `The license key found in NRIA_LICENSE_KEY passed our validation check when verifying against your account:` + "\n" + ` eu01xx66c637a29c3982469a3fe8d1982d00NRAL` + "\n" + `Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, please verify that your agent version is compatible with New Relic license keys that are 'region aware': https://docs.newrelic.com/docs/using-new-relic/welcome-new-relic/get-started/our-eu-us-region-data-centers. Reach out to Support if this is not the issue.` + "\n"))

			})
		})

		Context("when 2 license keys are found and it is an env var with a invalid format and the other has removed the sample config license key", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/LicenseKey": tasks.Result{
						Status: tasks.Success,
						Payload: []LicenseKey{
							LicenseKey{

								Value:  `x692c6460dc93f7c586a1bd1a6a98f030cdaf4498785150`,
								Source: "NEW_RELIC_LICENSE_KEY",
							},
							LicenseKey{
								Value:  "",
								Source: `myapp/newrelic.js`,
							},
						},
					},
				}
			})

			It("Should return a None status and summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("We validated 1 license key(s):\n" + "The license key found in NEW_RELIC_LICENSE_KEY does not have a valid format: x692c6460dc93f7c586a1bd1a6a98f030cdaf4498785150. \nThe NR license key is 40 alphanumeric characters. \nReview this documentation to make sure that you have the proper format of a New Relic Personal API key: \nhttps://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys" + "\n\n"))
			})
		})
	})
})
