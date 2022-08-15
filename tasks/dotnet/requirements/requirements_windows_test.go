package requirements

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDotnetRequirements(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DotNet/Requirements/* test suite")
}
