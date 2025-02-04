package suites

import (
	"strings"
)

type Suite struct {
	Identifier  string   //java
	DisplayName string   // Java Agent
	Description string   //Optional if display name is not intuitive
	Tasks       []string //TaskIdentifier Strings
}

type SuiteManager struct {
	Suites         []Suite
	SelectedSuites []Suite
}

func (s *SuiteManager) AddSelectedSuite(selectedSuite Suite) {
	s.SelectedSuites = append(s.SelectedSuites, selectedSuite)
}

func (s *SuiteManager) AddSelectedSuites(selectedSuites []Suite) {
	for _, selectedSuite := range selectedSuites {
		s.AddSelectedSuite(selectedSuite)
	}
}

func (s SuiteManager) FindSuiteByIdentifier(suiteIdentifier string) (Suite, bool) {
	sanitizedIdentifier := strings.TrimSpace(suiteIdentifier)

	for _, suite := range s.Suites {
		if strings.EqualFold(suite.Identifier, sanitizedIdentifier) {
			return suite, true
		}
	}

	return Suite{}, false
}

func (s SuiteManager) FindSuitesByIdentifiers(suiteIdentifiers []string) ([]Suite, []string) {
	unMatchedSuites := []string{}
	matchedSuites := []Suite{}

	for _, suiteIdentifier := range suiteIdentifiers {

		suite, ok := s.FindSuiteByIdentifier(suiteIdentifier)

		if !ok {
			unMatchedSuites = append(unMatchedSuites, suiteIdentifier)
			continue
		}

		matchedSuites = append(matchedSuites, suite)
	}

	return matchedSuites, unMatchedSuites
}

func (s SuiteManager) CaptureOutOfPlaceArgs(osArgs []string, suiteFlagArgs []string) []string {

	var extraArgs []string
	lastSuiteArg := suiteFlagArgs[len(suiteFlagArgs)-1]

	for index, osArg := range osArgs {
		//We want to ignore the first argument because in windows it contains the current directory name,
		//and then in linux & darwin it's just the executable which is also not a useful arg to evaluate here.
		if index == 0 {
			continue
		}
		if strings.Contains(osArg, lastSuiteArg) {
			extraArgs = osArgs[index+1:]
			break
		}
	}
	orphanedSuites, _ := s.FindSuitesByIdentifiers(extraArgs)

	var argsMatchedToSuitesIdentifiers []string

	for _, suite := range orphanedSuites {
		argsMatchedToSuitesIdentifiers = append(argsMatchedToSuitesIdentifiers, suite.Identifier)
	}

	return argsMatchedToSuitesIdentifiers
}

// FindTasksBySuites - When given a slice of Suite structs, it will return a slice of task identifier strings included in those suites.
func (s SuiteManager) FindTasksBySuites(suites []Suite) []string {
	var tasks []string
	for _, suite := range suites {
		tasks = append(tasks, suite.Tasks...)
	}
	return tasks
}

// DefaultSuiteManager - Eventually we'll want to move this to an app dependency struct
var DefaultSuiteManager = NewSuiteManager(suiteDefinitions)

func NewSuiteManager(suites []Suite) *SuiteManager {

	return &SuiteManager{
		Suites: suites,
	}
}
