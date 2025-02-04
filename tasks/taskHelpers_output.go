package tasks

import "github.com/newrelic/newrelic-diagnostics-cli/output/color"

func (s Status) GetColor() color.Color {
	switch s {
	case Success:
		return color.LightGreen
	case Warning:
		return color.LightYellow
	case Error, Failure:
		return color.LightRed
	case None:
		return color.LightBlue
	case Info:
		return color.White
	default:
		return color.Clear
	}
}

// StatusToString takes in integer, returns relevant Status statusEnum in a human readable string.
func (s Status) StatusToString() string {
	statuses := []string{"None", "Success", "Warning", "Failure", "Error", "Info"}
	return statuses[s]
}

// StatusToString returns the status for a result, used primarily for visual output
func (r Result) StatusToString() string {
	return r.Status.StatusToString()
}
