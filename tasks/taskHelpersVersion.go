package tasks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Ver is an exported type for storing a software version.
// It contains integers for Major, Minor, Patch, and Build.
type Ver struct {
	Major int
	Minor int
	Patch int
	Build int
}

// VersionIsCompatibleFunc - allows VersionIsCompatible to be dependency injected
type VersionIsCompatibleFunc func(string, []string) (bool, error)

// VersionIsCompatible checks a version against a slice of compatibility requirements.
// It accepts as input a string version (one or more numbers separated by periods) and a slice of compatibility requirements.
// Individual compatibility requirements can be expressed as a range ("4.0-7.4"), as a single ("8.0"), or a minimum ("4.9+")
// It returns a boolean indicating compatibility and an error in cases of invalid input.
func VersionIsCompatible(version string, requirements []string) (bool, error) {

	versionToCheck, err := ParseVersion(version)
	if err != nil {
		return false, errors.New("unable to parse version: " + version)
	}
	return versionToCheck.CheckCompatibility(requirements)
}

// CheckCompatibility takes a slice of string compatibility requirements and
// returns a bool indicating whether the receiver is compatible.
// Any unsupplied version components will default to zero, e.g. "8" becomes "8.0.0.0"
// Individual compatibility requirements can be expressed as follows:
// Inclusive Range ("4.0-7.4") - returns true for "4.11" and "7.4.0", false for "7.4.1"
// Range with wildcard upper bound: ("4.0-7.4.*") - returns true for versions "4.11" and true for "7.4.1"
// Minimum version: ("7.4+") - true for version "7.4" and true for "10.0"
// Single wildcard version ("7.*") - Returns true for any version with 7 major version ("7.9.1")
// Single version only ("7") - Match this version only:  matches "7", "7.0", "7.0.0" but not "7.1" or "7.0.2"

func (v Ver) CheckCompatibility(requirements []string) (bool, error) {

	for _, requirement := range requirements {
		// Convert plus sign to lower boundary + infinity for upper boundary
		if strings.Contains(requirement, "+") {
			requirement = strings.Replace(requirement, "+", "", -1) + "-" + "infinity"
		}

		// Extra logic to adapt single-component cases to range input e.g. "5" becomes "5-5"
		if !(strings.Contains(requirement, "-") || strings.Contains(requirement, "*")) {
			requirement = requirement + "-" + requirement
		} else if (!strings.Contains(requirement, "-")) && strings.Contains(requirement, "*") {
			requirement = strings.Replace(requirement, ".*", "", -1) + "-" + requirement
		}

		reqsToCheck := strings.Split(requirement, "-")
		// Handle cases like 5-7.5+ or the ever-possible 3-6-9.4
		if len(reqsToCheck) > 2 {
			return false, errors.New("received a range with too many values: " + requirement)
		}

		minReq, err := ParseVersion(reqsToCheck[0])
		if err != nil {
			return false, errors.New("unable to parse version: " + reqsToCheck[0])
		}

		maxReq, err := ParseVersion(reqsToCheck[1])
		if err != nil {
			return false, errors.New("unable to parse version: " + reqsToCheck[1])
		}
		if v.IsGreaterThanEq(minReq) && v.IsLessThanEq(maxReq) {
			return true, nil
		}

	}
	return false, nil
}

// IsGreaterThanEq takes a Ver as input and compares it to its receiver and
// returns true if the receiver version is greater than or equal to the
// supplied version.
func (v Ver) IsGreaterThanEq(min Ver) bool {
	if v.Major > min.Major {
		return true
	} else if v.Major == min.Major {
		if v.Minor > min.Minor {
			return true
		} else if v.Minor == min.Minor {
			if v.Patch > min.Patch {
				return true
			} else if v.Patch == min.Patch {
				if v.Build >= min.Build {
					return true
				}
			}
		}
	}
	return false
}

// IsLessThanEq takes a Ver as input and compares it to its receiver and
// returns true if the receiver version is less than or equal to the
// supplied version.
func (v Ver) IsLessThanEq(max Ver) bool {
	if v.Major < max.Major {
		return true
	} else if v.Major == max.Major {
		if v.Minor < max.Minor {
			return true
		} else if v.Minor == max.Minor {
			if v.Patch < max.Patch {
				return true
			} else if v.Patch == max.Patch {
				if v.Build <= max.Build {
					return true
				}
			}
		}
	}
	return false
}

//String() returns string representation of version struct: Maj.Min.P.B
func (v Ver) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// ParseVersion takes a string representation of a software version and
// returns a Ver struct instance and nil, or an empty Ver struct and an error.
func ParseVersion(version string) (Ver, error) {

	if version == "infinity" {
		return Ver{
			Major: int(^uint(0) >> 1),
			Minor: int(^uint(0) >> 1),
			Build: int(^uint(0) >> 1),
			Patch: int(^uint(0) >> 1),
		}, nil
	}
	emptyVersion := Ver{0, 0, 0, 0} // returned in error cases
	parsedVersion := emptyVersion   //  placeholder for successful output

	// 4.5.6 -> [4,5,6]
	vArr := strings.Split(strings.TrimSpace(version), ".")
	var (
		intNum int
		ierr   error
	)

	for i, num := range vArr {
		if num == "*" {
			// infinity for our purposes
			intNum = int(^uint(0) >> 1)
		} else {
			intNum, ierr = strconv.Atoi(num)
			if ierr != nil {
				return emptyVersion, errors.New("unable to convert " + num + " to an integer")
			}
		}

		switch i {
		case 0:
			parsedVersion.Major = intNum
		case 1:
			parsedVersion.Minor = intNum
		case 2:
			parsedVersion.Patch = intNum
		case 3:
			parsedVersion.Build = intNum
		}
	}
	return parsedVersion, nil
}

// VersionsJoin takes in a slice of Ver structs and delimiter and returns
// a string of all versions in the slice, joined on that delimiter
func VersionsJoin(versions []Ver, delimiter string) string {
	var temp []string

	for _, v := range versions {
		temp = append(temp, v.String())
	}

	return strings.Join(temp, delimiter)
}
