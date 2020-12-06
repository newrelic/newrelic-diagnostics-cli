package tasks

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/shirou/gopsutil/process"
)

// OsFunc - for dependency injecting osGetwd
type OsFunc func() (string, error)

// FindFiles - looks for files in the standard search paths that match the given string.
// automatically dedupes matches and attempts to resolve any symlinks in the paths slice.
func FindFiles(patterns []string, paths []string) []string {
	// map to automatically dedupe file matches
	foundFiles := make(map[string]interface{})

	for _, path := range paths {
		//Check if path is a symlink and if so, set symPath as path
		symPath, err := filepath.EvalSymlinks(path)
		if err == nil {
			path = symPath
		}
		filepath.Walk(path, func(pathInfo string, fileInfo os.FileInfo, walkErr error) error {
			if walkErr != nil {
				// log the error and move on to next item to be walked
				log.Debug("Error when walking filesystem:", walkErr)
				return nil
			}
			if !fileInfo.IsDir() {
				// loop through pattern list and add files that match to our string array
				for _, pattern := range patterns {
					var validID = regexp.MustCompile(pattern)
					match := validID.MatchString(fileInfo.Name())
					if match {
						foundFiles[pathInfo] = struct{}{} // empty struct is smallest memory footprint
					}
				}
			}
			return nil
		})
	}

	var uniqueFoundFiles []string
	for fileLocation := range foundFiles {
		uniqueFoundFiles = append(uniqueFoundFiles, fileLocation)
	}
	return uniqueFoundFiles
}

// FindProcessByNameFunc - allows FindProcessByName to be dependency injected
type FindProcessByNameFunc func(string) ([]process.Process, error)

// FindProcessByName - returns array of processes matching string name, or an error if we can't gather a list of processes, or an empty slice and nil if we found no processes with that specific name
func FindProcessByName(name string) ([]process.Process, error) {
	var processList []process.Process

	processIDs, err := process.Pids()

	if err != nil {
		log.Debug("error", err)
		return processList, err
	}
	for _, PID := range processIDs {

		processID := process.Process{Pid: PID}
		processName, err := processID.Name()

		if err != nil {
			log.Debug("error getting process name", PID)
			return processList, err

		}

		if filepath.Ext(name) != ".exe" { //exact match if searching for process with .exe, otherwise strip the .exe before comparing
			processName = strings.Replace(processName, ".exe", "", 1)
		}
		//Now search for a match
		if name == processName {
			processList = append(processList, processID)
		}

	}
	return processList, nil
}

//ReadFile - reads file from path to string
func ReadFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Debug("error reading file", err)
		return ""
	}
	return string(content)
}

// FileExistsFunc - allows FileExists to be dependency injected
type FileExistsFunc func(string) bool

//FileExists - checks for existence of file
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

//PromptUser - This takes the input string as the query to the end users and waits for a response
func PromptUser(msg string, options Options) bool {
	if options.Options["YesToAll"] == "true" {
		return true
	}

	prompt := "Choose 'y' or 'n', then press enter: "
	yesResponses := []string{"y", "yes"}
	noResponses := []string{"n", "no"}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(msg)
	fmt.Print(prompt)

	for scanner.Scan() {
		userInput := strings.ToLower(scanner.Text())

		if ContainsString(noResponses, userInput) {
			return false
		}

		if ContainsString(yesResponses, userInput) {
			return true
		}

		//Repeat prompt if invalid input is provided
		fmt.Printf("%s", prompt)
	}
	return false
}

// PosString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func PosString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// ContainsString returns true if slice contains element
func ContainsString(slice []string, element string) bool {
	return !(PosString(slice, element) == -1)
}

// CaseInsensitiveStringContains - case insensitive strings.Contains
func CaseInsensitiveStringContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

// DedupeStringSlice takes a slice of strings as input and returns a randomly-ordered slice with duplicates removed
func DedupeStringSlice(s []string) []string {
	deDuped := []string{}
	uniques := map[string]bool{}

	for _, element := range s {
		uniques[element] = true
	}

	for k := range uniques {
		deDuped = append(deDuped, k)
	}
	return deDuped
}

// StringInSlice checks if a supplied string exists in the supplied slice of strings
func StringInSlice(s string, list []string) bool {
	for _, element := range list {
		if s == element {
			return true
		}
	}
	return false
}

// ValidateBlob - This stores a validated config structure and has various methods attached to it allowing searching, conversion of values, etc..
type ValidateBlob struct {
	Key      string
	Path     string
	RawValue interface{}
	Children []ValidateBlob
}

//ByChild is a sort helper to sort an array of ValidateBlobs by their child nodes
type ByChild []ValidateBlob

func (t ByChild) Len() int {
	return len(t)
}
func (t ByChild) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t ByChild) Less(i, j int) bool {
	c1 := t[i].PathAndKey()
	c2 := t[j].PathAndKey()
	pathandkeys := []string{c1, c2}
	sort.Strings(pathandkeys)
	if pathandkeys[0] == c1 {
		return true
	}
	return false
}

//Sort sorts an inplace object, organizing the children alphabetically
func (v ValidateBlob) Sort() {

	if !v.IsLeaf() {
		sort.Sort(ByChild(v.Children))
	}
	for _, child := range v.Children {
		if !child.IsLeaf() {
			child.Sort()
		}
	}

}

// IsLeaf - returns true if a ValidateBlob is a leaf, i.e. has a value, or false if it has children
func (v ValidateBlob) IsLeaf() bool {
	if v.Children == nil {
		return true
	}
	return false
}

// Value - returns the value as a string, if is a node
func (v ValidateBlob) Value() string {

	if !v.IsLeaf() {
		log.Debug(v.Key, "has children")
		return ""
	}

	switch value := v.RawValue.(type) {
	case string:
		return value
	case bool:
		if value {
			return "true"
		}
		return "false"

	case float64:
		return strconv.FormatFloat(value, 'E', -1, 64)
	case int:
		return strconv.Itoa(value)
	case nil:
		return ""
		//results.Data <- [2]string{context + "/" + search, "nil"}
	case map[interface{}]interface{}:
		//results.Data <- [2]string{context + "/" + search, fmt.Sprintf("%v", value)}

	case []interface{}:
		var results string
		for _, arrayValue := range value {
			switch arrayValueType := arrayValue.(type) {

			case int:
				results += strconv.Itoa(arrayValueType) + " "
			case string:
				results += arrayValueType + " "
			default:
				log.Debug("Found key array but value fell through, value is type", reflect.TypeOf(value))
			}

		}
		return results // now return the string concatenated array
	default:
		log.Debug("Found key but value fell through, value is type:", reflect.TypeOf(value))

		return ""

	}
	return ""
}

// PathAndKey - returns the Path and Key of a ValidateBlob as a string
func (v ValidateBlob) PathAndKey() string {
	return v.Path + "/" + v.Key
}

//PathAndKeyContains - Just a quick wrapper around strings.Contains to make it easier to call
func (v ValidateBlob) PathAndKeyContains(searchKey string) bool {
	if strings.Contains(v.PathAndKey(), searchKey) {
		return true
	}
	return false
}

// FindKey - returns one or more ValidateBlobs within a given ValidateBlob that has the desired key
// can return multiple nodes as it will return everything that matched path/key
func (v ValidateBlob) FindKey(searchKey string) (results []ValidateBlob) {

	//check current node for matching
	if v.Key == searchKey {
		log.Debug("found match, adding key", v.Key)
		results = append(results, v)
	}

	if !v.IsLeaf() {
		results = append(results, v.searchChildren(searchKey)...)
	}
	return
}

//FindKeyByPath - returns a single ValidateBlob that matches the exact full path of the Key in question. This means the searchPath parameter MUST begin with a / character
func (v ValidateBlob) FindKeyByPath(searchPath string) ValidateBlob {

	var results []ValidateBlob
	//check current node for matching
	if v.PathAndKey() == searchPath {
		log.Debug("found match, adding key", v.Key)
		return v
	}

	split := strings.Split(searchPath, "/")

	searchKey := split[len(split)-1] //This just gives us the last element in the slice
	log.Debug("key to find is", searchKey)
	if !v.IsLeaf() {
		results = append(results, v.searchChildren(searchKey)...)
	}

	//Now loop through the results to find the one we actually want, will return the first match (should only ever be one match with a valid config anyway...)

	for _, validateBlob := range results {
		if searchPath == searchKey { //avoid returning key only match
			if validateBlob.PathAndKey() == searchPath {
				return validateBlob
			}
		} else { //look for partial path matches
			if strings.Contains(validateBlob.PathAndKey(), searchPath) {
				return validateBlob
			}
		}
	}

	return ValidateBlob{}
}

func (v ValidateBlob) searchChildren(searchKey string) (results []ValidateBlob) {

	for _, child := range v.Children {

		if child.Key == searchKey {
			log.Debug("adding", child.PathAndKey(), "to list of results")
			results = append(results, child)
		}

		if !child.IsLeaf() {
			results = append(results, child.searchChildren(searchKey)...)
		}

	}
	return
}

//UpdateKey - This requires an exact path to the key in question to update and returns the original blob with just the one key value updated
func (v ValidateBlob) UpdateKey(searchKey string, replacementValue interface{}) ValidateBlob {

	log.Debug("Updating key", searchKey, "with new value", replacementValue)
	//we're going to walk the tree, adding nodes as we find them to the new returned object and just create a new entry if and only if it's the one we want

	if v.PathAndKey() == searchKey {
		log.Debug("found replacement on first try", v.PathAndKey())
		v.RawValue = replacementValue
		return v
	}

	//next search for exact key location and confirm only one exists
	results := v.FindKeyByPath(searchKey)
	log.Debug("results in updateKey", results)
	// set children of this node to any updated children
	v.Children = v.updateChildren(searchKey, replacementValue)
	return v
}

func (v ValidateBlob) updateChildren(searchKey string, replacementValue interface{}) (output []ValidateBlob) {
	// walk the tree
	for _, node := range v.Children {
		if node.PathAndKey() == searchKey {
			log.Debug("Found matching PathAndKey", node.PathAndKey())
			node.RawValue = replacementValue
		}
		if !node.IsLeaf() {
			node.Children = node.updateChildren(searchKey, replacementValue)
		}
		output = append(output, node)
	}
	return
}

//InsertKey - This requires an exact path to the key in question to update and returns the original blob with just the one key value updated, the key inserted should NOT start with a / and infer the path due to nesting at the top level
func (v ValidateBlob) InsertKey(insertKey string, value interface{}) ValidateBlob {

	log.Debug("Inserting key", insertKey, "with value", value)

	//First check to see if the key's path is at the top
	if strings.Count(insertKey, "/") == 0 {
		log.Debug("Adding key to the top level ValidateBlob")
		blob := ValidateBlob{Key: insertKey, RawValue: value}
		log.Debug(blob)
		v.Children = append(v.Children, blob)
		log.Debug(v)
		return v
	}
	//we're going to walk the tree, adding nodes as we find them to the new returned object and just create a new entry if and only if it's the one we want

	v.Children = v.insertChildKey(insertKey, value)
	return v
}

func (v ValidateBlob) insertChildKey(insertKey string, value interface{}) (output []ValidateBlob) {
	// walk the tree

	depth := strings.Count(v.PathAndKey(), "/")

	//split up the key to have the actual key
	split := strings.Split(insertKey, "/") //Skip the first / here to ignore the preceeding / when calculating
	key := split[len(split)-1]
	if split[0] == "" {
		split = split[1:] // drop first element to ignore the preceeding /
	}
	newSplit := split[:len(split)-1] // drop last element of the slice to build the path
	path := ""
	for _, pathItem := range newSplit {
		path += "/" + pathItem
	}
	blob := ValidateBlob{Key: key, RawValue: value, Path: path}
	log.Debug("adding key with path:", path, "blob:", blob)

	var matches int
	for _, node := range v.Children {
		if node.PathAndKey() == path {
			matches++
			log.Debug("Found matching PathAndKey", node.PathAndKey())
			node.Children = append(node.Children, blob)
		}
		if !node.IsLeaf() {
			matches++
			node.Children = node.insertChildKey(insertKey, value)
		}

		output = append(output, node)
	}
	if matches == 0 {
		log.Debug("failed to find path to node, adding partial node to existing", v)
		if len(split)+1 == depth {
			return //depth of node reached, stop searching to avoid panics
		}
		node := ValidateBlob{Key: split[depth-1], Path: v.PathAndKey()}
		if depth != len(split) {
			node.Children = node.insertChildKey(insertKey, value)
		} else {
			node.RawValue = value
		}
		output = append(output, node)
	}

	return
}

//UpdateOrInsertKey - helper function to just update or insert without caring which one is used, key should NOT begin with /
func (v ValidateBlob) UpdateOrInsertKey(key string, value interface{}) ValidateBlob {

	if key == "" {
		log.Debug("Empty key, returning")
		return v
	}
	log.Debug("insert or update", key, "with value", value)
	//Check for the key first to see if we are doing an insert or an update
	exact := v.FindKeyByPath(key)
	log.Debug("log exact result is", exact, "with pathandkey", exact.PathAndKey())
	if exact.PathAndKey() != "/" {
		log.Debug("One result found, updating exact match", exact.PathAndKey(), "with value", value)
		return v.UpdateKey(exact.PathAndKey(), value)
	}

	//Now search more globally for a match
	search := v.FindKey(key)
	log.Debug("results of search is ", search)
	if v.PathAndKeyContains(key) {
		log.Debug("Path contains key, updating key", key, "with value:", value)
		return v.InsertKey(key, value)
	} else if len(search) == 0 {
		log.Debug("No results found, inserting", key, "with value", value)
		v = v.InsertKey(key, value)
	}

	log.Debug("Found ", len(search), "values, updating all of them")

	for _, results := range search {
		v = v.UpdateKey(results.PathAndKey(), value)
	}

	return v

}

//String - This dumps an entire ValidateBlob, including any/all children nodes to a string, with one line per ValidateBlob and nesting showing children
func (v ValidateBlob) String() (output string) {

	if v.IsLeaf() {
		output = fmt.Sprintf(v.PathAndKey() + ": " + v.Value() + "\n")
	} else {
		sort.Sort(ByChild(v.Children))
		for _, child := range v.Children {
			output += child.String()
		}
	}

	return
}

//AsMap - This builds a map of key to value for all leaf nodes of the given blob
func (v ValidateBlob) AsMap() map[string]interface{} {
	output := make(map[string]interface{})

	if v.IsLeaf() {
		output[v.PathAndKey()] = v.RawValue
	} else {
		sort.Sort(ByChild(v.Children))
		for _, child := range v.Children {
			for k, v := range child.AsMap() {
				output[k] = v
			}
		}
	}

	return output
}

// EnvironmentVariables - handles environment variables
type EnvironmentVariables struct {
	All   map[string]string
	Scope EnvVarScope
	PID   int32
}

// EnvVarScope - scope of the env variables - global, user, or process
type EnvVarScope int

//Constants for use by the status property above
const (
	// Global - the env variables contained in the struct are global level
	Global EnvVarScope = iota
	// Shell - the env variables contained in the struct are shell or user level
	Shell
	// Process - the env variables contained in the struct are process level
	Process
)

var defaultEnvVarFilter = []string{
	"NEWRELIC",
	"NEW_RELIC",
	"^NRIA",
	"^PATH$",
	"^HOME$",
	"^RUBY_ENV$",
	"^RAILS_ENV$",
	"^APP_ENV$",
	"^RACK_ENV$",
	"^LOCALAPPDATA$",
	"^DOTNET_SDK_VERSION$",
	"^DOTNET_INSTALL_PATH$",
	"^COR_PROFILER$",
	"^COR_PROFILER_PATH$",
	"^COR_ENABLE_PROFILER$",
	"^CORECLR_ENABLE_PROFILING$",
	"^CORECLR_PROFILER$",
	"^CORECLR_PROFILER_PATH$",
	"^ProgramFiles$",
	"^ProgramData$",
	"^APPDATA$",
	"^JBOSS_HOME$",
	"^WEBSITE_SITE_NAME$", //Needed for detecting Azure environment
}

// GetDefaultFilterRegex - returns the default filter string array with regex included
func (e EnvironmentVariables) GetDefaultFilterRegex() []string {
	return defaultEnvVarFilter
}

// GetDefaultFilterString - returns the default filter string with the regex removed
func (e EnvironmentVariables) GetDefaultFilterString() (envListSanitized string) {
	r := strings.NewReplacer("^", "",
		"$", "",
	)

	for _, envVar := range defaultEnvVarFilter {
		envListSanitized += " " + r.Replace(envVar)
	}
	return
}

// WithDefaultFilter - applies the default env var filter defined above
func (e EnvironmentVariables) WithDefaultFilter() (filtered map[string]string) {
	filtered = make(map[string]string)
	for key, value := range e.All {
		for _, fValue := range defaultEnvVarFilter {
			r := regexp.MustCompile("(?i)" + fValue) //case insensitive match
			if r.MatchString(key) {
				filtered[key] = value
			}
		}
	}

	return
}

// WithCustomFilter - applies a custom env var filter, can include the default filter. Custom filters take priority over default
func (e EnvironmentVariables) WithCustomFilter(customFilter []string, includeDefaults bool) (filtered map[string]string) {
	filtered = make(map[string]string)

	for key, value := range e.All {
		// first apply the custom filter
		for _, fValue := range customFilter {
			r := regexp.MustCompile("(?i)" + fValue) //case insensitive match
			if r.MatchString(key) {
				filtered[key] = value
			}
		}

		// include defaults
		if includeDefaults {
			// don't overwrite custom value with default value
			if _, ok := filtered[key]; ok {
				for _, fValue := range defaultEnvVarFilter {
					r := regexp.MustCompile("(?i)" + fValue)
					if r.MatchString(key) {
						filtered[key] = value
					}
				}
			}
		}
	}

	return
}

// FindCaseInsensitive - searches for a specific variable case insenstive
func (e EnvironmentVariables) FindCaseInsensitive(envVar string) string {
	r := regexp.MustCompile("(?i)" + envVar)
	for key, value := range e.All {
		if r.MatchString(key) {
			return value
		}
	}
	return ""
}

// GetProcessEnvVars - gathers a given process's Env Vars
func GetProcessEnvVars(pid int32) (envVars EnvironmentVariables, retErr error) {
	envVars.All = make(map[string]string)
	envVars.Scope = Process
	envVars.PID = pid

	switch runtime.GOOS {
	case "linux":
		pidString := strconv.FormatInt(int64(pid), 10)
		path := filepath.Join("/proc", pidString, "environ")
		environFile, err := ioutil.ReadFile(path)
		if err != nil {
			errorString := "Error reading process env variables: " + err.Error()
			log.Debug(errorString)
			retErr = errors.New(errorString)
			return
		}
		envVarString := string(environFile)
		lines := strings.Split(envVarString, "\x00")
		for _, line := range lines {
			split := strings.Split(line, "=")
			if len(split) > 1 {
				name := split[0]
				val := split[1]
				envVars.All[name] = val
			}
		}
	default:
		errorString := "GetProcessEnvVars is not implemented for " + runtime.GOOS
		log.Debug(errorString)
		retErr = errors.New(errorString)
		return
	}

	return
}

// GetShellEnvVars - gathers a given process's Env Vars
func GetShellEnvVars() (envVars EnvironmentVariables, retErr error) {
	envVars.All = make(map[string]string)
	envVars.Scope = Shell
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envVars.All[pair[0]] = pair[1]
	}

	if len(envVars.All) == 0 {
		log.Debug("Env vars slice is 0, assuming there was an error collecting them")
		retErr = errors.New("Unable to gather any Environment Variables from the current shell")
		return
	}

	return
}

// GetCmdLineArgs is a wrapper for Process.Cmdline
func GetCmdLineArgs(proc process.Process) ([]string, error) {
	return proc.CmdlineSlice() //Keep in mind that If an single argument looked like this: -Dnewrelic.config.app_name="my appname", Go for darwin will still separate by spaces and will split it into 2 arguments even if they were enclosed by quotes.
}

// JavaProcArgs has a Java process id with its associated cmdline arguments
type JavaProcArgs struct {
	ProcID int32
	Args   []string
}

// GetJavaProcArgs returns a slice of JavaProcArgs struct with two fields: ProcID(int32) and Args([]string)
func GetJavaProcArgs() []JavaProcArgs {
	javaProcs, err := FindProcessByName("java")
	if err != nil {
		log.Info("We encountered an error while detecting all running Java processes: " + err.Error())
	}
	if javaProcs == nil {
		return []JavaProcArgs{}
	}
	javaProcArgs := []JavaProcArgs{}
	for _, proc := range javaProcs {
		cmdLineArgsSlice, err := GetCmdLineArgs(proc)
		if err != nil {
			log.Info("Error getting command line options while running GetCmdLineArgs(proc)")
		}
		javaProcArgs = append(javaProcArgs, JavaProcArgs{ProcID: proc.Pid, Args: cmdLineArgsSlice})
	}
	return javaProcArgs
}

// ProcIDSysProps has a running java process id and its associated system properties organized as map of key system property and the value of the system property
type ProcIDSysProps struct {
	ProcID           int32
	SysPropsKeyToVal map[string]string
}

// GetNewRelicSystemProps - gathers a given process's System Properties that are New Relic related
func GetNewRelicSystemProps() []ProcIDSysProps {

	javaProcessesArgs := GetJavaProcArgs()

	if len(javaProcessesArgs) > 0 {
		procIDSysProps := []ProcIDSysProps{}

		for _, proc := range javaProcessesArgs {

			sysPropsKeyToVal := make(map[string]string)
			for _, arg := range proc.Args {
				if strings.Contains(arg, "-Dnewrelic") || strings.Contains(arg, "-Djava.io.tmpdir") {
					keyVal := strings.Split(arg, "=")
					sysPropsKeyToVal[keyVal[0]] = keyVal[1]
					procIDSysProps = append(procIDSysProps, ProcIDSysProps{ProcID: proc.ProcID, SysPropsKeyToVal: sysPropsKeyToVal})
				}
			}
		}

		return procIDSysProps
	}
	// no java processes running :(
	return []ProcIDSysProps{}
}

// TrimQuotes helper func for cleaning up single and double quoted yml values
// from ValidateBlobs when it's time to use them (filepaths for example)
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// GetVersionSplit - Splits a string version into integers. Supports up to 4 numbers ie 1.2.3.4
// 1 = major, 2 = minor, 3 = patch, 4 = build
func GetVersionSplit(version string) (majorVer int, minorVer int, patchVer int, buildVer int) {
	verSplit := strings.Split(version, ".")
	majorVer, err := strconv.Atoi(verSplit[0])
	if err != nil {
		log.Debug("Error converting major version string to an Int ", err.Error())
		majorVer = -1
	}
	if len(verSplit) > 1 {
		minorVer, err = strconv.Atoi(verSplit[1])
		if err != nil {
			log.Debug("Error converting minor version string to an Int ", err.Error())
			minorVer = -1
		}
	}
	if len(verSplit) > 2 {
		patchVer, err = strconv.Atoi(verSplit[2])
		if err != nil {
			log.Debug("Error converting patch version string to an Int ", err.Error())
			patchVer = -1
		}
	}
	if len(verSplit) > 3 {
		buildVer, err = strconv.Atoi(verSplit[3])
		if err != nil {
			log.Debug("Error converting build version string to an Int ", err.Error())
			buildVer = -1
		}
	}
	return
}

//MakeMapFromString parses a string `src` a string map.
//param: `pairSeparator` is the string value that separates pairs of key values such as '\n',
//param: `value separator` is the string value that separates keys from values, such as ':'.
//It will also cleanup leading and trailing space for keys and values. This means nested values will be flattened.
//As a string map keys will be indexed in alpha order and will not preserve original order of `src`.
//("a:b\nc:d", "\n", ":") -> map[string]string{ "a":"b", "c":"d" }
func MakeMapFromString(src string, pairSeparator string, valueSeparator string) map[string]string {
	valuePairsMap := make(map[string]string)
	valuePairsSlice := strings.Split(src, pairSeparator)

	for _, pair := range valuePairsSlice {
		if strings.Contains(pair, valueSeparator) {
			pairSlice := strings.Split(pair, valueSeparator)
			key := strings.TrimSpace(pairSlice[0])
			value := strings.TrimSpace(pairSlice[1])

			valuePairsMap[key] = value
		}
	}

	return valuePairsMap
}

//StreamWrapper provides a means to bundle a channel for publishing data,
// and a channel for publishing errors.
//Each is paired with a waitgroup if needed
type StreamWrapper struct {
	Stream      chan string
	StreamWg    *sync.WaitGroup
	ErrorStream chan error
	ErrorWg     *sync.WaitGroup
}

//BufferedCommandExecFunc allows us to declare BufferedCommandExec as a dependency type
type BufferedCommandExecFunc func(limit int64, cmd string, arg ...string) (*bufio.Scanner, error)

//BufferedCommandExec provides buffered execution of a command.
//@limit int64 of the max bytes to read. If 0, no limit is imposed.
//@cmd - the command to run
//@args - the args to run for the command
//This command combines stdout and stderr and provides a Scanner for buffered read of the pipes
//Or returns an error if encountered in creating the pipes or starting the command
//Scanner has a default token size of the constant MaxScanTokenSize (64 * 1024)
//Default buffer size is 4096: https://github.com/golang/go/blob/13cfb15cb18a8c0c31212c302175a4cb4c050155/src/bufio/scan.go#L76
func BufferedCommandExec(limit int64, cmd string, args ...string) (*bufio.Scanner, error) {
	cmdBuild := exec.Command(cmd, args...)

	stdoutPipe, stdoutPipeError := cmdBuild.StdoutPipe()
	if stdoutPipeError != nil {
		return nil, stdoutPipeError
	}

	stderrPipe, stderrPipeError := cmdBuild.StderrPipe()
	if stderrPipeError != nil {
		return nil, stderrPipeError
	}

	multiReader := io.MultiReader(stdoutPipe, stderrPipe)

	if limit != 0 {
		multiReader = io.LimitReader(multiReader, limit)
	}

	scanner := bufio.NewScanner(multiReader)

	if cmdStartError := cmdBuild.Start(); cmdStartError != nil {
		return scanner, cmdStartError
	}

	return scanner, nil
}

func BytesToPrettyJSONBytes(bytes []byte) ([]byte, error) {
	var unmarshaled interface{}

	firstByte := string(bytes[0])

	if firstByte == "[" {
		unmarshaled = []interface{}{}
	} else if firstByte == "{" {
		unmarshaled = make(map[string]interface{})
	} else {
		return bytes, errors.New("Invalid JSON: First character is " + string(firstByte))
	}

	unMarshalError := json.Unmarshal(bytes, &unmarshaled)

	if unMarshalError != nil {
		return bytes, unMarshalError
	}

	prettyJSONBytes, marshalError := json.MarshalIndent(unmarshaled, "", " ")

	if marshalError != nil {
		return bytes, marshalError
	}

	return prettyJSONBytes, nil
}

// CmdExecFunc represents a type that matches the signature of the exec.Command
// to be used a struct field for dependency injecting exec.Command wrappers.
type CmdExecFunc func(name string, arg ...string) ([]byte, error)

// CmdExecutor wraps the exec.Command function to facilitate dependency
// injection for testing tasks.
func CmdExecutor(name string, arg ...string) ([]byte, error) {
	cmdBuild := exec.Command(name, arg...)
	return cmdBuild.CombinedOutput()
}

//cmdWrapper is used to specify commands & args to be passed to the multi-command executor (mCmdExecutor)
//allowing for: cmd1 args | cmd2 args
type CmdWrapper struct {
	Cmd  string
	Args []string
}

// takes multiple commands and pipes the first into the second
func MultiCmdExecutor(cmdWrapper1, cmdWrapper2 CmdWrapper) ([]byte, error) {

	cmd1 := exec.Command(cmdWrapper1.Cmd, cmdWrapper1.Args...)
	cmd2 := exec.Command(cmdWrapper2.Cmd, cmdWrapper2.Args...)

	// Get the pipe of Stdout from cmd1 and assign it
	// to the Stdin of cmd2.
	pipe, err := cmd1.StdoutPipe()
	if err != nil {
		return []byte{}, err
	}
	cmd2.Stdin = pipe

	// Start() cmd1, so we don't block on it.
	err = cmd1.Start()
	if err != nil {
		return []byte{}, err
	}

	// Run Output() on cmd2 to capture the output.
	return cmd2.CombinedOutput()

}

// HTTPRequestFunc represents a type that matches the signature of the httpHelper.MakeHTTPRequest
// to be used a struct field for dependency injecting httpHelper.MakeHTTPRequest.
type HTTPRequestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

// HTTPRequester does stuff
func HTTPRequester(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return httpHelper.MakeHTTPRequest(wrapper)
}

//GetWorkingDirectoriesFunc function type declaration for GetWorkingDirectories
type GetWorkingDirectoriesFunc func() []string

type osFunc func() (string, error)

var osGetwd = os.Getwd
var osExecutable = os.Executable

//GetWorkingDirectories returns a slice of local directories. It purposefully ignores any errors and simply returns an empty slice if the function is unable to return the list of directories.
func GetWorkingDirectories() []string {
	var directories []string
	localDir, errWd := osGetwd()
	if errWd != nil {
		log.Debug("Error getting current working dir", errWd)
	} else {
		directories = append(directories, localDir)
	}
	exeDir, errExe := osExecutable()
	if errExe != nil {
		log.Debug("Error getting current exe dir", errExe)
	} else {
		directories = append(directories, exeDir)
	}
	return directories
}

//CollectFileStatus returns a struct that provides information about the file we are going to collect from customer's environment: fullpath to the file, verifying if the path provided (by env var or config file) takes us to file that exists, and if doesn't we collect the error message to bubble up to the customer.
type CollectFileStatus struct {
	Path     string
	IsValid  bool
	ErrorMsg error
}

//ValidatePaths takes an slice of strings to check if customer is providing us with paths(that come from either env var or config file) from which can collect a file. It returns a slice of FileToCollect which informs if a file is invalid and the error we found
func ValidatePaths(paths []string) []CollectFileStatus {
	var filesToCollect []CollectFileStatus
	for _, path := range paths {
		filesToCollect = append(filesToCollect, ValidatePath(path))
	}
	return filesToCollect
}

//ValidatePath takes a string to check if customer is providing us with paths(that come from either env var or config file) from which can collect a file. It returns a FileToCollect which informs if the file is invalid and the error we found
func ValidatePath(path string) CollectFileStatus {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return CollectFileStatus{path, false, err}
	}

	if fileInfo.IsDir() {
		return CollectFileStatus{path, false, errors.New("Is directory and not a path to a file")}
	}

	file, err := os.Open(path)
	if err != nil {
		return CollectFileStatus{path, false, err}
	}

	file.Close()
	return CollectFileStatus{path, true, nil}
}
