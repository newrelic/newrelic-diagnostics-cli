package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"strings"

	"github.com/google/uuid"
	"github.com/newrelic/NrDiag/config"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

type override struct {
	tasks.Identifier
	key   string
	value string
}

func parseOverrides(overrides string) []override {
	var sliceOverrides []override

	//First we need to split on a comma to get each override provided
	splitOverrides := strings.Split(overrides, ",")
	if len(splitOverrides[0]) != 0 { //Check for an empty override
		for _, eachOverride := range splitOverrides {
			log.Debug("pre-parsing override is :", eachOverride)
			//Now split on equals to get the value
			taskValue := strings.Split(eachOverride, "=")
			if len(taskValue) > 1 {
				//And again to get the identifier and key
				taskKey := strings.Split(taskValue[0], ".")
				if len(taskKey) > 1 {
					log.Debug("Override Identifier is ", tasks.IdentifierFromString(taskKey[0]))
					log.Debug("Override Key is ", taskKey[1])
					log.Debug("Override Value is", taskValue[1])

					overridee := override{tasks.IdentifierFromString(taskKey[0]), taskKey[1], taskValue[1]}
					//Now add it to the list
					sliceOverrides = append(sliceOverrides, overridee)
				}
			}
		}
	}

	return sliceOverrides
}

func ProxyParseNSet() (set bool) {
	//This sets the default
	var DefaultDialer = &net.Dialer{Timeout: 1000 * time.Millisecond}
	http.DefaultTransport = &http.Transport{Dial: DefaultDialer.Dial, Proxy: http.ProxyFromEnvironment}
	return true

}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

//similar to tasks - This takes the input string as the query to the end users and waits for a response
func promptUser(msg string) bool {
	if config.Flags.YesToAll {
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

		if tasks.ContainsString(noResponses, userInput) {
			return false
		}

		if tasks.ContainsString(yesResponses, userInput) {
			return true
		}

		//Repeat prompt if invalid input is provided
		fmt.Printf("%s", prompt)
	}
	return false
}

func generateRunID() string {
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Debug("Error generating UUID", err)
		return ""
	}

	return uuid.String()
}
