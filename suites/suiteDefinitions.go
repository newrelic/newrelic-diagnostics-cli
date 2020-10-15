package suites

var suiteDefinitions = []Suite{
	{
		Identifier:  "java",
		DisplayName: "Java Agent",
		Description: "Java Agent installation",
		Tasks: []string{
			"Base/*",
			"Java/*",
		},
	},
	{
		Identifier:  "infra",
		DisplayName: "Infrastructure Agent",
		Description: "Infrastructure Agent installation",
		Tasks: []string{
			"Base/*",
			"Infra/*",
		},
	},
	{
		Identifier:  "infra:debug",
		DisplayName: "Infrastructure Agent (Debug)",
		Description: "Infrastructure Agent installation with 3 minutes of debug log collection",
		Tasks: []string{
			"Base/*",
			"Infra/*",
			"Infra/Agent/Debug",
		},
	},
	{
		Identifier:  "dotnet",
		DisplayName: ".NET Agent",
		Description: ".NET Agent installation",
		Tasks: []string{
			"Base/*",
			"DotNet/*",
		},
	},
	{
		Identifier:  "dotnetcore",
		DisplayName: ".NET Core Agent",
		Description: ".NET Core Agent installation",
		Tasks: []string{
			"Base/*",
			"DotNetCore/*",
		},
	},
	{
		Identifier:  "android",
		DisplayName: "Mobile Android Agent",
		Description: "Mobile Android Agent installation",
		Tasks: []string{
			"Base/*",
			"Android/*",
		},
	},
	{
		Identifier:  "ios",
		DisplayName: "Mobile iOS Agent",
		Description: "Mobile iOS Agent installation",
		Tasks: []string{
			"Base/*",
			"iOS/*",
		},
	},
	{
		Identifier:  "node",
		DisplayName: "Node Agent",
		Description: "Node Agent installation",
		Tasks: []string{
			"Base/*",
			"Node/*",
		},
	},
	{
		Identifier:  "php",
		DisplayName: "PHP Agent",
		Description: "PHP Agent installation",
		Tasks: []string{
			"Base/*",
			"Php/*",
		},
	},
	{
		Identifier:  "python",
		DisplayName: "Python Agent",
		Description: "Python Agent installation",
		Tasks: []string{
			"Base/*",
			"Python/*",
		},
	},
	{
		Identifier:  "ruby",
		DisplayName: "Ruby Agent",
		Description: "Ruby Agent installation",
		Tasks: []string{
			"Base/*",
			"Ruby/*",
		},
	},
	{
		Identifier:  "minion",
		DisplayName: "Synthetics Containerized Private Minion",
		Description: "Gather information about Containerized Private Minions",
		Tasks: []string{
			"Base/Env/HostInfo",
			"Base/Env/SELinux",
			"Synthetics/*",
		},
	},
	// { //These require a special option for target URL to be provided when running: -o Browser/Agent/Detect.url=http://localhost:3000
	// 	Identifier:  "browser",
	// 	DisplayName: "Browser Agent",
	// 	Description: "Diagnose Browser Agent installation",
	// 	Tasks: []string{
	// 		"Browser/*",
	// 	},
	// },
	{
		Identifier:  "all",
		DisplayName: "All New Relic Products",
		Description: "We only recommend this option if you are unsure of the NR product you are troubleshooting.",
		Tasks: []string{
			"*",
		},
	},
}
