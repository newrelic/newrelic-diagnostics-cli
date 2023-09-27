//go:build linux || darwin
// +build linux darwin

package env

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base/Env/RootUser", func() {
	var p BaseEnvRootUser //instance of our task struct to be used in tests

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when isUserRoot is successful and user is root", func() {
			BeforeEach(func() {
				p.isUserRoot = func() (bool, error) { return true, nil }
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

		})

		Context("when isUserRoot is successful and user is not root", func() {
			BeforeEach(func() {
				p.isUserRoot = func() (bool, error) { return false, nil }
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

		})

		Context("when isUserRoot has an error", func() {
			BeforeEach(func() {
				p.isUserRoot = func() (bool, error) { return false, errors.New("test") }
			})

			It("should return an expected Error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

		})
	})
})
