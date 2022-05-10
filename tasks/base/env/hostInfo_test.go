package env

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Base/Env/HostInfo", func() {

	var p BaseEnvHostInfo

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Env",
				Name:        "HostInfo",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Collect host system info"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return a slice ", func() {
			expectedDependencies := []string{}
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

		Context("If hostInfo returns an error", func() {
			BeforeEach(func() {
				options = tasks.Options{}

				p.HostInfoProvider = func() (HostInfo, error) {
					hostInfo := HostInfo{}
					return hostInfo, errors.New("host is haunted")
				}
				p.HostInfoProviderWithContext = func(context.Context) (HostInfo, error) {
					hostInfo := HostInfo{}
					return hostInfo, errors.New("host is haunted")
				}
			})
			It("Should return", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return", func() {
				Expect(result.Summary).To(Equal("Error collecting complete host information:\nhost is haunted"))
			})

		})

		Context("If host info collection is successful", func() {
			BeforeEach(func() {
				options = tasks.Options{}

				p.HostInfoProvider = func() (HostInfo, error) {
					hostInfo := HostInfo{}
					return hostInfo, nil
				}
				p.HostInfoProviderWithContext = func(context.Context) (HostInfo, error) {
					hostInfo := HostInfo{}
					return hostInfo, nil
				}
			})
			It("Should return", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("Should return", func() {
				Expect(result.Summary).To(Equal("Collected host information"))
			})

		})

	})
})
