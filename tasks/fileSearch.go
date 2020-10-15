package tasks

import (
	"bufio"
	"os"
	"regexp"

	log "github.com/newrelic/NrDiag/logger"
)

//global var for mocking to ensure test coverage
var osOpen = os.Open

//FindStringInFileFunc - function signature for FindStringInFile
type FindStringInFileFunc func(string, string) bool

//FindStringInFile - This takes in a regex string and searched line by line through the file indicated as a string to the path, can be relative or absolute path.
//Returns a boolean if the desired regex is found in the file
func FindStringInFile(search string, filepath string) bool {
	regexKey, _ := regexp.Compile(search)
	log.Debug("Opening " + filepath + "searching for " + search)
	//Read file
	file, err := osOpen(filepath)
	if err != nil {
		log.Debug("error reading file", filepath, err)
		//result.Status = tasks.Error
		return false
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if regexKey.Match(scanner.Bytes()) {
			log.Debug("Found key, setting true")
			return true
		}
	}
	log.Debug("string not found in file " + filepath + " setting false")
	return false

}

//ReturnStringInFileFunc - function definition for ReturnStringInFile
type ReturnStringInFileFunc func(string, string) ([]string, error)

//ReturnStringSubmatchInFileAllMatches Searches through an entire file, applying the supplied regex (search) term to each line individually
//Returns a slice [] of []string matches, with each element being returned from regex.FindStringSubmatch()
//Returns an error if the desired regex is not found in the file
func ReturnStringSubmatchInFileAllMatches(search string, filepath string) ([][]string, error) {
	regexKey, err := regexp.Compile(search)
	if err != nil {
		return nil, err
	}
	log.Debug("Opening " + filepath + " searching for " + search)
	//Read file
	file, err := osOpen(filepath)
	if err != nil {
		log.Debug("error reading file", filepath, err)
		return nil, err
	}
	var results [][]string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// scan the lines, look for submatch as string
		matches := regexKey.FindStringSubmatch(scanner.Text())
		if len(matches) > 0 {
			results = append(results, matches)
		}
	}
	if len(results) > 0 {
		return results, nil
	}
	log.Debug("string not found in file" + filepath + " setting false")
	return [][]string{}, nil
}

// ReturnStringSubmatchInFile - This takes in a regex string and searched line by line through the file indicated as a string to the path, can be relavtive or absolute path.
//Returns the first matching line if found.
func ReturnStringSubmatchInFile(search string, filepath string) ([]string, error) {
	fullResult, err := ReturnStringSubmatchInFileAllMatches(search, filepath)
	if err != nil {
		return []string{}, err
	}
	if len(fullResult) == 0 {
		return []string{}, nil
	}
	return fullResult[0], nil
}

// ReturnLastStringSubmatchInFile - This takes in a regex string and searched line by line through the file indicated as a string to the path, can be relavtive or absolute path.
//Returns the last matching line if found.
func ReturnLastStringSubmatchInFile(search string, filepath string) ([]string, error) {
	fullResult, err := ReturnStringSubmatchInFileAllMatches(search, filepath)
	if err != nil {
		return nil, err
	}
	if len(fullResult) == 0 {
		return []string{}, nil
	}
	return fullResult[len(fullResult)-1], nil
}
