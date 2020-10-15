package version

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/newrelic/NrDiag/config"
	"github.com/newrelic/NrDiag/helpers/httpHelper"
	"github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/output/color"
)

const versionURL = `http://download.newrelic.com/nrdiag/version.txt`
const downloadURL = `http://download.newrelic.com/nrdiag/nrdiag_latest.zip`

// ProcessAutoVersionCheck - looks at the program version and warnds the user if it is out of date, takes no actions
func ProcessAutoVersionCheck() {
	processAutoVersionCheck(logger.Log, getOnlineVersion)
}

// ProcessVersion - looks at the program version and warnds the user if it is out of date, prompts user and is able to download
func ProcessVersion(promptUser func(string) bool) {
	processVersion(logger.Log, promptUser, getOnlineVersion, getLatestVersion)
}

func getOnlineVersion(log logger.API) string {
	var version string

	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            versionURL,
		TimeoutSeconds: 5,
	}

	resp, err := httpHelper.MakeHTTPRequest(wrapper)
	if err != nil {
		log.Info("Error checking latest version: ", err)
		return ""
	}

	responseBody, er := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if er != nil {
		log.Info("error reading file download", er)
		return ""
	}

	//Remove newline character from the version.txt file
	version = strings.TrimSuffix(string(responseBody), "\n")
	return version
}

func getLatestVersion(log logger.API) error {

	// Create the file
	out, err := os.Create("nrdiag_latest.zip")
	if err != nil {
		log.Info("Error downloading file,", err)
		// should panic without `return` here
		return err
	}
	defer out.Close()

	wrapper := httpHelper.RequestWrapper{
		Method: "GET",
		URL:    downloadURL,
	}

	resp, httpErr := httpHelper.MakeHTTPRequest(wrapper)
	if httpErr != nil {
		log.Info("Error downloading file,", httpErr)
		return httpErr
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Info("Error writing file", err)
		return err
	}

	return nil
}

func logVersionString(log logger.API) {
	if !config.Flags.VeryQuiet {
		log.Infof("New Relic Diagnostics - release version - %s - Build timestamp - %s\n", config.Version, config.BuildTimestamp)
	}
}

func processAutoVersionCheck(log logger.API, getOnlineVersion func(logger.API) string) {
	onlineVersion := getOnlineVersion(log)
	if onlineVersion != "" && onlineVersion != config.Version {
		if !config.Flags.VeryQuiet {
			logVersionString(log)
			log.Infof(color.ColorString(color.Yellow, "Version %s has been released and is newer than your version.\n"), onlineVersion)
			log.Info("Please run with the -version flag to download the new release.")
		}
	}
}

func processVersion(log logger.API, promptUser func(input string) bool, getOnlineVersion func(logger.API) string, getLatestVersion func(logger.API) error) {
	logVersionString(log)

	if config.Flags.SkipVersionCheck {
		return
	}

	// Ask the user if they want to check for a newer version
	if promptUser("Do you want to check for a newer version?") {
		if !config.Flags.Quiet {
			log.Info("Checking for newer version")
		}
		onlineVersion := getOnlineVersion(log)
		if !config.Flags.VeryQuiet {
			log.Info("Online version found:", onlineVersion)
		}

		if onlineVersion != config.Version {

			if promptUser("Do you want to download the latest version?") {
				if !config.Flags.Quiet {
					log.Info("Downloading latest version")
				}
				err := getLatestVersion(log)
				if !config.Flags.Quiet {
					if err == nil {
						log.Infof("Version %s downloaded and saved as nrdiag_latest.zip\n", onlineVersion)
					}
				}
			}
		} else {
			if !config.Flags.VeryQuiet {
				log.Info("Already running the latest version", onlineVersion)
			}
		}
	}

}
