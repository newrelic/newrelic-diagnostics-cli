package env

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
)

var _ = Describe("Infra/Env/NrjmxMbeans", func() {

	var p InfraEnvNrjmxMbeans

	Describe("Dependencies()", func() {
		It("Should return expected dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Infra/Config/ValidateJMX"}))
		})
	})

	Describe("Execute()", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When one of the mbeans is not found", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/ValidateJMX": tasks.Result{
						Status: tasks.Success,
						Payload: infraConfig.JmxConfig{
							Host:            "localhost",
							Port:            "8080",
							User:            "Admin",
							Password:        "Admin",
							CollectionFiles: "/etc/newrelic-infra/integrations.d/jvm-metrics.yml,/etc/newrelic-infra/integrations.d/tomcat-metrics.yml",
							JavaVersion:     "openjdk version \"11.0.9.1\" 2020-11-04\nOpenJDK Runtime Environment (build 11.0.9.1+1-Ubuntu-0ubuntu1.18.04)\nOpenJDK 64-Bit Server VM (build 11.0.9.1+1-Ubuntu-0ubuntu1.18.04, mixed mode, sharing)\n",
							JmxProcessCmdlineArgs: []string{
								"-Dcom.sun.management.jmxremote.password.file=_REDACTED_",
								"-Dcom.sun.management.jmxremote",
								"-Dcom.sun.management.jmxremote.authenticate=true",
								"-Dcom.sun.management.jmxremote.port=9010",
								"-Dcom.sun.management.jmxremote.ssl=false",
							},
						},
					},
				}
				p.getMBeanQueriesFromJMVMetricsYml = func(string) ([]string, error) {
					return []string{"java.lang:type=OperatingSystem", "java.lang:type=GarbageCollector"}, nil
				}
				/*Sample of unsuccesful output returned by cmdExecutor: []byte("{}\nNov 24, 2020 3:50:29 PM org.newrelic.nrjmx.JMXFetcher run\nFINE: Stopped receiving data, leaving...\n"), nil*/
				p.executeNrjmxCmdToFindBeans = func([]string, infraConfig.JmxConfig) ([]string, map[string]string) {
					return []string{"java.lang:type=OperatingSystem"}, make(map[string]string)
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Failure result summary", func() {
				Expect(result.Summary).To(Equal("In order to validate your queries defined in your metrics yml file against our JMX integration, we attempted to parsed them and ran each of them with the command echo {yourquery} | nrjmx -H localhost -P 8080 -v -d -\nThese queries returned an empty object({}): java.lang:type=OperatingSystem\nThis can mean that either those mBeans are not available to this JMX server or that the queries targetting them may need to be reformatted in the metrics yml file.\n"))
			})
		})

		Context("When mbean found returns metrics", func() {

			BeforeEach(func() {

				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/ValidateJMX": tasks.Result{
						Status: tasks.Success,
						Payload: infraConfig.JmxConfig{
							Host:            "localhost",
							Port:            "8080",
							User:            "Admin",
							Password:        "Admin",
							CollectionFiles: "/etc/newrelic-infra/integrations.d/jvm-metrics.yml",
							JavaVersion:     "openjdk version \"11.0.9.1\" 2020-11-04\nOpenJDK Runtime Environment (build 11.0.9.1+1-Ubuntu-0ubuntu1.18.04)\nOpenJDK 64-Bit Server VM (build 11.0.9.1+1-Ubuntu-0ubuntu1.18.04, mixed mode, sharing)\n",
							JmxProcessCmdlineArgs: []string{
								"-Dcom.sun.management.jmxremote.password.file=_REDACTED_",
								"-Dcom.sun.management.jmxremote",
								"-Dcom.sun.management.jmxremote.authenticate=true",
								"-Dcom.sun.management.jmxremote.port=9010",
								"-Dcom.sun.management.jmxremote.ssl=false",
							},
						},
					},
				}
				p.getMBeanQueriesFromJMVMetricsYml = func(string) ([]string, error) {
					return []string{"java.lang:type=OperatingSystem"}, nil
				}
				/*Sample of successful output returned by cmdExecutor: []byte("Nov 24, 2020 3:50:15 PM org.newrelic.nrjmx.JMXFetcher queryAttributes\nFINE: Unsuported data type (class javax.management.ObjectName) for bean java.lang:type=OperatingSystem,attr=ObjectName" + `{"java.lang:type\u003dOperatingSystem,attr\u003dSystemLoadAverage":0.55,"java.lang:type\u003dOperatingSystem,attr\u003dArch":"amd64","java.lang:type\u003dOperatingSystem,attr\u003dOpenFileDescriptorCount":36,"java.lang:type\u003dOperatingSystem,attr\u003dProcessCpuLoad":0.018476791347453808,"java.lang:type\u003dOperatingSystem,attr\u003dMaxFileDescriptorCount":1048576,"java.lang:type\u003dOperatingSystem,attr\u003dCommittedVirtualMemorySize":4074438656,"java.lang:type\u003dOperatingSystem,attr\u003dFreePhysicalMemorySize":3892604928,"java.lang:type\u003dOperatingSystem,attr\u003dTotalSwapSpaceSize":0,"java.lang:type\u003dOperatingSystem,attr\u003dName":"Linux","java.lang:type\u003dOperatingSystem,attr\u003dVersion":"4.15.0-72-generic","java.lang:type\u003dOperatingSystem,attr\u003dTotalPhysicalMemorySize":5193482240,"java.lang:type\u003dOperatingSystem,attr\u003dSystemCpuLoad":0.05002253267237495,"java.lang:type\u003dOperatingSystem,attr\u003dAvailableProcessors":2,"java.lang:type\u003dOperatingSystem,attr\u003dFreeSwapSpaceSize":0,"java.lang:type\u003dOperatingSystem,attr\u003dProcessCpuTime":24500000000}` + "\nNov 24, 2020 3:50:15 PM org.newrelic.nrjmx.JMXFetcher run\nFINE: Stopped receiving data, leaving...\n"), nil*/
				p.executeNrjmxCmdToFindBeans = func([]string, infraConfig.JmxConfig) ([]string, map[string]string) {
					return []string{}, make(map[string]string)
				}
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected success result summary", func() {
				Expect(result.Summary).To(Equal("In order to validate your queries defined in your metrics yml file against our JMX integration, we attempted to parsed them and ran each of them with the command echo {yourquery} | nrjmx -H localhost -P 8080 -v -d -\nAll queries returned successful metrics! The nrjmx integration is able to connect to the JMX server and query all the Mbeans that you had configured through your collection files."))
			})
		})
	})

})
