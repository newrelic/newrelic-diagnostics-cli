package log

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

func TestBaseLog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Log/* test suites")
}

var _ = Describe("Base/Log/ReportingTo", func() {
	var p BaseLogReportingTo

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When upstream provides an empty payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status:  tasks.Success,
						Payload: []LogElement{},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Logs not found"))
			})

		})

		Context("When log file does not contain a reporting to line", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
						Payload: []LogElement{
							{
								FileName: "reportingTo_empty.log",
								FilePath: "./fixtures/",
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Logs founds, no instances of reporting to within logs."))
			})

		})

		Context("When a single instance of reporting to is found as plain text", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
						Payload: []LogElement{
							{
								FileName: "reportingTo_plain.log",
								FilePath: "./fixtures/",
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found a reporting to."))
			})
		})

		Context("When a single instance of reporting to is found as JSON", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
						Payload: []LogElement{
							{
								FileName: "reportingTo_json.log",
								FilePath: "./fixtures/",
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found a reporting to."))
			})

			It("should trim ReportingTo attribute to only the URL", func() {
				payload := result.Payload.([]LogNameReportingTo)
				Expect(payload[0].ReportingTo[0]).To(Equal("https://rpm.newrelic.com/accounts/111/applications/222"))
			})
		})

		Context("When multiple reporting to lines are found in multiple log files", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
						Payload: []LogElement{
							{
								FileName: "reportingTo_plain.log",
								FilePath: "./fixtures/",
							},
							{
								FileName: "reportingTo_json.log",
								FilePath: "./fixtures/",
							},
							{
								FileName: "reportingTo_empty.log",
								FilePath: "./fixtures/",
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found a reporting to."))
			})

			It("should only include log files with found reporting to line, in payload", func() {
				payload := result.Payload.([]LogNameReportingTo)
				Expect(len(payload)).To(Equal(2))
			})

			It("should trim ReportingTo attribute to only the URL", func() {
				payload := result.Payload.([]LogNameReportingTo)
				Expect(payload[0].ReportingTo[0]).To(Equal("https://rpm.newrelic.com/accounts/111/applications/222"))
			})
		})

	})

})
