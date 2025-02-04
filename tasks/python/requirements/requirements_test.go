package requirements

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRequirements(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Python/Requirements Suite")
}
