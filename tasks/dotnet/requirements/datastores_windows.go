package requirements

import (
	"path/filepath"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// DotnetRequirementsDatastores - This struct defines the .Net datastores check
type DotnetRequirementsDatastores struct {
	findFiles             func([]string, []string) []string
	getWorkingDirectories tasks.GetWorkingDirectoriesFunc
	getFileVersion        tasks.GetFileVersionFunc
}

//This data type will store and track info on the different datastores
//This allows for the logic to be agnostic to the requirements of the
//individual datastore requrements, only the structs for a database should need to change if requirements change
type dataStore struct {
	name        string
	version     string
	installed   bool
	versionGood bool
}

//Datastore names and slice defined here for easier updating if agent requirements change and passing between funcs

//Names of dlls to check.
//No check for MS SQL as it's in the system.data.dll,
//which can be there even with out MS SQL being used by the app.
var couchBaseName = "CouchbaseNetClient.dll"
var couchBaseVersion = []string{"2+"}
var ibmDb2Name = "IBM.Data.DB2.dll"
var mongoName = "MongoDB.Driver.dll"
var mongoVersion = []string{"1.0-1.10", "2.3-2.7.*"}

var mySQLName = "MySql.Data.dll"
var oracleName = "Oracle.ManagedDataAccess.dll"
var npgsqlName = "Npgsql.dll"
var serviceStackRedisName = "ServiceStack.Redis.dll"
var stackExchangeRedisName = "StackExchange.Redis.dll"

// Identifier - This returns the Category, Subcategory and Name of this task
func (p DotnetRequirementsDatastores) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/Datastores")
}

// Explain - Returns the help text for this task
func (p DotnetRequirementsDatastores) Explain() string {
	return "Check database version compatibility with New Relic .NET agent"
}

// Dependencies - Returns the dependencies for this task.
func (p DotnetRequirementsDatastores) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within this task
func (p DotnetRequirementsDatastores) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {

		return tasks.Result{
			Status:  tasks.None,
			Summary: ".Net Agent not installed, this task didn't run",
		}
	}

	dllsFound := p.findDlls()

	//If no dlls found then return none
	if len(dllsFound) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Did not find any dlls associated with Datastores supported by the .Net Agent",
		}
	}

	//process dlls found
	//and populate the structs with additional info
	installedDatastores := p.processFoundDatastoreDlls(dllsFound)
	cntNoVersion := 0
	cntBadVersion := 0
	var summarySlice []string
	var summary string
	var payload []string
	//Need to check the results and figure out what status to set
	for _, ds := range installedDatastores {
		if !ds.versionGood {
			if ds.version == "" {
				temp := "Unable to find version for " + ds.name
				payload = append(payload, temp)
				summarySlice = append(summarySlice, temp)

				cntNoVersion++
				continue
			}
			temp := "Incompatible version of " + ds.name + " detected. Found version " + ds.version
			summarySlice = append(summarySlice, temp)
			payload = append(payload, "Found "+ds.name+" with version "+ds.version)
			cntBadVersion++
			continue
		}
		payload = append(payload, "Found "+ds.name+" with version "+ds.version)

	}

	if cntBadVersion == 0 && cntNoVersion == 0 {
		//sucess
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "All datastores detected as compatible, see plugin.json for more details.",
			Payload: payload,
		}
	} else if cntBadVersion > 0 {
		//failure

		summary = "Incompatible datastores detected, see plugin.json for more details. Detected the following datastores as incompatible: \n"
		for _, s := range summarySlice {
			summary = summary + s + "\n"
		}

		return tasks.Result{
			Status:  tasks.Failure,
			Summary: summary,
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#database",
			Payload: payload,
		}
	}

	// BadVersion 0 but NoVersion > 0, give a warning
	summary = "Couldn't get version of some datastore dlls: \n"
	for _, s := range summarySlice {
		summary = summary + s + "\n"
	}
	return tasks.Result{
		Status:  tasks.Warning,
		Summary: summary,
		URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#database",
		Payload: payload,
	}
}

//Most data stores will add one or more dlls to an apps bin dir
//Checking for these is a strong indicator that an app is using the db
func (p DotnetRequirementsDatastores) findDlls() []string {

	dataStoreNames := []string{
		couchBaseName,
		ibmDb2Name,
		mongoName,
		mySQLName,
		oracleName,
		npgsqlName,
		serviceStackRedisName,
		stackExchangeRedisName,
	}

	localDirs := p.getWorkingDirectories()

	dllsFound := p.findFiles(dataStoreNames, localDirs)

	return dllsFound
}

func (p DotnetRequirementsDatastores) processFoundDatastoreDlls(dllsFound []string) []dataStore {

	var installedDatastores []dataStore

	for _, dll := range dllsFound {
		fileName := filepath.Base(dll)
		fileVersion, err := p.getFileVersion(dll)
		if err != nil {
			log.Debug("Error getting file version for ", dll, " ", err)
			installedDatastores = append(installedDatastores, dataStore{name: fileName, installed: true, versionGood: false})
			continue
		}

		compatible := false
		// check version for CouchBase and Mongo
		switch fileName {
		case couchBaseName:
			compatible = checkCouchBaseVer(fileVersion)
		case mongoName:
			compatible = checkMongoVer(fileVersion)
		default:
			compatible = true
		}
		installedDatastores = append(installedDatastores, dataStore{name: fileName, version: fileVersion, installed: true, versionGood: compatible})

	}

	return installedDatastores

}

//Define functions to check version below
func checkMongoVer(version string) bool {
	compatible, _ := tasks.VersionIsCompatible(version, mongoVersion)
	return compatible
}
func checkCouchBaseVer(version string) bool {
	compatible, _ := tasks.VersionIsCompatible(version, couchBaseVersion)
	return compatible
}
