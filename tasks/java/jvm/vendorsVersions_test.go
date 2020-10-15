package jvm

import (
	"errors"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
	"github.com/newrelic/NrDiag/tasks"
)

func TestJavaJVMVendorsVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Java/JVM/* test suite")

}
func TestExtractVersionFromArgs(t *testing.T) {
	matchVerTests := []struct {
		expectedVersion string
		cmdLineArgs     string
	}{
		{expectedVersion: "1.6", cmdLineArgs: "-Djava.util.logging.manager=com.ibm.ws.bootstrap.WsLogManager, -Djava.runtime.version=pap6460sr16fp1ifx-20140908_01 (SR16 FP1), -Djavax.management.builder.initial=com.ibm.ws.management.PlatformMBeanServerBuilder"},
		{expectedVersion: "1.8.0_101", cmdLineArgs: "-Djava.runtime.version=1.8.0_101-b04"},
		{expectedVersion: "1.6.0_17", cmdLineArgs: "-Djava.vm.version=R28.0.0-617-125986-1.6.0_17-20091215-2120-windows-x86_64, compiled mode"},
		{expectedVersion: "", cmdLineArgs: "9-ea+138-jigsaw-nightly-h5561-20161003"},
	}
	for _, test := range matchVerTests {
		actualResult := extractVersionFromArgs(test.cmdLineArgs)
		if test.expectedVersion != actualResult {
			t.Errorf("Test failed. Version extracted is %s. Result expected is %v. Cmd Line Args are %s", actualResult, test.expectedVersion, test.cmdLineArgs)
		}
	}
}

func TestExtractVendorFromArgs(t *testing.T) {
	matchVenTests := []struct {
		expectedVendor string
		cmdLineArgs    string
	}{
		{expectedVendor: "IBM", cmdLineArgs: "-Djava.vm.name=IBM J9 VM"},
		{expectedVendor: "HotSpot", cmdLineArgs: "-Djava.vm.name=Java HotSpot(TM) Client VM"},
		{expectedVendor: "Oracle", cmdLineArgs: "-Djava.vm.vendor=Oracle Corporation"},
		{expectedVendor: "", cmdLineArgs: "-Djava.vm.vendor=Sun Microsystems Inc."},
	}

	for _, test := range matchVenTests {
		actualResult := extractVendorFromArgs(test.cmdLineArgs)
		if actualResult != test.expectedVendor {
			t.Errorf("Test failed. Vendor extracted is %s. Result expected is %v. Cmd Line Args are %s", actualResult, test.expectedVendor, test.cmdLineArgs)
		}
	}
}

func TestCheckSupported(t *testing.T) {
	type testStruct struct {
		vendor  string
		version string
		result  bool
		os      string
	}
	p := JavaJVMVendorsVersions{}

	matchFullySupportedTests := []testStruct{

		// Compatibility tests taken from:
		// https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent

		// Updated: 9/23/2019

		//Azul Zing version 8 to 11 for linux, darwin and windows
		{os: "linux", vendor: "Zulu", version: "1.7.0.152", result: false},
		{os: "linux", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "darwin", vendor: "Zulu", version: "1.7.0.152", result: false},
		{os: "darwin", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "windows", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "windows", vendor: "Zulu", version: "1.7.0.152", result: false},

		// IBM JVM versions 7 and 8 for Linux
		{os: "linux", vendor: "IBM", version: "1.7.5", result: true},
		{os: "linux", vendor: "IBM", version: "1.8.1", result: true},

		{os: "windows", vendor: "IBM", version: "1.8.1", result: false},
		{os: "darwin", vendor: "IBM", version: "1.8.1", result: false},

		// OpenJDK JVM versions 7 to 13 for Linux, Windows, and OS X
		{os: "linux", vendor: "OpenJDK", version: "1.7.0.152", result: true},
		{os: "linux", vendor: "OpenJDK", version: "1.8", result: true},
		{os: "linux", vendor: "OpenJDK", version: "9", result: true},
		{os: "linux", vendor: "OpenJDK", version: "10.0", result: true},
		{os: "linux", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "linux", vendor: "OpenJDK", version: "12", result: true},

		{os: "windows", vendor: "OpenJDK", version: "1.7.0.152", result: true},
		{os: "windows", vendor: "OpenJDK", version: "1.8", result: true},
		{os: "windows", vendor: "OpenJDK", version: "9", result: true},
		{os: "windows", vendor: "OpenJDK", version: "10.0", result: true},
		{os: "windows", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "windows", vendor: "OpenJDK", version: "12", result: true},

		{os: "darwin", vendor: "OpenJDK", version: "1.7.0.152", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "1.8", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "9", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "10.0", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "12", result: true},

		{os: "darwin", vendor: "OpenJDK", version: "1.5.0.152", result: false},
		{os: "darwin", vendor: "OpenJDK", version: "1.6", result: false},
		{os: "darwin", vendor: "OpenJDK", version: "5", result: false},
		{os: "darwin", vendor: "OpenJDK", version: "5", result: false},
		{os: "darwin", vendor: "OpenJDK", version: "3", result: false},

		// Oracle HotSpot JVM versions 7 to 12 for Linux, Solaris, Windows, and OS X
		{os: "linux", vendor: "HotSpot", version: "1.7.0.152", result: true},
		{os: "linux", vendor: "HotSpot", version: "1.8", result: true},
		{os: "linux", vendor: "HotSpot", version: "7", result: true},
		{os: "linux", vendor: "HotSpot", version: "8", result: true},
		{os: "linux", vendor: "HotSpot", version: "9", result: true},
		{os: "linux", vendor: "HotSpot", version: "10.0", result: true},
		{os: "linux", vendor: "HotSpot", version: "11.0.1", result: true},
		{os: "linux", vendor: "HotSpot", version: "12", result: true},

		{os: "windows", vendor: "HotSpot", version: "1.7.0.152", result: true},
		{os: "windows", vendor: "HotSpot", version: "1.8", result: true},
		{os: "windows", vendor: "HotSpot", version: "7", result: true},
		{os: "windows", vendor: "HotSpot", version: "8", result: true},
		{os: "windows", vendor: "HotSpot", version: "9", result: true},
		{os: "windows", vendor: "HotSpot", version: "10.0", result: true},
		{os: "windows", vendor: "HotSpot", version: "11.0.1", result: true},
		{os: "windows", vendor: "HotSpot", version: "12", result: true},

		{os: "darwin", vendor: "HotSpot", version: "1.7.0.152", result: true},
		{os: "darwin", vendor: "HotSpot", version: "1.8", result: true},
		{os: "darwin", vendor: "HotSpot", version: "7", result: true},
		{os: "darwin", vendor: "HotSpot", version: "8", result: true},
		{os: "darwin", vendor: "HotSpot", version: "9", result: true},
		{os: "darwin", vendor: "HotSpot", version: "10.0", result: true},
		{os: "darwin", vendor: "HotSpot", version: "11.0.1", result: true},
		{os: "darwin", vendor: "HotSpot", version: "12", result: true},

		// Amazon Coretto JVM versions 8-11 for Linux, Windows, and OS X
		{os: "linux", vendor: "Coretto", version: "1.7.0.152", result: false},
		{os: "linux", vendor: "Coretto", version: "1.8", result: true},
		{os: "linux", vendor: "Coretto", version: "8", result: true},
		{os: "linux", vendor: "Coretto", version: "9", result: true},
		{os: "linux", vendor: "Coretto", version: "10.0", result: true},
		{os: "linux", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "linux", vendor: "Coretto", version: "12", result: false},

		{os: "windows", vendor: "Coretto", version: "1.7.0.152", result: false},
		{os: "windows", vendor: "Coretto", version: "1.8", result: true},
		{os: "windows", vendor: "Coretto", version: "8", result: true},
		{os: "windows", vendor: "Coretto", version: "9", result: true},
		{os: "windows", vendor: "Coretto", version: "10.0", result: true},
		{os: "windows", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "windows", vendor: "Coretto", version: "12", result: false},

		{os: "darwin", vendor: "Coretto", version: "1.7.0.152", result: false},
		{os: "darwin", vendor: "Coretto", version: "1.8", result: true},
		{os: "darwin", vendor: "Coretto", version: "8", result: true},
		{os: "darwin", vendor: "Coretto", version: "9", result: true},
		{os: "darwin", vendor: "Coretto", version: "10.0", result: true},
		{os: "darwin", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "darwin", vendor: "Coretto", version: "12", result: false},

		// Azul Zulu JVM versions 8 to 12 for Linux, Windows, and OS X
		{os: "linux", vendor: "Zulu", version: "1.7.0.152", result: false},
		{os: "linux", vendor: "Zulu", version: "7", result: false},
		{os: "linux", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "linux", vendor: "Zulu", version: "9", result: true},
		{os: "linux", vendor: "Zulu", version: "10.0", result: true},
		{os: "linux", vendor: "Zulu", version: "11.0.1", result: true},
		{os: "linux", vendor: "Zulu", version: "12", result: true},

		{os: "windows", vendor: "Zulu", version: "1.7.0.152", result: false},
		{os: "windows", vendor: "Zulu", version: "7", result: false},
		{os: "windows", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "windows", vendor: "Zulu", version: "9", result: true},
		{os: "windows", vendor: "Zulu", version: "10.0", result: true},
		{os: "windows", vendor: "Zulu", version: "11.0.1", result: true},
		{os: "windows", vendor: "Zulu", version: "12", result: true},

		{os: "darwin", vendor: "Zulu", version: "1.7.0.152", result: false},
		{os: "darwin", vendor: "Zulu", version: "7", result: false},
		{os: "darwin", vendor: "Zulu", version: "1.8.0.152", result: true},
		{os: "darwin", vendor: "Zulu", version: "9", result: true},
		{os: "darwin", vendor: "Zulu", version: "10.0", result: true},
		{os: "darwin", vendor: "Zulu", version: "11.0.1", result: true},
		{os: "darwin", vendor: "Zulu", version: "12", result: true},

		//Linux tests

		//Only supported in linux
		{os: "linux", vendor: "IBM", version: "1.7.0.152", result: true},
		{os: "linux", vendor: "IBM", version: "7", result: true},
		{os: "linux", vendor: "IBM", version: "1.8", result: true},
		{os: "linux", vendor: "IBM", version: "8", result: true},
		{os: "linux", vendor: "IBM", version: "9", result: false},
		{os: "linux", vendor: "IBM", version: "10.0", result: false},
		{os: "linux", vendor: "IBM", version: "11.0.1", result: false},
		{os: "linux", vendor: "IBM", version: "12", result: false},

		// Failures across the board for windows/darwin
		{os: "windows", vendor: "IBM", version: "1.7.0.152", result: false},
		{os: "windows", vendor: "IBM", version: "7", result: false},
		{os: "windows", vendor: "IBM", version: "1.8", result: false},
		{os: "windows", vendor: "IBM", version: "8", result: false},
		{os: "windows", vendor: "IBM", version: "9", result: false},
		{os: "windows", vendor: "IBM", version: "10.0", result: false},
		{os: "windows", vendor: "IBM", version: "11.0.1", result: false},
		{os: "windows", vendor: "IBM", version: "12", result: false},

		{os: "darwin", vendor: "IBM", version: "1.7.0.152", result: false},
		{os: "darwin", vendor: "IBM", version: "7", result: false},
		{os: "darwin", vendor: "IBM", version: "1.8", result: false},
		{os: "darwin", vendor: "IBM", version: "8", result: false},
		{os: "darwin", vendor: "IBM", version: "9", result: false},
		{os: "darwin", vendor: "IBM", version: "10.0", result: false},
		{os: "darwin", vendor: "IBM", version: "11.0.1", result: false},
		{os: "darwin", vendor: "IBM", version: "12", result: false},

		//Should be supported in all OSes
		{os: "linux", vendor: "OpenJDK", version: "1.8.0.152", result: true},
		{os: "linux", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "linux", vendor: "OpenJDK", version: "1.6.0", result: false},
		{os: "linux", vendor: "HotSpot", version: "1.7.0.16", result: true},
		{os: "linux", vendor: "HotSpot", version: "1.8.0.161", result: true},
		{os: "linux", vendor: "HotSpot", version: "Unknown", result: false},
		{os: "linux", vendor: "JRockit", version: "1.6.0.49", result: true},
		{os: "linux", vendor: "JRockit", version: "1.6.0.101", result: false},
		{os: "linux", vendor: "JRockit", version: "9.0.4", result: false},
		{os: "linux", vendor: "Coretto", version: "1.8.1", result: true},
		{os: "linux", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "linux", vendor: "Zulu", version: "12.1.3", result: true},
		{os: "linux", vendor: "Unknown", version: "Unknown", result: false},
		{os: "linux", vendor: "J9", version: "1.5.0", result: false},
		{os: "linux", vendor: "J9", version: "9", result: false},

		//Windows tests
		{os: "windows", vendor: "OpenJDK", version: "1.8.0.152", result: true},
		{os: "windows", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "windows", vendor: "OpenJDK", version: "1.6.0", result: false},
		{os: "windows", vendor: "HotSpot", version: "1.7.0.16", result: true},
		{os: "windows", vendor: "HotSpot", version: "1.8.0.161", result: true},
		{os: "windows", vendor: "HotSpot", version: "Unknown", result: false},
		{os: "windows", vendor: "JRockit", version: "1.6.0.49", result: true},
		{os: "windows", vendor: "JRockit", version: "1.6.0.101", result: false},
		{os: "windows", vendor: "JRockit", version: "9.0.4", result: false},
		{os: "windows", vendor: "Coretto", version: "1.8.1", result: true},
		{os: "windows", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "windows", vendor: "Zulu", version: "12.1.3", result: true},
		{os: "windows", vendor: "Unknown", version: "Unknown", result: false},
		{os: "windows", vendor: "J9", version: "1.5.0", result: false},
		{os: "windows", vendor: "J9", version: "9", result: false},
		{os: "windows", vendor: "IBM", version: "1.7.1", result: false},

		//Darwin tests
		{os: "darwin", vendor: "OpenJDK", version: "1.8.0.152", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "11.0.1", result: true},
		{os: "darwin", vendor: "OpenJDK", version: "1.6.0", result: false},
		{os: "darwin", vendor: "HotSpot", version: "1.7.0.16", result: true},
		{os: "darwin", vendor: "HotSpot", version: "1.8.0.161", result: true},
		{os: "darwin", vendor: "HotSpot", version: "Unknown", result: false},
		{os: "darwin", vendor: "JRockit", version: "1.6.0.49", result: true},
		{os: "darwin", vendor: "JRockit", version: "1.6.0.101", result: false},
		{os: "darwin", vendor: "JRockit", version: "9.0.4", result: false},
		{os: "darwin", vendor: "Coretto", version: "1.8.1", result: true},
		{os: "darwin", vendor: "Coretto", version: "11.0.1", result: true},
		{os: "darwin", vendor: "Zulu", version: "12.1.3", result: true},
		{os: "darwin", vendor: "Unknown", version: "Unknown", result: false},
		{os: "darwin", vendor: "J9", version: "1.5.0", result: false},
		{os: "darwin", vendor: "J9", version: "9", result: false},
		{os: "darwin", vendor: "IBM", version: "1.7.1", result: false},
	}

	for _, test := range matchFullySupportedTests {
		p.runtimeGOOS = test.os

		supported := p.isFullySupported(test.vendor, test.version)

		if supported != test.result {
			t.Errorf("Test failed. Supported status is %t. Result expected is %v. OS is %s. Vendor input is %s and version input is %s", supported, test.result, test.os, test.vendor, test.version)
		}
	}
}

func TestCheckLegacySupported(t *testing.T) {
	type testStruct struct {
		vendor  string
		version string
		result  bool
		os      string
	}
	p := JavaJVMVendorsVersions{}

	matchFullySupportedTests := []testStruct{

		// Compatibility tests taken from:
		// https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent

		// Updated: 4/8/2019

		// Apple Hotspot JVM version 6 for OS X
		{os: "darwin", vendor: "HotSpot", version: "6", result: true},
		{os: "darwin", vendor: "HotSpot", version: "7", result: false},

		{os: "darwin", vendor: "Apple", version: "6", result: true},
		{os: "darwin", vendor: "Apple", version: "7", result: false},

		{os: "windows", vendor: "Apple", version: "6", result: false},
		{os: "linux", vendor: "Apple", version: "6", result: false},

		// Oracle Hotspot JVM version 6.0 for Linux, Solaris, Windows, OS X
		{os: "linux", vendor: "HotSpot", version: "6", result: true},
		{os: "linux", vendor: "HotSpot", version: "7", result: false},

		{os: "windows", vendor: "HotSpot", version: "6", result: true},
		{os: "windows", vendor: "HotSpot", version: "7", result: false},

		{os: "darwin", vendor: "HotSpot", version: "6", result: true},
		{os: "darwin", vendor: "HotSpot", version: "7", result: false},

		// No "os: solaris" tests because we don't compile NR Diag for Solaris

		// IBM JVM version 6 for Linux
		{os: "linux", vendor: "IBM", version: "6", result: true},
		{os: "linux", vendor: "IBM", version: "7", result: false},

		{os: "darwin", vendor: "IBM", version: "6", result: false},
		{os: "darwin", vendor: "IBM", version: "7", result: false},

		{os: "windows", vendor: "IBM", version: "6", result: false},
		{os: "windows", vendor: "IBM", version: "7", result: false},

		{os: "windows", vendor: "HotSpot", version: "7", result: false},
	}

	for _, test := range matchFullySupportedTests {
		p.runtimeGOOS = test.os

		supported := p.isLegacySupported(test.vendor, test.version)

		if supported != test.result {
			t.Errorf("Test failed. Supported status is %t. Result expected is %v. OS is %s. Vendor input is %s and version input is %s", supported, test.result, test.os, test.vendor, test.version)
		}
	}
}
func TestParseVersion(t *testing.T) {
	matchVerTests := []struct {
		execOutput string
		version    string
	}{
		{execOutput: `openjdk version "1.8.0_152"
            OpenJDK Runtime Environment (Zulu 8.25.0.1-linux64) (build 1.8.0_152-b16)
            OpenJDK 64-Bit Server VM (Zulu 8.25.0.1-linux64) (build 25.152-b16, mixed mode)`, version: "1.8.0_152"},
		{execOutput: `java version "1.8.0_152"
            Java(TM) SE Runtime Environment (build 1.8.0_152-b16)
            Java HotSpot(TM) 64-Bit Server VM (build 25.152-b16, mixed mode)`, version: "1.8.0_152"},
		{execOutput: `java version "9"
            Java(TM) SE Runtime Environment (build 9+181)
            Java HotSpot(TM) 64-Bit Server VM (build 9+181, mixed mode)`, version: "9"},
		{execOutput: `java version "1.8.0_141"
            Java(TM) SE Runtime Environment (build 8.0.5.0 - pmz3180sr5-20170629_01(SR5))
            IBM J9 VM (build 2.9, JRE 1.8.0 z/OS s390-31 20170622_353577 (JIT enabled, AOT enabled)`, version: "1.8.0_141"},
		{execOutput: `java version "1.6.0_31"
            Java(TM) SE Runtime Environment (build 1.6.0_31-b05)
			Oracle JRockit(R) (build R28.2.3-13-149708-1.6.0_31-20120327-1523-linux-x86_64, compiled mode`, version: "1.6.0_31"},
		{execOutput: `javarooskies
			Java(TM) SE Runtime Environment (build 1.8.0_144-b01)
			Java(TM SE Runtime Environment)`, version: "1.8.0_144"},
		{execOutput: `java version "1.7.0"
			Java(TM SE Runtime Environment)`, version: ""},
		{execOutput: `Pick up dry cleaning
		Take mittens to the vet
		Buy earthquake survival kit`, version: ""},
	}

	for _, test := range matchVerTests {
		verCheck := extractVersionFromJavaExecutable(test.execOutput)
		if verCheck != test.version {
			t.Errorf("Test failed with exec output %s. Version extracted is %s. Version expected is %v", test.execOutput, verCheck, test.version)
		}
	}

}

func TestParseVendor(t *testing.T) {
	matchVendTests := []struct {
		execOutput string
		vendor     string
	}{
		{execOutput: `openjdk version "1.8.0_152"
            OpenJDK Runtime Environment (Zulu 8.25.0.1-linux64) (build 1.8.0_152-b16)
            OpenJDK 64-Bit Server VM (Zulu 8.25.0.1-linux64) (build 25.152-b16, mixed mode)`, vendor: "Zulu"},
		{execOutput: `java version "1.8.0_152"
            Java(TM) SE Runtime Environment (build 1.8.0_152-b16)
            Java HotSpot(TM) 64-Bit Server VM (build 25.152-b16, mixed mode)`, vendor: "HotSpot"},
		{execOutput: `java version "9"
            Java(TM) SE Runtime Environment (build 9+181)
            Java HotSpot(TM) 64-Bit Server VM (build 9+181, mixed mode)`, vendor: "HotSpot"},
		{execOutput: `java version "1.8.0_141"
            Java(TM) SE Runtime Environment (build 8.0.5.0 - pmz3180sr5-20170629_01(SR5))
            IBM J9 VM (build 2.9, JRE 1.8.0 z/OS s390-31 20170622_353577 (JIT enabled, AOT enabled)`, vendor: "IBM"},
		{execOutput: `java version "1.6.0_31"
            Java(TM) SE Runtime Environment (build 1.6.0_31-b05)
			Oracle JRockit(R) (build R28.2.3-13-149708-1.6.0_31-20120327-1523-linux-x86_64, compiled mode`, vendor: "JRockit"},
		{execOutput: `java version "1.8.0_201"
		Java(TM) SE Runtime Environment (build 8.0.5.31 - pxa6480sr5fp31-20190311_03(SR5 FP31))
		IBM J9 VM (build 2.9, JRE 1.8.0 Linux amd64-64-Bit Compressed References 20190306_411656 (JIT enabled, AOT enabled)
		OpenJ9   - d97ae0f
		OMR      - a975735
		IBM      - ad515e8)
		JCL - 20190307_01 based on Oracle jdk8u201-b09`, vendor: "IBM"},
		{execOutput: `foo`, vendor: ""},
	}

	for _, test := range matchVendTests {
		vendCheck := extractVendorFromJavaExecutable(test.execOutput)
		if vendCheck != test.vendor {
			t.Errorf("Test failed with exec output %s. Version extracted is %s. Version expected is %v", test.execOutput, vendCheck, test.vendor)
		}
	}
}

// From here on adding new testing with Ginkgo after bug fix.
var _ = Describe("JavaJVMVendorsVersions", func() {
	var p JavaJVMVendorsVersions

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Java",
				Subcategory: "JVM",
				Name:        "VendorsVersions",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Check Java process JVM compatibility with New Relic Java agent"

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
			expectedArray := []string{}

			Expect(p.Dependencies()).To(Equal(expectedArray))
		})
	})

	Describe("isFullySupported()", func() {
		var (
			isSupported bool
			vendor      string
			version     string
		)

		JustBeforeEach(func() {

			isSupported = p.isFullySupported(vendor, version)
		})

		Context("when the vendor is not part of requirements", func() {
			BeforeEach(func() {
				vendor = "FakeVendor"
				version = "1.3.3"
			})

			It("should return an expected Success result status", func() {
				Expect(isSupported).To(BeFalse())
			})

		})
		Context("when IBM vendor is running on Linux", func() {
			BeforeEach(func() {
				vendor = "IBM"
				version = "1.7"
				p.runtimeGOOS = "linux"
			})

			It("should return true", func() {
				Expect(isSupported).To(BeTrue())
			})

		})

	})
	Describe("isLegacySupported()", func() {
		var (
			isSupported bool
			vendor      string
			version     string
		)

		JustBeforeEach(func() {

			isSupported = p.isLegacySupported(vendor, version)
		})

		Context("when the vendor is not part of requirements", func() {
			BeforeEach(func() {
				vendor = "PeruvianChickenSauce"
				version = "1.6"
			})

			It("should return false", func() {
				Expect(isSupported).To(BeFalse())
			})

		})
		Context("when IBM vendor is running on Linux", func() {
			BeforeEach(func() {
				vendor = "IBM"
				version = "1.6"
				p.runtimeGOOS = "linux"
			})

			It("should return an expected Success result status", func() {
				Expect(isSupported).To(BeTrue())
			})

		})
		Context("when IBM vendor is not running on Linux", func() {
			BeforeEach(func() {
				vendor = "IBM"
				version = "1.6"
				p.runtimeGOOS = "darwin"
			})

			It("should return false", func() {
				Expect(isSupported).To(BeFalse())
			})

		})
		Context("when Apple hotspot vendor is running on OS X", func() {
			BeforeEach(func() {
				vendor = "Apple"
				version = "1.6"
				p.runtimeGOOS = "darwin"
			})

			It("should return an expected Success result status", func() {
				Expect(isSupported).To(BeTrue())
			})

		})
		Context("when Apple hotspot vendor is running on OS X", func() {
			BeforeEach(func() {
				vendor = "Apple"
				version = "1.6"
				p.runtimeGOOS = "Linux"
			})

			It("should return false", func() {
				Expect(isSupported).To(BeFalse())
			})

		})

	})

	Describe("parseJavaExecutable()", func() {
		var (
			//input
			cmdLineArgs string

			//output
			javaExecutable string
		)

		JustBeforeEach(func() {
			javaExecutable = parseJavaExecutable(cmdLineArgs)
		})

		Context("When given valid command line arguments with no classname passed to java executable", func() {
			BeforeEach(func() {
				cmdLineArgs = "/Applications/IntelliJ IDEA CE.app/Contents/jdk/Contents/Home/jre/bin/java -Djava.awt.headless=true -Didea.version==2018.3.4 -Xmx768m -Didea.maven.embedder.version=3.3.9 -Dfile.encoding=UTF-8 -classpath"
			})

			It("Should return parsed java executable", func() {
				Expect(javaExecutable).To(Equal("/Applications/IntelliJ IDEA CE.app/Contents/jdk/Contents/Home/jre/bin/java"))
			})
		})

		Context("When given valid command line arguments with no classname passed to java executable", func() {
			BeforeEach(func() {
				cmdLineArgs = `"C:\Program Files (x86)\Common Files\Oracle\Java\javapath\java.exe" -jar -javaagent:newrelic\newrelic.jar build\libs\lucessqs-1.0-SNAPSHOT.jar`
			})

			It("Should return parsed java executable", func() {
				Expect(javaExecutable).To(Equal(`C:\Program Files (x86)\Common Files\Oracle\Java\javapath\java.exe`))
			})
		})

		Context("When given valid command line arguments with classname passed to java executable", func() {
			BeforeEach(func() {
				cmdLineArgs = "/Applications/IntelliJIDEACE.app/Contents/jdk/Contents/Home/jre/bin/java CommandLineArgs -Djava.awt.headless=true -Didea.version==2018.3.4 -Xmx768m -Didea.maven.embedder.version=3.3.9 -Dfile.encoding=UTF-8 -classpath"
			})

			It("Should return parsed java executable", func() {
				Expect(javaExecutable).To(Equal("/Applications/IntelliJIDEACE.app/Contents/jdk/Contents/Home/jre/bin/java"))
			})
		})

		Context("When given invalid command line arguments with no java executable", func() {
			BeforeEach(func() {
				cmdLineArgs = "/my application/bin/python -true"
			})

			It("Should return an empty string", func() {
				Expect(javaExecutable).To(Equal(""))
			})
		})

		Context("When given valid command line arguments with classname passed to java executable and space in executable path", func() {
			BeforeEach(func() {
				cmdLineArgs = "/Applications/IntelliJ IDEA CE.app/Contents/jdk/Contents/Home/jre/bin/java MyClass -Djava.awt.headless=true -Didea.version==2018.3.4 -Xmx768m -Didea.maven.embedder.version=3.3.9 -Dfile.encoding=UTF-8 -classpath"
			})

			It("Should return an incorrect path", func() {
				Expect(javaExecutable).To(Equal("CE.app/Contents/jdk/Contents/Home/jre/bin/java"))
			})
		})
	})

	Describe("parseVendorDetailsByArgs()", func() {
		var (
			//input
			cmdLineArgs string

			//output
			vendor  string
			version string
			ok      bool
		)

		JustBeforeEach(func() {
			vendor, version, ok = parseVendorDetailsByArgs(cmdLineArgs)
		})

		Context("When given valid command line arguments of recognized vendor and parseable version", func() {
			BeforeEach(func() {
				cmdLineArgs = "-Djava.vm.name=Java HotSpot(TM) Client VM -Djava.util.logging.manager=com.ibm.ws.bootstrap.WsLogManager -Djava.runtime.version=pap6460sr16fp1ifx-20140908_01 (SR16 FP1) -Djavax.management.builder.initial=com.ibm.ws.management.PlatformMBeanServerBuilder"
				//HotSpot 1.6
			})
			It("Should return parsed vendor and version", func() {
				Expect(vendor).To(Equal("HotSpot"))
				Expect(version).To(Equal("1.6"))
				Expect(ok).To(Equal(true))
			})
		})

		Context("When given valid command line arguments of recognized vendor and unparseable version", func() {
			BeforeEach(func() {
				cmdLineArgs = "-Djava.vm.name=Java HotSpot(TM) Client VM -Djava.util.logging.manager=com.ibm.ws.bootstrap.WsLogManager -Djava.runtime.version=invalid_version_string -Djavax.management.builder.initial=com.ibm.ws.management.PlatformMBeanServerBuilder"
				//HotSpot 1.6
			})
			It("Should return empty vendor version and !ok", func() {
				Expect(vendor).To(Equal(""))
				Expect(version).To(Equal(""))
				Expect(ok).To(Equal(false))
			})
		})

		Context("When given valid command line arguments of unrecognized vendor", func() {
			BeforeEach(func() {
				cmdLineArgs = "-Djava.vm.name=Java Toolsy(TM) Client VM -Djava.util.logging.manager=com.ibm.ws.bootstrap.WsLogManager -Djava.runtime.version=pap6460sr16fp1ifx-20140908_01 (SR16 FP1) -Djavax.management.builder.initial=com.ibm.ws.management.PlatformMBeanServerBuilder"
				//HotSpot 1.6
			})
			It("Should return empty vendor version and !ok", func() {
				Expect(vendor).To(Equal(""))
				Expect(version).To(Equal(""))
				Expect(ok).To(Equal(false))
			})
		})
	})

	Describe("parseVendorDetailsByExe()", func() {
		var (
			//input
			javaExecutable string

			//output
			vendor  string
			version string
			ok      bool
		)

		JustBeforeEach(func() {
			vendor, version, ok = p.parseVendorDetailsByExe(javaExecutable)
		})

		Context("When given a valid executable that returns parseable version", func() {
			BeforeEach(func() {
				javaExecutable = "/path/to/java"
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					cmdOutput := `openjdk version "1.8.0_152"
					OpenJDK Runtime Environment (Zulu 8.25.0.1-linux64) (build 1.8.0_152-b16)
					OpenJDK 64-Bit Server VM (Zulu 8.25.0.1-linux64) (build 25.152-b16, mixed mode)`
					return []byte(cmdOutput), nil
				}
			})

			It("Should return the expected Vendor, Version, and Status", func() {
				Expect(vendor).To(Equal("Zulu"))
				Expect(version).To(Equal("1.8.0_152"))
				Expect(ok).To(Equal(true))
			})
		})
		Context("When there is an error executing java -version", func() {
			BeforeEach(func() {
				javaExecutable = ""
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return nil, errors.New("Unable to execute")
				}
			})

			It("Should return empty vendor version and !ok", func() {
				Expect(vendor).To(Equal(""))
				Expect(version).To(Equal(""))
				Expect(ok).To(Equal(false))
			})
		})

		Context("When given a valid executable that returns an unparseable version", func() {
			BeforeEach(func() {
				javaExecutable = "/path/to/java"
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					cmdOutput := `openste version "1.8.0_152"
					OpenSTE Runtime Environment (Toolsy 8.25.0.1-linux64) (build 1.8.0_152-b16)
					OpenSTE 64-Bit Server VM (Toolsy 8.25.0.1-linux64) (build 25.152-b16, mixed mode)`
					return []byte(cmdOutput), nil
				}
			})

			It("Should return an empty Vendor, Version, and false Status", func() {
				Expect(vendor).To(Equal(""))
				Expect(version).To(Equal(""))
				Expect(ok).To(Equal(false))
			})
		})

		Context("When given a valid executable that returns a recognized vendor with an unparseable version", func() {
			BeforeEach(func() {
				javaExecutable = "/path/to/java"
				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					cmdOutput := `openjdk version "ljkasdf"
					OpenJDK Runtime Environment (Zulu asdf-linux64) (build 1.8.0_152-b16)
					OpenJDK 64-Bit Server VM (Zulu aasdfefdf-linux64) (build 25.152-b16, mixed mode)`
					return []byte(cmdOutput), nil
				}
			})

			It("Should return an empty Vendor, Version, and false Status", func() {
				Expect(vendor).To(Equal(""))
				Expect(version).To(Equal(""))
				Expect(ok).To(Equal(false))
			})
		})

	})

	Describe("getSupportabilityCounts()", func() {
		var (
			//input
			javaPIDs []PIDInfo

			//output
			counts map[supportabilityStatus]int
		)

		JustBeforeEach(func() {
			counts = getSupportabilityCounts(javaPIDs)
		})

		Context("When provided an empty slice of JavaPIDs", func() {
			BeforeEach(func() {
				javaPIDs = []PIDInfo{}
			})

			It("Should return 0 for all supportability counts", func() {
				Expect(counts[NotSupported]).To(Equal(0))
				Expect(counts[LegacySupported]).To(Equal(0))
				Expect(counts[FullySupported]).To(Equal(0))
			})
		})

		Context("When provided a slice of JavaPIDs", func() {
			BeforeEach(func() {
				javaPIDs = []PIDInfo{
					{
						Supported: FullySupported,
					},
					{
						Supported: FullySupported,
					},
					{
						Supported: LegacySupported,
					},
					{
						Supported: LegacySupported,
					},
					{
						Supported: NotSupported,
					},
				}
			})

			It("Should return expected supportability counts", func() {
				Expect(counts[NotSupported]).To(Equal(1))
				Expect(counts[LegacySupported]).To(Equal(2))
				Expect(counts[FullySupported]).To(Equal(2))
			})
		})

	})

	Describe("determineSummaryStatus()", func() {
		var (
			//input
			counts map[supportabilityStatus]int

			//output
			status  tasks.Status
			summary string
		)

		JustBeforeEach(func() {
			status, summary = determineSummaryStatus(counts)
		})

		Context("When provided 0 total supportability counts", func() {
			BeforeEach(func() {

				counts = map[supportabilityStatus]int{}
			})

			It("Should return error status and expected summary", func() {
				Expect(status).To(Equal(tasks.Error))
				Expect(summary).To(Equal("Java processes were found, but an error occurred determining supportability."))
			})
		})

		Context("When provided counts where only fullysupported > 0", func() {
			BeforeEach(func() {

				counts = map[supportabilityStatus]int{
					FullySupported: 3,
				}
			})

			It("Should return success status and expected summary", func() {
				Expect(status).To(Equal(tasks.Success))
				lines := []string{
					"There are 3 Java process(es) running on this server. We detected:",
					"3 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
				}

				Expect(summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

		Context("When provided counts where any unsupported > 0", func() {
			BeforeEach(func() {

				counts = map[supportabilityStatus]int{
					NotSupported:   2,
					FullySupported: 2,
				}
			})

			It("Should return success status and expected summary", func() {
				lines := []string{
					"There are 4 Java process(es) running on this server. We detected:",
					"2 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
					"2 process(es) with an unsupported vendor/version combination.",
					"Please see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.",
				}

				Expect(status).To(Equal(tasks.Failure))
				Expect(summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

		Context("When provided counts where any legacySupported > 0", func() {
			BeforeEach(func() {

				counts = map[supportabilityStatus]int{
					FullySupported:  2,
					LegacySupported: 2,
				}
			})

			It("Should return warning status and expected summary", func() {
				lines := []string{
					"There are 4 Java process(es) running on this server. We detected:",
					"2 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
					"2 process(es) with a vendor/version requiring a legacy version of the Java Agent.",
					"Please see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.",
				}

				Expect(status).To(Equal(tasks.Warning))
				Expect(summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

	})

	Describe("determineSupportability()", func() {
		var (
			//input
			vendor  string
			version string

			//output
			supportability supportabilityStatus
		)

		JustBeforeEach(func() {
			supportability = p.determineSupportability(vendor, version)
		})

		Context("When given a fully supported vendor & version", func() {
			BeforeEach(func() {
				vendor = "OpenJDK"
				version = "1.8"
			})

			It("Should return FullySupported status", func() {
				Expect(supportability).To(Equal(FullySupported))
			})
		})

		Context("When given a fully supported vendor but unsupported version", func() {
			BeforeEach(func() {
				vendor = "OpenJDK"
				version = "1.6"
			})

			It("Should return FullySupported status", func() {
				Expect(supportability).To(Equal(NotSupported))
			})
		})

		Context("When given an OS specific legacy supported vendor and legacy supported version", func() {
			BeforeEach(func() {
				p.runtimeGOOS = "darwin"
				vendor = "Apple"
				version = "1.6"
			})

			It("Should return legacy supported status", func() {
				Expect(supportability).To(Equal(LegacySupported))
			})
		})

		Context("When given a legacy supported vendor and legacy supported version", func() {
			BeforeEach(func() {
				vendor = "HotSpot"
				version = "1.6"
			})

			It("Should return legacy supported status", func() {
				Expect(supportability).To(Equal(LegacySupported))
			})
		})

		Context("When given an unknown vendor and version", func() {
			BeforeEach(func() {
				vendor = "ToolsySDK"
				version = "1.6"
			})

			It("Should return not supported status", func() {
				Expect(supportability).To(Equal(NotSupported))
			})
		})
	})

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when there is an error getting Java processes", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{"Java/Agent/Version": tasks.Result{
					Status:  tasks.Success,
					Payload: "4.3",
				}}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, errors.New("an error message")
				}
			})

			It("should return an expected Error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("The task Java/JVM/VendorsVersions encountered an error while detecting all running Java processes."))
			})
		})
		Context("when there are no Java processes running", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{"Java/Agent/Version": tasks.Result{
					Status:  tasks.Success,
					Payload: "4.3",
				}}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, nil
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("This task did not detect any running Java processes on this system."))
			})
		})
		Context("when there are fully supported Java processes", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.runtimeGOOS = "linux" // IBM only compatible on Linux
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
						process.Process{
							Pid: 2,
						},
					}, nil
				}
				p.getCmdLineArgs = func(proc process.Process) (string, error) {
					var cmdLineArgs string
					if proc.Pid == 1 {
						cmdLineArgs = "/usr/local/bin/hotspot/java"
					} else if proc.Pid == 2 {
						cmdLineArgs = "/usr/local/bin/ibmJ9/java"
					}

					return cmdLineArgs, nil
				}

				p.cmdExec = func(cmd string, args ...string) ([]byte, error) {
					var cmdOutput []byte
					if cmd == "/usr/local/bin/hotspot/java" {
						cmdOutput = []byte(`java version "9"
							Java(TM) SE Runtime Environment (build 9+181)
							Java HotSpot(TM) 64-Bit Server VM (build 9+181, mixed mode)`)
					} else if cmd == "/usr/local/bin/ibmJ9/java" {
						cmdOutput = []byte(`java version "1.8.0_201"
							"Java(TM) SE Runtime Environment (build 8.0.5.31 - pxa6480sr5fp31-20190311_03(SR5 FP31))"
							"IBM VM (build 2.9, JRE 1.8.0 Linux amd64-64-Bit Compressed References 20190306_411656 (JIT enabled, AOT enabled)"
							"OpenJ9   - d97ae0f"
							"OMR      - a975735"
							"IBM      - ad515e8)"
							"JCL - 20190307_01 based on Oracle jdk8u201-b09"`)
					}

					return cmdOutput, nil
				}

			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				lines := []string{
					"There are 2 Java process(es) running on this server. We detected:",
					"2 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

		Context("When one of the attempts to retrieve the cmdLineArgs returns an error, but the other succeeds", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
						process.Process{
							Pid: 2,
						},
					}, nil
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte(`java version "9"
				Java(TM) SE Runtime Environment (build 9+181)
				Java HotSpot(TM) 64-Bit Server VM (build 9+181, mixed mode)`), nil
				}

				p.getCmdLineArgs = func(proc process.Process) (string, error) {
					// Generate an error for one process
					if proc.Pid == 1 {
						return "", errors.New("Couldn't do that! Error")
					} else {
						// All other processes are valid
						return "/usr/local/bin/javarooski/java", nil
					}
				}

			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Success result summary", func() {
				lines := []string{
					"There are 2 Java process(es) running on this server. We detected:",
					"1 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
					"1 process(es) with an unsupported vendor/version combination.",
					"Please see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

		Context("when unable to determine Java executable from cmdLineArgs", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
					}, nil
				}

				p.getCmdLineArgs = func(process.Process) (string, error) {
					// unparseable java executable, but valid cmdlineArgs for HotSpot and Java 1.8
					return `/home/duke/jeva -Xmx700m -Djava.vm.name=Java HotSpot™ 64-Bit Server VM -Djava.version=1.8 -Djava.awt.headless=true -Djava.endorsed.dirs="" -Djdt.compiler.useSingleThread=true -Dpreload.project.path=/Users/jmcgrath/code/samsa -Dpreload.config.path=/Users/jmcgrath/Library/Preferences/IdeaIC2018.3/options -Dexternal.project.config=/Users/jmcgrath/Library/Caches/IdeaIC2018.3/external_build_system/samsa.48129cba -Dcompile.parallel=false -Drebuild.on.dependency.change=true -Djava.net.preferIPv4Stack=true -Dio.netty.initialSeedUniquifier=-768099212347918098 -Dfile.encoding=UTF-8 -Duser.language=en -Duser.country=US -Didea.paths.selector=IdeaIC2018.3 -Didea.home.path=/Applications/IntelliJ IDEA CE.app/Contents -Didea.config.path=/Users/jmcgrath/Library/Preferences/IdeaIC2018.3 -Didea.plugins.path=/Users/jmcgrath/Library/Application Support/IdeaIC2018.3 -Djps.log.dir=/Users/jmcgrath/Library/Logs/IdeaIC2018.3/build-log -Djps.fallback.jdk.home=/Applications/IntelliJ IDEA CE.app/Contents/jdk/Contents/Home/jre -Djps.fallback.jdk.version=1.8.0_152-release -Dio.netty.noUnsafe=true -Djava.io.tmpdir=/Users/jmcgrath/Library/Caches/IdeaIC2018.3/compile-server/samsa_80e1e690/_temp_ -Djps.backward.ref.index.builder=true -Dkotlin.incremental.compilation=true -Dkotlin.daemon.enabled -Dkotlin.daemon.client.alive.path="/var/folders/8t/zvmxntvd4w7_j6flmkcj4jmr0000gn/T/kotlin-idea-3860326101882290058-is-running" -classpath /Applications/IntelliJ IDEA CE.app/Contents/lib/jps-launcher.jar:/Library/Java/JavaVirtualMachines/jdk1.8.0_152.jdk/Contents/Home/lib/tools.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/optimizedFileManager.jar org.jetbrains.jps.cmdline.Launcher /Applications/IntelliJ IDEA CE.app/Contents/lib/util.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jna-platform.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-aether-provider-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-builder-support-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-util-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/log4j.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/lz4-1.3.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-model-builder-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/asm-all-7.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-codec-1.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/nanoxml-2.2.3.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-repository-metadata-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-transport-file-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/trove4j.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-utils-3.0.22.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/httpcore-4.4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-codec-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-builders.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jna.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-buffer-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-component-annotations-1.6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-dependency-resolver.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-artifact-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-api-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-lang3-3.4.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-model-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-impl-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/httpclient-4.5.6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/idea_rt.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/resources_en.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-resolver-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-interpolation-1.21.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/oro-2.0.8.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/protobuf-java-3.4.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-model.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-transport-http-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-connector-basic-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/platform-api.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/forms-1.1-preview.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-api-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-builders-6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jdom.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-transport-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-spi-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/annotations.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-logging-1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/javac2.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-common-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-impldep-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/ant/lib/ant.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/groovy-all-2.4.15.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-api-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-log4j12-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/manifest-merger-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdk-common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-model-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-test-api-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/ddmlib-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/repository-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/manifest-merger-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdk-common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-model-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-test-api-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/ddmlib-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/repository-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/ant/lib/ant-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/uiDesigner/lib/jps/ui-designer-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/IntelliLang/lib/intellilang-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Groovy/lib/groovy-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Groovy/lib/groovy-rt-constants.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/eclipse/lib/eclipse-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/eclipse/lib/common-eclipse-util.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/maven/lib/maven-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/devkit/lib/devkit-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jps/android-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/android-common.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/build-common.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/android-rt.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdklib.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/layoutlib-api.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/jps/kotlin-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-stdlib.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-reflect.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/android-extensions-ide.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/android-extensions-compiler.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/javaFX/lib/javaFX-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/javaFX/lib/common-javaFX-plugin.jar org.jetbrains.jps.cmdline.BuildMain 127.0.0.1 58192 cb15d722-c706-4bbc-87e2-80e6e2700ba8 /Users/jmcgrath/Library/Caches/IdeaIC2018.3/compile-server
23429 ttys003    0:00.00 grep --color=auto --exclude-dir=.bzr --exclude-dir=CVS --exclude-dir=.git --exclude-dir=.hg --exclude-dir=.svn`, nil
				}

			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				lines := []string{
					"There are 1 Java process(es) running on this server. We detected:",
					"1 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})

		})
		Context("when unable to determine Java executable from cmdLineArgs and insufficient details present for fallback behavior", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
					}, nil
				}

				p.getCmdLineArgs = func(process.Process) (string, error) {
					// unparseable java executable, insufficient details in cmdLineArgs
					return `/home/duke/jeva HelloWorld -Djava.is.cool`, nil
				}

			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Failure result summary", func() {
				lines := []string{
					"There are 1 Java process(es) running on this server. We detected:",
					"1 process(es) with an unsupported vendor/version combination.",
					"Please see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})

		})
		Context("when cmdLine execution returns an error, but sufficient details present to determine Vendor and Version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
					}, nil
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte(""), errors.New("Duke wuz here")
				}

				p.getCmdLineArgs = func(process.Process) (string, error) {
					// parseable java executable (but we're going to get an error running it from the cmdExec mock)
					// fallback available with valid cmdlineArgs for HotSpot and Java 1.8
					return `/home/duke/java -Xmx700m -Djava.vm.name=Java HotSpot™ 64-Bit Server VM -Djava.version=1.8 -Djava.awt.headless=true -Djava.endorsed.dirs="" -Djdt.compiler.useSingleThread=true -Dpreload.project.path=/Users/jmcgrath/code/samsa -Dpreload.config.path=/Users/jmcgrath/Library/Preferences/IdeaIC2018.3/options -Dexternal.project.config=/Users/jmcgrath/Library/Caches/IdeaIC2018.3/external_build_system/samsa.48129cba -Dcompile.parallel=false -Drebuild.on.dependency.change=true -Djava.net.preferIPv4Stack=true -Dio.netty.initialSeedUniquifier=-768099212347918098 -Dfile.encoding=UTF-8 -Duser.language=en -Duser.country=US -Didea.paths.selector=IdeaIC2018.3 -Didea.home.path=/Applications/IntelliJ IDEA CE.app/Contents -Didea.config.path=/Users/jmcgrath/Library/Preferences/IdeaIC2018.3 -Didea.plugins.path=/Users/jmcgrath/Library/Application Support/IdeaIC2018.3 -Djps.log.dir=/Users/jmcgrath/Library/Logs/IdeaIC2018.3/build-log -Djps.fallback.jdk.home=/Applications/IntelliJ IDEA CE.app/Contents/jdk/Contents/Home/jre -Djps.fallback.jdk.version=1.8.0_152-release -Dio.netty.noUnsafe=true -Djava.io.tmpdir=/Users/jmcgrath/Library/Caches/IdeaIC2018.3/compile-server/samsa_80e1e690/_temp_ -Djps.backward.ref.index.builder=true -Dkotlin.incremental.compilation=true -Dkotlin.daemon.enabled -Dkotlin.daemon.client.alive.path="/var/folders/8t/zvmxntvd4w7_j6flmkcj4jmr0000gn/T/kotlin-idea-3860326101882290058-is-running" -classpath /Applications/IntelliJ IDEA CE.app/Contents/lib/jps-launcher.jar:/Library/Java/JavaVirtualMachines/jdk1.8.0_152.jdk/Contents/Home/lib/tools.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/optimizedFileManager.jar org.jetbrains.jps.cmdline.Launcher /Applications/IntelliJ IDEA CE.app/Contents/lib/util.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jna-platform.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-aether-provider-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-builder-support-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-util-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/log4j.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/lz4-1.3.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-model-builder-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/asm-all-7.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-codec-1.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/nanoxml-2.2.3.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-repository-metadata-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-transport-file-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/trove4j.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-utils-3.0.22.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/httpcore-4.4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-codec-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-builders.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jna.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-buffer-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-component-annotations-1.6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-dependency-resolver.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-artifact-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-api-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-lang3-3.4.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/maven-model-3.3.9.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-impl-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/httpclient-4.5.6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/idea_rt.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/resources_en.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-resolver-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/plexus-interpolation-1.21.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/oro-2.0.8.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/protobuf-java-3.4.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-model.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-transport-http-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-connector-basic-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/platform-api.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/forms-1.1-preview.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-api-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jps-builders-6.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/jdom.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-transport-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/aether-spi-1.1.0.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/annotations.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/commons-logging-1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/javac2.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/netty-common-4.1.30.Final.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-impldep-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/ant/lib/ant.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/groovy-all-2.4.15.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-api-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/slf4j-log4j12-1.7.25.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/manifest-merger-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdk-common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-model-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-test-api-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/ddmlib-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/repository-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/gson-2.8.5.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/lib/guava-25.1-jre.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/manifest-merger-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdk-common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-model-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/builder-test-api-3.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/ddmlib-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/repository-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-api-4.10.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/ant/lib/ant-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/uiDesigner/lib/jps/ui-designer-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/IntelliLang/lib/intellilang-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Groovy/lib/groovy-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Groovy/lib/groovy-rt-constants.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/eclipse/lib/eclipse-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/eclipse/lib/common-eclipse-util.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/maven/lib/maven-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/gradle/lib/gradle-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/devkit/lib/devkit-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jps/android-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/android-common.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/build-common.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/android-rt.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/sdklib.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/common-26.1.2.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/jarutils.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/android/lib/layoutlib-api.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/jps/kotlin-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-stdlib.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-reflect.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/kotlin-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/android-extensions-ide.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/Kotlin/lib/android-extensions-compiler.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/javaFX/lib/javaFX-jps-plugin.jar:/Applications/IntelliJ IDEA CE.app/Contents/plugins/javaFX/lib/common-javaFX-plugin.jar org.jetbrains.jps.cmdline.BuildMain 127.0.0.1 58192 cb15d722-c706-4bbc-87e2-80e6e2700ba8 /Users/jmcgrath/Library/Caches/IdeaIC2018.3/compile-server
23429 ttys003    0:00.00 grep --color=auto --exclude-dir=.bzr --exclude-dir=CVS --exclude-dir=.git --exclude-dir=.hg --exclude-dir=.svn`, nil
				}

			})

			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				lines := []string{
					"There are 1 Java process(es) running on this server. We detected:",
					"1 process(es) with a vendor/version fully supported by the latest version of the Java Agent.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})

		})
		Context("when cmdLine execution returns an error, and insufficient details present to determine Vendor and Version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{
						process.Process{
							Pid: 1,
						},
					}, nil
				}
				p.cmdExec = func(string, ...string) ([]byte, error) {
					return []byte(""), errors.New("Duke wuz here")
				}

				p.getCmdLineArgs = func(process.Process) (string, error) {
					// parseable java executable (but we're going to get an error running it from the cmdExec mock)
					// fallback available but missing -Djava.vm.name
					return `/home/duke/java -Xmx700m -Djava.version=1.8 -Djava.awt.headless=true -Djava.endorsed.dirs="" --exclude-dir=.svn`, nil
				}

			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Success result summary", func() {
				lines := []string{
					"There are 1 Java process(es) running on this server. We detected:",
					"1 process(es) with an unsupported vendor/version combination.",
					"Please see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.",
				}
				Expect(result.Summary).To(Equal(strings.Join(lines, "\n")))
			})

		})
	})

})
