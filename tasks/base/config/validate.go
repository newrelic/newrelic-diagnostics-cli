package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/clbanning/mxj"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"gopkg.in/yaml.v3"
)

var quoted = regexp.MustCompile("^['`\"](.*)['`\"]$")

// BaseConfigValidate - Primary task to validate found config files. Will optionally take command line input as source
type BaseConfigValidate struct {
}

// ValidateElement - the validation that was done against the config
type ValidateElement struct {
	Config       ConfigElement
	Status       tasks.Status
	ParsedResult tasks.ValidateBlob
	Error        string
}

var (
	errConfigFileNotParse = errors.New("we cannot parse this file extension for this New Relic config file")
	errConfigFileNotRead  = "We ran into an error when trying to read your New Relic config file"
	errReaderMock         = errors.New("a reader error")
	errParsingYML         = "This can mean that you either have incorrect spacing/indentation around this line or that you have a syntax error, such as a missing/invalid character"
)

// MarshalJSON - custom JSON marshaling for this task, in this case we ignore the parsed config
func (el ValidateElement) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ConfigElement
		Status tasks.Status
		Error  string
	}{
		ConfigElement: el.Config,
		Status:        el.Status,
		Error:         el.Error,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseConfigValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/Validate")
}

// Explain - Returns the help text for each individual task
func (p BaseConfigValidate) Explain() string {
	return "Validate and parse New Relic configuration file(s)"
}

// Dependencies - Relies on the Collect task to be complete since it will attempt to validate the config files found
func (p BaseConfigValidate) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
	}
}

// Execute - The core work within each task
func (p BaseConfigValidate) Execute(options tasks.Options, results map[string]tasks.Result) tasks.Result {

	if results["Base/Config/Collect"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Config file collection was not successful, skipping validation step.",
		}
	}

	// Payload is a slice of config elements so let's build the structure to map them in
	// Confirm our payload is valid and expected format via type assertion
	configs, ok := results["Base/Config/Collect"].Payload.([]ConfigElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	if len(configs) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: no configs",
		}
	}

	validatedResults := []ValidateElement{}
	for _, config := range configs {
		processedConfig, err := processConfig(config)
		if err != nil {
			log.Debugf("%s - %s", config.FileName, err.Error())
			continue
		}
		validatedResults = append(validatedResults, processedConfig)
	}

	log.Debug("Done parsing results")
	log.Debug(len(validatedResults), "result(s) found")

	var parsingErrors string
	var successCounter, failureCounter int
	for _, validationResult := range validatedResults {
		if validationResult.Status == tasks.Success {
			log.Debug("Validation for", validationResult.Config.FileName, "Successful")
			successCounter++
		} else {
			log.Debug("Validation for", validationResult.Config.FileName, "Failed")
			failureCounter++
			parsingErrors += fmt.Sprintf("\n%s%s\n\tError: %s", validationResult.Config.FilePath, validationResult.Config.FileName, validationResult.Error)
		}
	}

	if successCounter > 0 && failureCounter == 0 {
		return tasks.Result{
			Summary: "Successfully parsed config file(s) - See json for full detail",
			Status:  tasks.Success,
			Payload: validatedResults,
		}
	}

	if failureCounter > 0 && successCounter == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Errors parsing the following %d configuration file(s):%s", failureCounter, parsingErrors),
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/configuration/configure-agent",
		}
	}

	log.Debug("Recorded ", successCounter, "Successful files parsed and ", failureCounter, "failures to parse config files")
	return tasks.Result{
		Status:  tasks.Warning,
		Summary: fmt.Sprintf("We were able to parse %d of %d configuration file(s).\nErrors parsing the following configuration file(s):%s", successCounter, (successCounter + failureCounter), parsingErrors),
		URL:     "https://docs.newrelic.com/docs/new-relic-diagnostics#run-diagnostics",
		Payload: validatedResults,
	}
}

func processConfig(config ConfigElement) (ValidateElement, error) {
	file := config.FilePath + config.FileName
	log.Debug("Validating " + file)

	//Read file
	content, err := os.Open(file)
	if err != nil {
		log.Debug("error reading file", err)
		return ValidateElement{
			Config: config,
			Status: tasks.Failure,
			Error:  err.Error(),
		}, nil
	}
	defer content.Close()
	// initialize variables for data
	var parsedConfig tasks.ValidateBlob

	// Depending on file type, route it to the appropriate parser
	fileType := filepath.Ext(file)

	switch fileType {
	case ".yml", ".yaml":
		log.Debug(".yml file found, validating")
		parsedConfig, err = ParseYaml(content)

	case ".xml", ".config":
		log.Debug(".xml file found, validating")
		parsedConfig, err = parseXML(content)

	case ".json":
		log.Debug(".json file found, validating")
		parsedConfig, err = ParseJSON(content)

	case ".js":
		log.Debug(".js file found, validating")
		parsedConfig, _ = parseJs(content)

	case ".ini", ".properties", ".cfg":
		log.Debug(".ini file found, validating")
		parsedConfig, _ = parseIni(content)

	default:
		return ValidateElement{}, errConfigFileNotParse
	}

	if err != nil {
		return ValidateElement{
			Config:       config,
			Status:       tasks.Failure,
			ParsedResult: parsedConfig,
			Error:        err.Error(),
		}, nil
	}
	return ValidateElement{
		Config:       config,
		Status:       tasks.Success,
		ParsedResult: parsedConfig,
	}, nil
}

// ParseYaml - This function reads a yml file to a map that can be searched via the FindString function
func ParseYaml(reader io.Reader) (tasks.ValidateBlob, error) {
	var t interface{}
	data, errFile := io.ReadAll(reader)
	if errFile != nil {
		return tasks.ValidateBlob{}, fmt.Errorf("%v : %v", errConfigFileNotRead, errFile)
	}
	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return tasks.ValidateBlob{}, fmt.Errorf("%v.\n%v", err, errParsingYML)
	}

	return convertToValidateBlob(t), nil
}

func parseXML(reader io.Reader) (results tasks.ValidateBlob, err error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return tasks.ValidateBlob{}, fmt.Errorf("%v : %v", errConfigFileNotRead, err)
	}
	m, merr := mxj.NewMapXml([]byte(data))
	if merr != nil {
		log.Debug("merr != nill = true")
		return results, merr
	}

	// Convert to a map[string]interface{} to match the structure
	i := make(map[string]interface{})
	for key, value := range m {
		i[key] = value
	}

	result := convertToValidateBlob(i)
	return result, nil
}

func ParseJSON(reader io.Reader) (result tasks.ValidateBlob, err error) {
	t := make(map[string]interface{})
	data, err := io.ReadAll(reader)
	if err != nil {
		return tasks.ValidateBlob{}, fmt.Errorf("%v : %v", errConfigFileNotRead, err)
	}
	err = json.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Debug("Error parsing json:", err)
		return
	}

	result = convertToValidateBlob(t)

	return
}

func ParseJSONarray(reader io.Reader) (result []tasks.ValidateBlob, err error) {
	var validateBlobs []tasks.ValidateBlob
	t := []map[string]interface{}{}
	data, err := io.ReadAll(reader)
	if err != nil {
		return validateBlobs, fmt.Errorf("%v : %v", errConfigFileNotRead, err)
	}
	err = json.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Debug("Error parsing json:", err)
		return
	}

	for _, jsonInterface := range t {
		validateBlob := convertToValidateBlob(jsonInterface)
		validateBlobs = append(validateBlobs, validateBlob)
	}

	return validateBlobs, nil
}

// formatJs - Outputs the newrelic.js file into a more parsable format with comments and other non-essential items removed
func formatJs(jsString string) ([]string, error) {
	// remove comments that start with //
	// have to be careful not to remove proxy configs
	slashCommentRe, err := regexp.Compile(`(?m)(^.*)([/][/].*)$`)
	if err != nil {
		return nil, err
	}
	quoteRe, err := regexp.Compile("['`\"]")
	if err != nil {
		return nil, err
	}
	removeSlashComments := slashCommentRe.ReplaceAllFunc([]byte(jsString), func(b []byte) []byte {
		s := string(b)
		checkForComment := strings.Split(s, "//")
		// nothing on the left of the //, whole line is a comment, just remove it
		if checkForComment[0] == "" {
			return nil
		}
		// check to see if the // is within quotes like 'http://...'
		quoteCount := len(quoteRe.FindAllStringIndex(checkForComment[0], -1))
		if quoteCount == 0 || quoteCount%2 == 0 {
			// not in quotes
			return []byte(checkForComment[0])
		}
		// keep the //, it was in quotes
		return b
	})
	// remove \n and \r
	removeLineBreaks := strings.ReplaceAll(string(removeSlashComments), "\n", "")
	removeCarriageReturn := strings.ReplaceAll(removeLineBreaks, "\r", "")

	// remove everything before exports.config =
	exportObj := strings.Split(removeCarriageReturn, "exports.config =")[1]

	// remove /* this type of comment */
	commentRe, err := regexp.Compile("[/][*].*?[*][/]")
	if err != nil {
		return nil, err
	}
	removeComments := commentRe.ReplaceAllString(exportObj, "")

	// find start curlies ({) and add a linebreak after it
	startCurliesRe, err := regexp.Compile("(^{|[^$]{)")
	if err != nil {
		return nil, err
	}
	fixStartCurlies := startCurliesRe.ReplaceAllString(removeComments, "{\n")

	// fix formatting for arrays, make them multi-line
	fixStartArrays := strings.ReplaceAll(fixStartCurlies, "[", "[\n")
	fixEndArrays := strings.ReplaceAll(fixStartArrays, "]", "\n]")

	// add a linebreak after commas
	fixCommas := strings.ReplaceAll(fixEndArrays, ",", ",\n")

	// create the array and trim whitespace
	var jsonString []string
	for _, s := range strings.Split(fixCommas, "\n") {
		// fix end curlies - only add a line break if { isn't also on the line
		if strings.Contains(s, "}") && !strings.Contains(s, "{") {
			fixEndCurlies := strings.ReplaceAll(s, "}", "\n}")
			splitAgain := strings.Split(fixEndCurlies, "\n")
			for _, ss := range splitAgain {
				jsonString = append(jsonString, strings.TrimSpace(ss))
			}
			continue
		}
		jsonString = append(jsonString, strings.TrimSpace(s))
	}
	return jsonString, nil
}

func parseJs(rawFile io.Reader) (result tasks.ValidateBlob, err error) {
	jsBytes, err := io.ReadAll(rawFile)
	if err != nil {
		return
	}
	jsonString, err := formatJs(string(jsBytes))
	if err != nil {
		return
	}
	log.Debug("Formatted js: ", strings.Join(jsonString, "\n"))
	tempMap := make(map[string]interface{})

	location := ""
	arrayLocation := ""
	var buildarray []string

loop:
	for lineNum, keyValue := range jsonString {
		log.Debugf("keyValue: `%s`", keyValue)
		switch keyValue {
		case "{": //do nothing here
		case "}", "},":
			log.Debug("location is", location)
			locationSplit := strings.Split(location, ".")
			location = locationSplit[len(locationSplit)-1]

		case "]", "],":
			log.Debug("done with array, adding ", buildarray)

			tempMap[arrayLocation] = buildarray
			arrayLocation = ""
			buildarray = nil

		case "};":
			break loop // end of config, just break the switch

		case "":

		default:
			if arrayLocation != "" {
				log.Debug("creating multi-line array", sanitizeValue(keyValue))
				buildarray = append(buildarray, sanitizeValue(keyValue))
				continue // we are in an array, go to next line
			}

			keyMap := strings.SplitN(keyValue, ":", 2)

			if len(keyMap) != 2 || keyMap[1] == "" {
				log.Debugf("Couldn't parse line number %d, line text: %v\n", lineNum, keyMap)
				break
			}

			log.Debug("adding: ", location+keyMap[0], " : ", strings.Replace(keyMap[1], ",", "", 1))
			// check for inline array value
			if string(strings.TrimSpace(keyMap[1])[0]) == "[" {
				log.Debug("array detected")
				//strip [] from string
				removeBrackets := regexp.MustCompile(`\[(.*)\]`)
				tempString := removeBrackets.FindStringSubmatch(keyMap[1])
				if len(tempString) != 2 {
					log.Debug("multi-line array detected", location+keyMap[0])
					if strings.TrimSpace(keyMap[1]) == "[" {
						arrayLocation = location + keyMap[0]
					}
					continue //move to the next line
				}
				// split on commas
				tempSlice := strings.Split(tempString[1], ",")

				//now loop through the slice to clean up whitespace and single quotes
				var finalSlice []string
				for _, value := range tempSlice {
					finalSlice = append(finalSlice, sanitizeValue(value))
				}
				tempMap[location+keyMap[0]] = finalSlice
			} else {
				tempMap[location+keyMap[0]] = sanitizeValue(keyMap[1])
			}

			if strings.TrimSpace(keyMap[1]) == "{" {
				log.Debug("creating adding location", keyMap[0])
				location = keyMap[0] + "." + location
			}

		}
	}

	return convertToValidateBlob(tempMap), nil
}

// "  'license-key-val-node',"  ->  "license-key-val-node"
func sanitizeValue(src string) string {
	//Remove whitespace
	whitespaceTrimmed := strings.TrimSpace(src)

	//Remove comma at end of line
	commaTrimmed := strings.Replace(whitespaceTrimmed, ",", "", 1)

	//Remove single/double quotes
	return trimQuotes(commaTrimmed)
}

func trimQuotes(src string) string {
	if quoted.Match([]byte(src)) {
		return quoted.ReplaceAllString(src, "$1")
	}
	return src
}

func parseIni(reader io.Reader) (result tasks.ValidateBlob, err error) {
	t := make(map[string]interface{})

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	// We take this regex and run through the file to map key value pairs. It should work on any file with = delimited key value pairs
	keyRegex, _ := regexp.Compile("^([^;#][a-zA-Z_.]*)([ ]*)[= ][ ]*(.*)")
	for scanner.Scan() {

		for _, value := range keyRegex.FindAllStringSubmatch(scanner.Text(), -1) {
			//Set key and value based on found matches, based on our regex search, this should be the active, uncommented values
			t[string(value[1])] = trimQuotes(string(value[3]))
		}
	}
	result = convertToValidateBlob(t)
	return
}

func convertToValidateBlob(input interface{}) (output tasks.ValidateBlob) {
	// peel back each layer of the configuration item and step through
	output.Path = ""
	output.Children = iterateMap("", input)
	// always sort the output before returning
	output.Sort()
	return
}

func iterateMap(parent string, input interface{}) []tasks.ValidateBlob {
	inputMap := make(map[string]interface{})
	// cast this input and switch on its type
	switch castInput := input.(type) {

	case map[string]interface{}:
		inputMap = castInput

	case map[interface{}]interface{}:
		// cast this thing to a map[string]interface{} so we can manipulate them the same way
		for key, value := range castInput {
			//key is type interface but should really be a string, generally
			stringKey, ok := key.(string)
			if !ok {
				log.Debug("Error casting key to string. Key is type:", reflect.TypeOf(key), "value is: ", key)
				continue
			}
			inputMap[stringKey] = value
		}

	case []interface{}:
		// walk a slice adding multiple items
		for i, input := range castInput {
			inputMap[strconv.Itoa(i)] = input
		}
	case []string:
		for i, input := range castInput {
			inputMap[strconv.Itoa(i)] = input
		}

	default:
		log.Debug("fallthrough, input type is ", reflect.TypeOf(castInput))

	}

	var parsedBlobs []tasks.ValidateBlob

	for key, value := range inputMap {
		var b tasks.ValidateBlob
		b.Key = key
		b.Path = parent

		//Add Children if value is a map or slice
		switch castValue := value.(type) {
		case map[interface{}]interface{}, map[string]interface{}, []interface{}, []string:
			b.Children = iterateMap(parent+"/"+key, castValue) // recursively convert children and their children
		default:
			b.RawValue = value
		}
		// add our fresh new blob to the list
		parsedBlobs = append(parsedBlobs, b)
	}
	return parsedBlobs

}
