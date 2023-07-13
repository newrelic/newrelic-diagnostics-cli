package env

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Infra/Env/ValidateZookeeperPath", func() {

	var p InfraEnvValidateZookeeperPath

	Describe("Dependencies()", func() {
		It("Should return expected dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{
				"Infra/Config/IntegrationsMatch",
				"Base/Env/CollectEnvVars",
			}))
		})
	})

	Describe("getKafkaBrokersList()", func() {
		var (
			zookeeperConfig                        ZookeeperConfig
			zookeeperShellPath, getArg, brokersArg string
			hasBrokersList                         bool
			resultSummary                          string
		)

		JustBeforeEach(func() {
			hasBrokersList, resultSummary = p.getKafkaBrokersList(zookeeperConfig, zookeeperShellPath, getArg, brokersArg)
		})

		Context("When zookeeper_path has not been set and we can connect to kafka brokers with the default path", func() {

			BeforeEach(func() {

				zookeeperConfig = ZookeeperConfig{
					Port: "2181",
					Path: "",
				}
				zookeeperShellPath = "/opt/kafka_2.13-2.6.0/bin/zookeeper-shell.sh"
				getArg = "ls"
				brokersArg = "/brokers/ids"

				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return []byte("Connecting to localhost:2181\n\nWATCHER::\n\nWatchedEvent state:SyncConnected type:None path:null\n[]\n"), nil
				}
			})

			It("should return an success result summary", func() {
				Expect(resultSummary).To(Equal("We ran the command /opt/kafka_2.13-2.6.0/bin/zookeeper-shell.sh localhost:2181 ls /brokers/ids and succesfully connected to a broker or list of brokers:\nConnecting to localhost:2181\n\nWATCHER::\n\nWatchedEvent state:SyncConnected type:None path:null\n[]\n")) //NOTE: that this an empty list of brokers (which may imply that kafka is not running. Otherwise we would at least see: [0]. But all it matters here is that we are able at least to connect to Zookeeper)
				Expect(hasBrokersList).To(BeTrue())
			})
		})
		Context("When zookeeper_path has not been set and we cannot connect to kafka brokers", func() {

			BeforeEach(func() {

				zookeeperConfig = ZookeeperConfig{
					Port: "2181",
					Path: "",
				}
				zookeeperShellPath = "/opt/kafka_2.13-2.6.0/bin/zookeeper-shell.sh"
				getArg = "ls"
				brokersArg = "/brokers/ids"

				p.cmdExec = func(name string, arg ...string) ([]byte, error) {
					return []byte("Connecting to localhost:2181\n\nWATCHER::\n\nWatchedEvent state:SyncConnected type:None path:null\nNode does not exist: /brokers/ids\n"), errors.New("exit status 1")
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(resultSummary).To(Equal("We ran the command - /opt/kafka_2.13-2.6.0/bin/zookeeper-shell.sh localhost:2181 ls /brokers/ids - and were unable to locate a list of brokers:\nexit status 1\nConnecting to localhost:2181\n\nWATCHER::\n\nWatchedEvent state:SyncConnected type:None path:null\nNode does not exist: /brokers/ids\n\nThis might be due to the Zookeeper nodes not being network accessible to where the integration is in place, or Zookeeper is not running, or it could be that the Zookeeper namespace has your broker information kept under a different path other than the default. Keep in mind that an alternative Zookeeper path can be set in the kafka-config.yml: https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/kafka-monitoring-integration#arguments"))
				Expect(hasBrokersList).To(BeFalse())
			})

		})
	})

})

func Test_getZookeeperConfigFromYml(t *testing.T) {
	type args struct {
		kafkaConfigPair *infraConfig.IntegrationFilePair
	}
	tests := []struct {
		name      string
		args      args
		want      ZookeeperConfig
		wantErr   bool
		wantedErr error
	}{
		// TODO: Add test cases.
		{
			name: "Initial Zookeeper Test",
			args: args{
				kafkaConfigPair: &infraConfig.IntegrationFilePair{
					Configuration: config.ValidateElement{
						ParsedResult: tasks.ValidateBlob{
							Key:      "zookeeper_path",
							Path:     "",
							RawValue: "",
							Children: []tasks.ValidateBlob{
								{
									Key:      "zookeeper_path",
									Path:     "",
									RawValue: "",
								},
							},
						},
					},
				},
			},
			want:      ZookeeperConfig{},
			wantErr:   true,
			wantedErr: fmt.Errorf("multiple keys found for zookeeper_path"),
		},
		{
			name: "With no zookeeper_hosts",
			args: args{
				kafkaConfigPair: &infraConfig.IntegrationFilePair{
					Configuration: config.ValidateElement{
						ParsedResult: tasks.ValidateBlob{
							Key:      "zookeeper_path",
							Path:     "",
							RawValue: "",
							Children: []tasks.ValidateBlob{
								{
									Key:      "zookeeper_hostssss",
									Path:     "",
									RawValue: []map[interface{}]interface{}{{"host": "something10.company.com", "port": 5101}, {"host": "something11.company.com", "port": 5101}},
									Children: []tasks.ValidateBlob{},
								},
							},
						},
					},
				},
			},
			want:      ZookeeperConfig{},
			wantErr:   true,
			wantedErr: fmt.Errorf("multiple keys found for zookeeper_path"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getZookeeperConfigFromYml(tt.args.kafkaConfigPair)
			if (err != nil) != tt.wantErr {
				t.Errorf("getZookeeperConfigFromYml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getZookeeperConfigFromYml() = %v, want %v", got, tt.want)
			}
		})
	}
}
