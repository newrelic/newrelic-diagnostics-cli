//go:build integration
// +build integration

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type testResult struct {
	name string
	err  error
}

type integrationTest struct {
	Name                string   `yaml:"test_name,omitempty"`
	DockerfileLines     []string `yaml:"dockerfile_lines,omitempty"`
	DockerCMD           string   `yaml:"docker_cmd,omitempty"`
	DockerFROM          string   `yaml:"docker_from,omitempty"`
	HostsFileAdditions  []string `yaml:"hosts_file_additions,omitempty"`
	LogEntryExpected    []string `yaml:"log_entry_expected,omitempty"`
	LogEntryNotExpected []string `yaml:"log_entry_not_expected,omitempty"`
	ZipFileExpected     []string `yaml:"zip_file_expected,omitempty"`
	ZipFileNotExpected  []string `yaml:"zip_file_not_expected,omitempty"`
	TestFile            string
}

var testNameArgs = flag.String("testNames", "", "Comma separated list of regexes. Run only integration tests matching these regexes")

func TestRunIntegrationTests(t *testing.T) {

	//Setup for recording integration test timing
	testTimingsChannel := make(chan IntegrationTestRun)
	var wgTestTiming sync.WaitGroup
	wgTestTiming.Add(1)
	go RecordTestTimings(testTimingsChannel, &wgTestTiming)

	//Parse tests to run
	allTests := parseTestsFromYML()
	var testQueue []integrationTest

	//If user provided test names
	if len(*testNameArgs) > 0 {
		testQueue = filterTestsFromArgs(*testNameArgs, allTests)
		if len(testQueue) == 0 {
			t.Fatalf("No tests found matching provided name: %s", *testNameArgs)
		}
	} else {
		testQueue = allTests
	}

	//Run the tests
	fmt.Printf("Running %d integration test(s) ...\n", len(testQueue))

	workerPoolSize := 5
	testQueueChan := make(chan integrationTest)

	var wgTestWorker sync.WaitGroup
	wgTestWorker.Add(workerPoolSize)

	var testResults []testResult
	var mu sync.Mutex //for appending to results slice from multiple go routines

	for i := 0; i < workerPoolSize; i++ {

		go func() {
			defer wgTestWorker.Done()
			for test := range testQueueChan {
				testTimings, err := runDockerTest(test)

				mu.Lock()
				testResults = append(testResults, testResult{
					name: test.Name,
					err:  err,
				})
				mu.Unlock()

				log.Info(testTimings.Name + " is done")
				testTimingsChannel <- testTimings
			}
		}()

	}

	//Populate queue with tests to run
	for _, test := range testQueue {
		testQueueChan <- test
	}

	close(testQueueChan)
	wgTestWorker.Wait()

	close(testTimingsChannel)
	wgTestTiming.Wait()

	//Once all tests are complete, fail if any errors
	var errCount int
	for _, r := range testResults {
		if r.err != nil {
			log.Infof("FAIL: %s failed with:\n\t%s\n", r.name, r.err.Error())
			errCount++
		}
	}
	if errCount > 0 {
		t.Errorf("%d integration test failure(s)", errCount)
	}

	log.Info("Done!")
}

func filterTestsFromArgs(args string, allTests []integrationTest) []integrationTest {
	argTests := strings.Split(args, ",")

	var filteredTests []integrationTest

	for _, test := range allTests {
		for _, argTest := range argTests {
			regex := regexp.MustCompile(argTest)
			if regex.MatchString(test.Name) {
				if regex.MatchString(test.Name) {
					filteredTests = append(filteredTests, test)
				}
			}
		}
	}
	return filteredTests
}

func parseTestsFromYML() []integrationTest {
	var tests []integrationTest

	var patterns []string
	if runtime.GOOS == "windows" {
		patterns = []string{"integrationTests_windows.yml"}
	} else {
		patterns = []string{"integrationTests.yml"}
	}

	//starting search of files from root dir
	paths := []string{"./"}

	integrationFiles := tasks.FindFiles(patterns, paths)

	for _, filename := range integrationFiles {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Info("error reading file", err)
		}
		var fileTests []integrationTest
		ymlErr := yaml.Unmarshal(content, &fileTests)
		if ymlErr != nil {
			log.Info("Error reading yml", filename, ymlErr)
			os.Exit(1)
		}
		// Now add filepath to the tests found
		var fileTestsWithFilename []integrationTest
		for _, test := range fileTests {
			test.TestFile = filename
			fileTestsWithFilename = append(fileTestsWithFilename, test)

		}
		// Add this file's tests to the main slice
		tests = append(tests, fileTestsWithFilename...)

	}

	return tests
}

func runDockerTest(test integrationTest) (IntegrationTestRun, error) {
	start := time.Now()

	// Create initial Insights event struct with available env/default values
	currentTest := IntegrationTestRun{
		Name:              test.Name,
		StartTime:         start,
		StatusDockerBuild: "INCOMPLETE",
		StatusDockerRun:   "INCOMPLETE",
		Build:             "NO BUILD DETECTED",
		Context:           "NO CONTEXT DETECTED",
		OperatingSystem:   runtime.GOOS,
	}

	// Populate with Jenkins environment variables, if available.
	if os.Getenv("NODE_NAME") != "" {
		currentTest.Context = os.Getenv("NODE_NAME")
	}

	if os.Getenv("ghprbSourceBranch") != "" && os.Getenv("BUILD_ID") != "" {
		buildString := os.Getenv("ghprbSourceBranch") + " - #" + os.Getenv("BUILD_ID")
		currentTest.Build = buildString
	}

	if os.Getenv("USER") != "" {
		currentTest.User = os.Getenv("USER")
	}

	if os.Getenv("ghprbActualCommitAuthorEmail") != "" {
		currentTest.CommitAuthor = os.Getenv("ghprbActualCommitAuthorEmail")
	}

	// Create Docker images
	log.Info("Running ", test.Name)
	imageName := strings.ToLower("ci_" + test.Name)

	currentTest.DockerBuild.StartTimer()
	err := CreateDockerImage(imageName, test.DockerFROM, test.DockerCMD, test.DockerfileLines)
	if err != nil {
		log.Info("Test", test.Name, "failed ", err)
		log.Info("Test located in ", test.TestFile)
		currentTest.Error = err.Error()
		return currentTest.WrapUp(), err
	}
	currentTest.DockerBuild.StopTimer()
	currentTest.StatusDockerBuild = "DONE"

	currentTest.DockerRun.StartTimer()

	dockerCMD := "docker run --rm -it "
	for _, host := range test.HostsFileAdditions {
		dockerCMD += "--add-host=" + host + " "
	}
	dockerCMD += imageName

	logs, dockerErr := RunDockerContainer(imageName, test.HostsFileAdditions)
	if dockerErr != nil {

		log.Info("Test", test.Name, "failed ", dockerErr)
		outputTestHelp(test, dockerCMD)
		currentTest.Error = dockerErr.Error()
		return currentTest.WrapUp(), dockerErr
	}
	currentTest.DockerRun.StopTimer()
	currentTest.StatusDockerRun = "DONE"

	//Look for expected log entries
	found, fregex := searchOutput([]byte(logs), test.LogEntryExpected, true)
	//Look to ensure the Log entries not expected are not present
	notFound, nfregex := searchOutput([]byte(logs), test.LogEntryNotExpected, false)

	if found {
		currentTest.Status = "FAILED"
		currentTest.Error = "Log entry not found in output - " + fregex

		log.Info("Test", test.Name, "failed - Log entry not found in output -", fregex)
		outputTestHelp(test, dockerCMD)
		writeLogFile(logs, test.Name)
		return currentTest.WrapUp(), errors.New(currentTest.Error)
	} else if notFound {
		currentTest.Status = "FAILED"
		currentTest.Error = "Log entry found in output - " + nfregex

		outputTestHelp(test, dockerCMD)
		writeLogFile(logs, test.Name)
		return currentTest.WrapUp(), errors.New(currentTest.Error)
	}

	currentTest.Status = "SUCCESS"
	return currentTest.WrapUp(), nil
}

func writeLogFile(logs, testName string) {
	logfile := "logs/test_results_" + testName
	testFile := filepath.Clean(logfile)
	_ = os.WriteFile(testFile, []byte(logs), 0644)
}

func outputTestHelp(test integrationTest, dockerCMD string) {
	//log.Info("Full output was: --------------------\n", logs)
	log.Info("Command to run the docker container to manually inspect is: ", dockerCMD)
	log.Info("Test located in ", test.TestFile)
	log.Info("Command to run the test individually on Mac/Linux is: ./scripts/integrationTest.sh ", test.Name)
	log.Info("Command to run the test individually on Windows is: ./scripts/integrationTest_windows.ps1 ", test.Name)
	log.Info("Test output in file", "logs/test_results_"+test.Name)
}

// SearchOutput - searching a byte buffer for an expected or not expected regex string, returns true and the offending regex that failed to match (or not match)
func searchOutput(input []byte, regexes []string, expected bool) (bool, string) {

	// there isn't a verbose flag for tests, I change these to "Info" when I want to see them
	log.Debug("-------------------")
	log.Debugf("%s", input)
	log.Debug("-------------------")

	for _, regex := range regexes {
		regexKey, err := regexp.Compile(regex)
		if err != nil {
			log.Info("Error compiling regexes", err)
		}
		if expected {
			if !regexKey.Match(input) {
				log.Info("expected true, didn't find", regex)
				return true, regex
			}
		} else {
			// looking for a negative match
			if regexKey.Match(input) {
				log.Info("expected false, found", regex)
				return true, regex
			}
		}
	}
	// regex found (or not found), returning true
	return false, ""
}
