package compatibilityVars

var RubyVersionAgentSupportability = map[string][]string{
	//the keys are the ruby version and the values are the agent versions that support that specific version
	"2.7":   []string{"6.9.0.363+"},
	"2.6":   []string{"5.7.0.350+"},
	"2.5":   []string{"4.8.0.341+"},
	"2.4":   []string{"3.18.0.329+"},
	"2.3":   []string{"3.9.9.275+"},
	"2.2":   []string{"3.9.9.275+"},
	"2.1":   []string{"3.9.9.275+"},
	"2.0":   []string{"3.9.6.257+"},
	"1.9.3": []string{"3.9.6.257-3.18.1.330"},
	"1.9.2": []string{"3.9.6.257-3.18.1.330"},
	"1.8.7": []string{"3.9.6.257-3.18.1.330"},
}

var PythonVersionAgentSupportability = map[string][]string{
	//the keys are the python version and the values are the agent versions that support that specific version
	"3.8": []string{"5.2.3.131+"},
	"3.7": []string{"3.4.0.95+"},
	"3.6": []string{"2.80.0.60+"},
	"3.5": []string{"2.78.0.57+"},
	"3.4": []string{"2.42.0.35-4.20.0.120"},
	"3.3": []string{"2.42.0.35-3.4.0.95"},
	"2.7": []string{"2.42.0.35+"},
	"2.6": []string{"2.42.0.35-3.4.0.95"},
}

/*
	List of supported JRE distributions
	The keys to this map are used verbatim to generate
	a regular expression in `extractVendorFromJavaExecutable`
	They should exactly match how they appear in the output
	of `java -version.`

	Any vendors not found in this map will be flagged as
	unsupported. Known unsupported vendors can be called out
	explicitly by using an empty slice of compatibility
	requirements.


*/

var SupportedJavaVersions = map[string][]string{
	// supported vendors
	"OpenJDK":    []string{"1.7-1.9.*", "7-15.*"},
	"HotSpot":    []string{"1.7-1.9.*", "7-15.*"},
	"JRockit":    []string{"1-1.6.0.50"},
	"Coretto":    []string{"1.8-1.9.*", "8-11.*"},
	"Zulu":       []string{"1.8-1.9.*", "8-12.*"},
	"IBM":        []string{"1.7-1.8.*", "7-8.*"},
	"Oracle":     []string{"1.5.*", "5.0.*"},
	"Zing":       []string{"1.8-1.9.*", "8-11.*"},
	"OpenJ9":     []string{"1.8-1.9.*", "8-13.*"},
	"Dragonwell": []string{"1.8-1.9.*", "8-11.*"},
}

//Supported only with Java agent 4.3.x:
var SupportedForJavaAgent4 = map[string][]string{
	"Apple":   []string{"1.6.*", "6.*"},
	"IBM":     []string{"1.6.*", "6.*"},
	"HotSpot": []string{"1.6.*", "6.*"},
}

var NodeSupportedVersions = map[string][]string{
	"12": []string{"6.0.0+"},
	"10": []string{"4.6.0-7.*"},
}

//https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-framework#net-version
// .NET framework as keys and .NET agent as values
var DotnetFrameworkSupportedVersions = map[string][]string{
	"4.8": []string{"7.0.0+"},
	"4.7": []string{"7.0.0+"},
	"4.6": []string{"7.0.0+"}, //should be inclusive of version such as 4.6.1
	"4.5": []string{"7.0.0+"},
}

var DotnetFrameworkOldVersions = map[string][]string{
	//To instrument applications running on .NET Framework version 4.0 and lower, you must run a version of the New Relic .NET agent earlier than 7.0
	"4.0": []string{"5.1.*-6.*"}, //5.0 and lower are EOL versions
	"3.5": []string{"5.1.*-6.*"},
	//Doc says .NET Framework 3.0 and 2.0 are no longer supported as September 2020:https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-framework
}

//.NET Core 2.0 or higher is supported by the New Relic .NET agent version 6.19 or higher

var DotnetCoreSupportedVersions = map[string][]string{
	"6.0": []string{"9.2.0+"},
	"5.0": []string{"8.35.0+"},
	"3.1": []string{"8.21.34.0+"},
	"3.0": []string{"8.21.34.0+"},
	"2.2": []string{"8.19.353.0+"},
	"2.1": []string{"8.19.353.0+"},
	"2.0": []string{"8.19.353.0+"},
}

//https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-core#net-version
