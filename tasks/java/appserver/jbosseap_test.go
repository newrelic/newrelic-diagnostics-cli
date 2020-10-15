package appserver

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tasks "github.com/newrelic/NrDiag/tasks"
)

//this mock is used for testing the listDir function to create a mock os.FileInfo object
type mockFileInfo struct {
	name string
}

func (m mockFileInfo) Name() string {
	return m.name
}
func (m mockFileInfo) Size() int64 {
	return 0
}
func (m mockFileInfo) Mode() os.FileMode {
	return 0
}
func (m mockFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (m mockFileInfo) IsDir() bool {
	return true
}
func (m mockFileInfo) Sys() interface{} {
	return nil
}



func TestVersionCheckJboss(t *testing.T) {

	JbossVersionsTests := []struct {
		version string
		want    tasks.Status
	}{
		{version: "7.0", want: tasks.Failure},
		{version: "6.2", want: tasks.Success},
		{version: "5.2", want: tasks.Failure},
		{version: "2.11", want: tasks.Failure},
		{version: "not valid version", want: tasks.Error},
		{version: "7.11.2.GA", want: tasks.Failure},
		{version: "6.1.2.GA", want: tasks.Success},
	}

	for _, JbossVersionsTest := range JbossVersionsTests {

		jbossCheckStatus, _ := versionCheckJboss(JbossVersionsTest.version)
		if jbossCheckStatus != JbossVersionsTest.want {

			t.Errorf("Test failed with version %v. Had %v wanted %v", JbossVersionsTest.version, jbossCheckStatus, JbossVersionsTest.want)
		}
	}

}

var _ = Describe("JavaAppserverJbossEapCheck", func() {

	var p JavaAppserverJbossEapCheck

	Describe("Identifier", func() {
		It("Should return the identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "JbossEapCheck", Category: "Java", Subcategory: "Appserver"}))
		})
	})
	Describe("Explain", func() {
		It("Should return explain", func() {
			Expect(p.Explain()).To(Equal("Check JBoss EAP version compatibility with New Relic Java agent"))
		})
	})
	Describe("Dependencies", func() {
		It("Should return dependencies list", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Java/Config/Agent", "Base/Env/CollectEnvVars"}))
		})
	})
	Describe("Execute", func() {
		var (
			upstream map[string]tasks.Result
			result   tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(tasks.Options{}, upstream)
		})

		Context("When Java/Config/Agent was not successful", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{"Java/Config/Agent": tasks.Result{Status: tasks.None}}
			})
			It("Should return none result", func() {
				Expect(result.Status).To(Equal(tasks.None))

			})
			It("Should return none Summary", func() {
				Expect(result.Summary).To(Equal("Java config file not detected, this task didn't run"))
			})
		})
		Context("When JBOSS_HOME ENV var set and version file doesn't exist", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"JBOSS_HOME": "/foo"}},
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("Version file didn't exist")
				}
				p.fileExists = func(string) bool {
					return true
				}

			})
			It("Should return Result from getAndParseJBossAsReadMeChecker", func() {
				Expect(result.Summary).To(Equal("JBossEAP detected but unable to detect version: Version file not found"))
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
		})
		Context("When JBOSS_HOME ENV var set and error reading version file", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"JBOSS_HOME": "/foo"}},
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("There they are all sitting in a row")
				}
				p.fileExists = func(string) bool {
					return true
				}

			})
			It("Should return Result from getAndParseJBossAsReadMeChecker", func() {
				Expect(result.Summary).To(Equal("JBossEAP detected but unable to detect version: Error reading version file"))
			})
			It("Should return expected task status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})

		Context("When JBOSS_HOME ENV var set and version file exists", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"JBOSS_HOME": "/foo"}},
				}
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(search string, _ string) ([]string, error) {
					return []string{"", "6.0"}, nil
				}
			})
			It("Should return a success Status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return a success Summary", func() {
				Expect(result.Summary).To(Equal("Jboss EAP version is compatible. Version is 6.0"))
			})
		})
		Context("When JBOSS_HOME ENV var not set and on windows with valid version", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"Program_files": filepath.FromSlash("fixtures/eapFiles")}},
				}
				p.runtimeOs = "windows"
				p.fileExists = func(string) bool {
					return true
				}
				p.listDir = func(directory string, _ readDirFunc) []string {
					if directory == filepath.FromSlash("fixtures/eapFiles") {
						return []string{"EAP-5"}
					} else if directory == filepath.FromSlash("fixtures/eapFiles/EAP-5") {
						return []string{"jboss-eap-5"}
					}
					return []string{}
				}
				p.returnSubstringInFile = func(search string, _ string) ([]string, error) {
					return []string{"", "6.1"}, nil
				}
			})
			It("Should return success result Summary", func() {
				Expect(result.Summary).To(Equal("Jboss EAP version is compatible. Version is 6.1"))
			})
			It("Should return success result Status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
		})
		Context("When JBOSS_HOME ENV var not set and on linux with valid version", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.runtimeOs = "linux"
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(search string, _ string) ([]string, error) {
					if search == "JBOSS_HOME=.*" {
						return []string{`JBOSS_HOME="/foo/bar"`, ""}, nil
					} else if search == "Enterprise Application Platform - Version ([0-9.]+.*)" {
						return []string{"", "6.2"}, nil
					}
					return []string{}, errors.New("I shouldn't get here")
				}
			})
			It("Should return success result Summary", func() {
				Expect(result.Summary).To(Equal("Jboss EAP version is compatible. Version is 6.2"))
			})
			It("Should return success result Status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
		})
		Context("When JBOSS_HOME ENV var not set and on unrecognized OS", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.runtimeOs = "chocolate"
			})
			It("Should return none result Summary", func() {
				Expect(result.Summary).To(Equal("OS not detected as Linux or Windows, not running."))
			})
			It("Should return none result Status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
		})
		Context("When JBOSS_HOME is not set and error getting version file", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{Status: tasks.Success},
				}
				p.runtimeOs = "linux"
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("I've got a lovely bunch of coconuts")
				}
				p.listDir = func(directory string, _ readDirFunc) []string {
					if directory == filepath.FromSlash("fixtures/eapFiles") {
						return []string{"EAP-5"}
					} else if directory == filepath.FromSlash("fixtures/eapFiles/EAP-5") {
						return []string{"jboss-eap-5"}
					}
					return []string{}
				}
				p.fileExists = func(string) bool {
					return true
				}
			})
			It("Should return success result Summary", func() {
				Expect(result.Summary).To(Equal("Error getting jboss version file: file doesn't contain JBOSS_HOME"))
			})
			It("Should return success result Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
		Context("When JBOSS_HOME is not set and version file doesn't exist", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"Program_files": "foo/eapFiles"}},
				}
				p.runtimeOs = "windows"
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("Version file didn't exist")
				}
				p.listDir = func(directory string, _ readDirFunc) []string {
					return []string{"foo"}
				}
				p.fileExists = func(string) bool {
					return false
				}
			})
			It("Should return success result Summary", func() {
				Expect(result.Summary).To(Equal("Can't find version file, didn't exist"))
			})
			It("Should return success result Status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
		})

		Context("When JBOSS_HOME is not set and error reading version file", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Java/Config/Agent":       tasks.Result{Status: tasks.Success},
					"Base/Env/CollectEnvVars": tasks.Result{Payload: map[string]string{"Program_files": filepath.FromSlash("foo/eapFiles")}},
				}
				p.runtimeOs = "windows"
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("I am the black knight")
				}
				p.listDir = func(directory string, _ readDirFunc) []string {
					if directory == filepath.FromSlash("foo/eapFiles") {
						return []string{"EAP-5"}
					} else if directory == filepath.FromSlash("foo/eapFiles/EAP-5") {
						return []string{"jboss-eap-5"}
					}
					return []string{}
				}
				p.fileExists = func(string) bool {
					return true
				}
			})
			It("Should return success result Summary", func() {
				Expect(result.Summary).To(Equal("JBossEAP detected but unable to detect version: Error reading version file"))
			})
			It("Should return success result Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
	})

	Describe("getJbossFileWindows", func() {
		var (
			programFiles string
			result       string
			resultErr    error
		)
		JustBeforeEach(func() {
			result, resultErr = p.getJbossFileWindows(programFiles)
		})
		Context("When unable to list the directory", func() {
			BeforeEach(func() {
				programFiles = `C:\foo\bar`
				p.listDir = func(string, readDirFunc) []string {
					return []string{}
				}
			})
			It("Should return empty result", func() {
				Expect(result).To(Equal(""))
			})
			It("Should return err", func() {
				Expect(resultErr.Error()).To(Equal("unable to list directory"))
			})
		})
		Context("When filepaths found", func() {
			BeforeEach(func() {
				programFiles = filepath.FromSlash("fixtures/eapFiles")
				p.listDir = func(directory string, _ readDirFunc) []string {
					if directory == filepath.FromSlash("fixtures/eapFiles") {
						return []string{"EAP-5"}
					} else if directory == filepath.FromSlash("fixtures/eapFiles/EAP-5") {
						return []string{"jboss-eap-5"}
					}
					return []string{}
				}

			})
			It("Should return matching file", func() {
				Expect(result).To(Equal(filepath.FromSlash("fixtures/eapFiles/EAP-5/jboss-eap-5/version.txt")))
			})
			It("Should return nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("When program files not set", func() {
			BeforeEach(func() {
				programFiles = ""
			})
			It("Should return empty result", func() {
				Expect(result).To(Equal(""))
			})
			It("Should return err", func() {
				Expect(resultErr.Error()).To(Equal("Unable to detect JBoss directory"))
			})
		})
	})

	Describe("listDir", func() {
		var (
			directory           string
			returnedDirectories []string
			mockReadDir         readDirFunc
		)

		JustBeforeEach(func() {
			returnedDirectories = listDir(directory, mockReadDir)
		})
		Context("When unable to list the directory", func() {
			BeforeEach(func() {
				mockReadDir = func(string) ([]os.FileInfo, error) {
					return []os.FileInfo{}, errors.New("Totally tubular error here")
				}
			})
			It("Should return empty result", func() {
				Expect(returnedDirectories).To(Equal([]string{}))
			})
		})
		Context("When listing the directory", func() {
			BeforeEach(func() {
				mockReadDir = func(string) ([]os.FileInfo, error) {
					return []os.FileInfo{mockFileInfo{name: "foo"}, mockFileInfo{name: "bar"}}, nil
				}
			})
			It("Should return slice of paths", func() {
				Expect(returnedDirectories).To(Equal([]string{"foo", "bar"}))
			})
		})
	})

	Describe("readJbossVersionFile", func() {
		var (
			versionFilePath string
			result          string
			resultErr       error
		)
		JustBeforeEach(func() {
			versionFilePath = ""
			result, resultErr = p.readJbossVersionFile(versionFilePath)
		})
		Context("When versionFilePath doesn't exist", func() {
			BeforeEach(func() {

				p.fileExists = func(string) bool {
					return false
				}
			})
			It("Should return empty result", func() {
				Expect(result).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(resultErr.Error()).To(Equal("Version file didn't exist"))
			})
		})
		Context("When versionFilePath exists and an error is returned reading file", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("not even a good error message")
				}

			})
			It("Should return empty result", func() {
				Expect(result).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(resultErr.Error()).To(Equal("not even a good error message"))
			})
		})
		Context("When versionFilePath exists and file is read correctly", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"blarg!", "more stuff"}, nil
				}
			})
			It("Should return result", func() {
				Expect(result).To(Equal("more stuff"))
			})
			It("Should return expected error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
	})

	Describe("getJbossFileLinux", func() {
		var (
			file      string
			fileError error
		)
		JustBeforeEach(func() {
			file, fileError = p.getJbossFileLinux()
		})
		Context("When default jboss file exists and file doesn't container JBOSS_HOME", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return true
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("not even a good error message")
				}
			})
			It("Should return empty file", func() {
				Expect(file).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(fileError.Error()).To(Equal("file doesn't contain JBOSS_HOME"))
			})

		})
		Context("When default jboss file exists and error parsing JBOSS_HOME ", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"Bananas"}, nil
				}
			})
			It("Should return empty file", func() {
				Expect(file).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(fileError.Error()).To(Equal("Error parsing JBOSS_HOME path in jboss-eap.conf"))
			})

		})
		Context("When default jboss file exists and is parsed properly", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return true
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{`JBOSS_HOME="/app"`}, nil
				}
			})
			It("Should return expected file", func() {
				Expect(file).To(Equal(filepath.FromSlash("/app/version.txt")))
			})
			It("Should return nil error", func() {
				Expect(fileError).To(BeNil())
			})

		})
		Context("When default jboss 7 file exists", func() {
			BeforeEach(func() {
				p.fileExists = func(file string) bool {
					if file == "/etc/default/jboss-eap.conf" {
						return false
					}
					return true
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{`JBOSS_HOME="/app"`}, nil
				}
			})
			It("Should return jboss 7 file", func() {
				Expect(file).To(Equal("/opt/rh/eap7/root/usr/share/wildfly/version.txt"))
			})
			It("Should return nil error", func() {
				Expect(fileError).To(BeNil())
			})

		})
		Context("When expected files are not found anywhere", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return false
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
			})
			It("Should return empty file", func() {
				Expect(file).To(Equal(""))
			})
			It("Should return an error", func() {
				Expect(fileError.Error()).To(Equal("Unable to detect JBoss directory"))
			})

		})
		Context("When expected files are found but error parsing JBOSS_HOME", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return false
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"not here dude"}, nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"awesomeist_file"}
				}
			})
			It("Should return empty file", func() {
				Expect(file).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(fileError.Error()).To(Equal("error parsing JBOSS_HOME path in jboss-eap.conf"))
			})

		})
		Context("When expected files are found and JBOSS_HOME parsed ", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return false
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{`JBOSS_HOME="app/stuff/awesome`}, nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"awesomeist_file"}
				}
			})
			It("Should return expected file", func() {
				Expect(file).To(Equal(filepath.FromSlash("app/stuff/awesome/version.txt")))
			})
			It("Should return nil error", func() {
				Expect(fileError).To(BeNil())
			})

		})
		Context("When fallthrough occurs ", func() {
			BeforeEach(func() {
				p.fileExists = func(string) bool {
					return false
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("This is bad")
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"awesomeist_file"}
				}
			})
			It("Should return empty file", func() {
				Expect(file).To(Equal(""))
			})
			It("Should return expected error", func() {
				Expect(fileError.Error()).To(Equal("Unable to detect JBoss directory"))
			})

		})
	})

})
