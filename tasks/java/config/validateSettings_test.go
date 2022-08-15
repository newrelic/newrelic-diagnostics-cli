package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func BeValid() types.GomegaMatcher {
	return ResultMatcher{Valid}
}

type ResultMatcher struct {
	Status ValidationStatus
}

func (r ResultMatcher) Match(actual interface{}) (bool, error) {
	result, ok := actual.(ValidationResult)
	if !ok {
		return false, errors.New("need a ValidationResult object")
	}
	return result.Status == r.Status, nil
}

func (r ResultMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected a %s result (got %s)", r.Status, actual.(ValidationResult).Status)
}
func (r ResultMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Did not expect result to be %s (expected %s)", actual.(ValidationResult).Status, r.Status)
}

func ExpectValidator(kind string) func(interface{}) Assertion {
	return func(value interface{}) Assertion { return Expect(ValidateSetting(value, kind)) }
}

func TestValidateSettings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Java/Config/ValidateSettings test suite")
}

var YamlDictionary = map[interface{}]interface{}{"foo": "bar"}

var _ = Describe("Java/Config/ValidateSettings", func() {
	Describe("ValidateString", func() {
		ExpectString := ExpectValidator("String")
		It("Should always be valid", func() {
			ExpectString("whatever").To(BeValid())
			ExpectString(972).To(BeValid())
			ExpectString(nil).To(BeValid())
		})
	})
	Describe("ValidateAppName", func() {
		ExpectAppName := ExpectValidator("AppName")
		It("Should fail with more than two ;", func() {
			ExpectAppName("foo;bar;baz;quux").ToNot(BeValid())
		})
		It("Should succeed with one or two ;", func() {
			ExpectAppName("foo;bar").To(BeValid())
			ExpectAppName("foo;bar;baz").To(BeValid())
		})
		It("Should fail if app name is a number", func() {
			ExpectAppName(3.1415).ToNot(BeValid())
			ExpectAppName(3).ToNot(BeValid())
		})
		It("Should fail if app name is empty", func() {
			ExpectAppName(nil).ToNot(BeValid())
			ExpectAppName("").ToNot(BeValid())
		})
	})
	Describe("ValidateLabelList", func() {
		ExpectLabelList := ExpectValidator("LabelList")
		It("Should fail with no :", func() {
			ExpectLabelList("foo").ToNot(BeValid())
		})
		It("Should fail with more than two :", func() {
			ExpectLabelList("foo:bar:baz").ToNot(BeValid())
		})
		It("Should accept empty", func() {
			ExpectLabelList(nil).To(BeValid())
		})
		It("Should accept multiple labels", func() {
			ExpectLabelList("foo:bar;baz:quux").To(BeValid())
		})
		It("Should not accept empty labels", func() {
			ExpectLabelList("foo:bar;").ToNot(BeValid())
			ExpectLabelList(";").ToNot(BeValid())
			ExpectLabelList("foo:bar;;baz:quux").ToNot(BeValid())
		})
		It("Should not accept half-empty labels", func() {
			ExpectLabelList("foo:").ToNot(BeValid())
			ExpectLabelList(":bar").ToNot(BeValid())
		})
		It("Should accept a sub-dictionary", func() {
			ExpectLabelList(YamlDictionary).To(BeValid())
		})
	})
	Describe("ValidateProxyScheme", func() {
		ExpectProxyScheme := ExpectValidator("ProxyScheme")
		It("Should accept http and https", func() {
			ExpectProxyScheme("http").To(BeValid())
			ExpectProxyScheme("https").To(BeValid())
			ExpectProxyScheme(nil).To(BeValid())
		})
		It("Should not accept anything else", func() {
			ExpectProxyScheme(972).ToNot(BeValid())
			ExpectProxyScheme("Horseshoes").ToNot(BeValid())
		})
	})
	Describe("ValidateLogLevel", func() {
		ExpectLogLevel := ExpectValidator("LogLevel")
		It("Should accept all the good values", func() {
			ExpectLogLevel("off").To(BeValid())
			ExpectLogLevel("severe").To(BeValid())
			ExpectLogLevel("warning").To(BeValid())
			ExpectLogLevel("info").To(BeValid())
			ExpectLogLevel("fine").To(BeValid())
			ExpectLogLevel("finer").To(BeValid())
			ExpectLogLevel("finest").To(BeValid())
			ExpectLogLevel(nil).To(BeValid())
		})
		It("Should not accept anything else", func() {
			ExpectLogLevel("Hedgehog").ToNot(BeValid())
			ExpectLogLevel(0).ToNot(BeValid())
			ExpectLogLevel(YamlDictionary).ToNot(BeValid())
			ExpectLogLevel(true).ToNot(BeValid())

		})
	})
	Describe("ValidateRecordSql", func() {
		ExpectRecordSql := ExpectValidator("RecordSql")
		It("Should accept the good values", func() {
			ExpectRecordSql("off").To(BeValid())
			ExpectRecordSql("raw").To(BeValid())
			ExpectRecordSql("obfuscated").To(BeValid())
			ExpectRecordSql(nil).To(BeValid())
		})
		It("Should not accept anything else", func() {
			ExpectRecordSql(true).ToNot(BeValid())
			ExpectRecordSql("Pants").ToNot(BeValid())
			ExpectRecordSql(YamlDictionary).ToNot(BeValid())
		})
	})
	Describe("ValidateTransactionThreshold", func() {
		ExpectRecordSql := ExpectValidator("TransactionThreshold")
		It("Should accept numbers", func() {
			ExpectRecordSql(1).To(BeValid())
			ExpectRecordSql(1.5).To(BeValid())
			ExpectRecordSql(-92.1749).To(BeValid())
		})
		It("Should accept apdex_f", func() {
			ExpectRecordSql("apdex_f").To(BeValid())
		})
		It("Should accept nil", func() {
			ExpectRecordSql(nil).To(BeValid())
		})
		It("Should not accept anything else", func() {
			ExpectRecordSql(true).ToNot(BeValid())
			ExpectRecordSql("Unacceptable").ToNot(BeValid())
			ExpectRecordSql(YamlDictionary).ToNot(BeValid())
		})
	})
	Describe("ValidateStatusCodeList", func() {
		ExpectStatusCodeList := ExpectValidator("StatusCodeList")
		It("Should accept nil", func() {
			ExpectStatusCodeList(nil).To(BeValid())
		})
		It("Should accept numbers within range", func() {
			ExpectStatusCodeList(500).To(BeValid())
			ExpectStatusCodeList(0).To(BeValid())
			ExpectStatusCodeList(1000).To(BeValid())
		})
		It("Should not accept numbers out of range", func() {
			ExpectStatusCodeList(-1).ToNot(BeValid())
			ExpectStatusCodeList(1001).ToNot(BeValid())
		})
		It("Should accept proper ranges", func() {
			ExpectStatusCodeList("100-200").To(BeValid())
			ExpectStatusCodeList("0-1000").To(BeValid())
		})
		It("Should not accept ranges with out of bounds numbers", func() {
			ExpectStatusCodeList("100-2000").ToNot(BeValid())
			ExpectStatusCodeList("-234-1000").ToNot(BeValid())
		})
		It("Should not accept backwards ranges", func() {
			ExpectStatusCodeList("500-100").ToNot(BeValid())
		})
		It("Should not accept weird things", func() {
			ExpectStatusCodeList(true).ToNot(BeValid())
			ExpectStatusCodeList("Horseshoes").ToNot(BeValid())
			ExpectStatusCodeList(YamlDictionary).ToNot(BeValid())
		})
	})
})
