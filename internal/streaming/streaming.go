package streaming

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/newrelic/newrelic-client-go/v2/newrelic"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdstorage"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
)

type NSDocument struct {
	Results string
	Done    bool
}

type results struct {
	Results []registration.TaskResult
}

var apiClient *newrelic.NewRelic
var fullResults []string
var taskResults []registration.TaskResult
var initialized bool
var packageId = "00000000-0000-0000-0000-000000000000"
var collection = "cli_stream"
var documentId = "cli_stream_doc"

func Initialize(apiKey string) error {
	if apiKey == "" {
		return errors.New("no apikey")
	}
	client, err := newrelic.New(newrelic.ConfigPersonalAPIKey(apiKey), newrelic.ConfigRegion("STAGING"))
	if err != nil {
		return err
	}
	apiClient = client
	delete()
	initialized = true
	return nil
}

func WriteLine(line string) {
	if !initialized {
		return
	}
	fullResults = append(fullResults, line)
	input := nerdstorage.WriteDocumentInput{
		PackageID:  packageId,
		Collection: collection,
		DocumentID: documentId,
		Document:   NSDocument{Results: strings.Join(fullResults, "\n")},
	}
	_, err := apiClient.NerdStorage.WriteDocumentWithUserScope(input)
	if err != nil {
		log.Infof("Error streaming: %s", err.Error())
	}
}

func WriteTask(taskResult registration.TaskResult) {
	if !initialized {
		return
	}
	taskResults = append(taskResults, taskResult)
	streamData := results{
		Results: taskResults,
	}
	json, _ := json.MarshalIndent(streamData, "", "	")
	input := nerdstorage.WriteDocumentInput{
		PackageID:  packageId,
		Collection: collection,
		DocumentID: documentId,
		Document:   NSDocument{Results: string(json), Done: false},
	}
	_, err := apiClient.NerdStorage.WriteDocumentWithUserScope(input)
	if err != nil {
		log.Infof("Error streaming: %s", err.Error())
	}
}

func Close() {
	if !initialized {
		return
	}
	streamData := results{
		Results: taskResults,
	}
	json, _ := json.MarshalIndent(streamData, "", "	")
	input := nerdstorage.WriteDocumentInput{
		PackageID:  packageId,
		Collection: collection,
		DocumentID: documentId,
		Document:   NSDocument{Results: string(json), Done: true},
	}
	_, err := apiClient.NerdStorage.WriteDocumentWithUserScope(input)
	if err != nil {
		log.Infof("Error streaming: %s", err.Error())
	}
}

func delete() {
	input := nerdstorage.DeleteCollectionInput{PackageID: packageId, Collection: collection}
	apiClient.NerdStorage.DeleteCollectionWithUserScope(input)
}
