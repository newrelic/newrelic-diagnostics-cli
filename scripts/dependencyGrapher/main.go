package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type taskInfo struct {
	Identifier   string
	Dependencies []string
}

func main() {

	workingDir, err := getWorkingDir()
	if err != nil {
		panic(err)
	}

	goFiles, walkErr := getGoFiles(workingDir)
	if walkErr != nil {
		panic(walkErr)
	}

	parsedTasks, parseErr := parseTasks(goFiles)
	if parseErr != nil {
		panic(parseErr)
	}

	taskCSV := taskMapToCSV(parsedTasks)

	for _, line := range taskCSV {
		fmt.Println(line)
	}

}

//parseTasks takes in slice of filepaths for the Diagnostics CLI tasks. It reads each file and parses out task identifier and dependencies
func parseTasks(goFiles []string) (map[string]map[string]bool, error) {
	foundTasks := map[string]map[string]bool{}
	for _, file := range goFiles {
		if isTaskFile(file) {
			task, err := parseTaskInfo(file)
			if err != nil {
				return foundTasks, err
			}
			if task.Identifier == "" {
				continue
			}
			if foundTasks[task.Identifier] == nil {
				foundTasks[task.Identifier] = make(map[string]bool)
			}
			for _, dependency := range task.Dependencies {
				foundTasks[task.Identifier][dependency] = true
			}
		}
	}

	return foundTasks, nil
}

func taskMapToCSV(tasksMap map[string]map[string]bool) []string {
	taskCSV := []string{}
	for identifier, dependencies := range tasksMap {

		if strings.Contains(identifier, "Example") {
			continue
		}
		if len(dependencies) > 0 { //Does this task have dependencies?
			for dependency := range dependencies {
				taskCSV = append(taskCSV, dependency+","+identifier)
			}
		} else {
			taskCSV = append(taskCSV, ","+identifier)
		}
	}

	return taskCSV
}

func parseTaskInfo(taskFilePath string) (taskInfo, error) {
	var task taskInfo
	taskFileContent, err := ioutil.ReadFile(taskFilePath)
	if err != nil {
		return task, err
	}

	identifier := parseIdentifier(string(taskFileContent))
	if identifier != "" {
		task.Identifier = identifier
	}

	dependencies := parseDependencies(string(taskFileContent))
	task.Dependencies = dependencies

	return task, nil
}

func parseDependencies(fileContent string) []string {

	//remove comments
	re := regexp.MustCompile(`\/\/.*\r?\n`)
	filteredContent := re.ReplaceAllString(fileContent, "")

	//remove newlines
	re = regexp.MustCompile(`\r?\n`)
	flattendedConent := re.ReplaceAllString(filteredContent, "")

	//grab slice of strings returned from Dependencies()
	re = regexp.MustCompile(`Dependencies\(\) \[\]string\s*{\s*return\s*\[\]string{\s*([^}]+)\s*}`)
	matches := re.FindStringSubmatch(flattendedConent)

	dependencies := []string{}

	if len(matches) > 1 {
		dependencySlice := strings.Split(matches[1], ",")

		for _, dependency := range dependencySlice {
			re = regexp.MustCompile(`"(.+)"`)
			matches = re.FindStringSubmatch(dependency)
			if len(matches) > 1 {
				dependencies = append(dependencies, matches[1])
			}
		}
	}

	return dependencies
}

func parseIdentifier(fileContent string) string {
	re := regexp.MustCompile(`tasks\.IdentifierFromString\("(.+)"\)`)
	matches := re.FindStringSubmatch(fileContent)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

//check file line by line and return true if all task interface method signatures are found
func isTaskFile(filePath string) bool {
	foundInterfaceMethods := 0
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLine := scanner.Text()
		if strings.Contains(currentLine, "Identifier() tasks.Identifier") ||
			strings.Contains(currentLine, "Explain() string") ||
			strings.Contains(currentLine, "Dependencies() []string") {
			foundInterfaceMethods++
		}
		if foundInterfaceMethods == 3 {
			return true
		}

	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return false
}

//recursively find all .go files in working directory/sub directories
func getGoFiles(expath string) ([]string, error) {
	foundFiles := []string{}
	err := filepath.Walk(expath, func(path string, f os.FileInfo, err error) error {

		// Check if we're running via go test by checking for registered test flag
		if flag.Lookup("test.v") == nil && strings.Contains(path, "fixtures") {
			return nil
		}
		if strings.Contains(path, ".go") {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})
	return foundFiles, err
}

//get working directory of process
func getWorkingDir() (string, error) {
	exPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exPath), nil
}
