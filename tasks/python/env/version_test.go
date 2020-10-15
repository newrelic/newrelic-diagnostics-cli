package env

import (
	"errors"
	"testing"

	"github.com/newrelic/NrDiag/tasks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Env Suite")
}

func mockCommandExecuteError(name string, arg ...string) ([]byte, error) {
	return []byte{}, errors.New("mock error")
}

func mockCommandExecuteSuccess(name string, arg ...string) ([]byte, error) {
	return []byte("Python 123"), nil
}

func mockCommandExecuteBadOutput(name string, arg ...string) ([]byte, error) {
	return []byte("Not Python 123"), nil
}

var _ = Describe("PythonEnvVersion", func() {
	p := PythonEnvVersion{}

	Describe("Identifier()", func() {
		It("Should return identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "Version", Category: "Python", Subcategory: "Env"}))
		})
	})

	Describe("Explain()", func() {
		It("should explain what the task does", func() {
			Expect(p.Explain()).To(Equal("Determine Python version"))
		})
	})

	Describe("Dependencies()", func() {
		It("should return list of dependencies' ", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Python/Config/Agent"}))
		})
	})

	Describe("Execute()", func() {
		Context("When the tasks dependencies were unsuccessfull", func() {
			It("should return a 'none' result.", func() {
				failedDependency := map[string]tasks.Result{
					"Python/Config/Agent": tasks.Result{Status: tasks.Failure},
				}
				Expect(p.Execute(tasks.Options{}, failedDependency).Status).To(Equal(tasks.None))
			})
		})

		Context("When 'python --version' command can not be run", func() {
			It("should return an error result", func() {
				successfullDependency := map[string]tasks.Result{
					"Python/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.cmdExec = mockCommandExecuteError
				Expect(p.Execute(tasks.Options{}, successfullDependency).Status).To(Equal(tasks.Error))
			})
		})

		Context("When 'python --version' runs successfully", func() {
			It("should return an info result when python version can be parsed", func() {
				successfullDependency := map[string]tasks.Result{
					"Python/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.cmdExec = mockCommandExecuteSuccess
				Expect(p.Execute(tasks.Options{}, successfullDependency).Status).To(Equal(tasks.Info))
			})
			It("should return an error result when python version can not be parsed", func() {
				successfullDependency := map[string]tasks.Result{
					"Python/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.cmdExec = mockCommandExecuteBadOutput
				Expect(p.Execute(tasks.Options{}, successfullDependency).Status).To(Equal(tasks.Error))
			})
		})
	})

	Describe("parsePythonVersion()", func() {
		It("should return ('123', true) when passed 'Python 123 '", func() {
			versionString, isValid := parsePythonVersion([]byte("Python 123 "))
			Expect(versionString).To(Equal("123"))
			Expect(isValid).To(BeTrue())
		})
		It("should return ('', false) when passed 'Pyrhon 123 '", func() {
			versionString, isValid := parsePythonVersion([]byte("Pyrhon 123 "))
			Expect(versionString).To(Equal(""))
			Expect(isValid).To(BeFalse())
		})
	})
})
