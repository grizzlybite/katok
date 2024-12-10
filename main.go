package main

import (
	"cmp"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/alecthomas/kong"
	"github.com/grizzlybite/katok/internal/version"
	capi "github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

type KafkaConfig struct {
	Brokers []string           `yaml:"kafka_brokers"`
	Topics  []KafkaTopicConfig `yaml:"topics"`
}

type KafkaTopicConfig struct {
	Name              string             `yaml:"name"`
	NumPartitions     *int32             `yaml:"num.partitions,omitempty"`
	ReplicationFactor *int16             `yaml:"replication.factor,omitempty"`
	ConfigEntries     map[string]*string `yaml:",inline,omitempty"`
}

var (
	DefaultNumPartitions     int32 = -1
	DefaultReplicationFactor int16 = -1
	Сonfig                   KafkaConfig
	provider                 ConfigProvider
)

var CLI struct {
	ConsulEnabled    string      `help:"Use consul: true || false." env:"CONSUL_ENABLED" default:"${consul_enabled}"`
	ConsulURL        string      `help:"Set consul url." env:"CONSUL_URL" default:"${consul_url}"`
	ConsulToken      string      `help:"Set consul acl token." env:"CONSUL_TOKEN" default:"${consul_token}"`
	ConsulConfigPath string      `help:"Set consul config path." env:"CONSUL_CONFIG_PATH" default:"${consul_config_path}"`
	ConfigPath       string      `help:"Set path to yaml config file." env:"CONFIG_PATH" default:"${config_path}"`
	Version          VersionFlag `name:"version" help:"Print version information and quit"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type ConfigProvider interface {
	GetConfig() (KafkaConfig, error)
}

type FileConfigProvider struct {
	configFile string
}

func (f *FileConfigProvider) GetConfig() (KafkaConfig, error) {
	var config KafkaConfig
	data, err := os.ReadFile(f.configFile)
	if err != nil {
		return config, fmt.Errorf("Error while reading yaml-config: %w", err)
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

type ConsulConfigProvider struct {
	consulAddress string
}

func (c *ConsulConfigProvider) GetConfig() (KafkaConfig, error) {
	var config KafkaConfig
	consulConfig := capi.DefaultConfig()
	consulConfig.Address = c.consulAddress
	if os.Getenv("CONSUL_TOKEN") != "" || CLI.ConsulToken != "" {
		consulConfig.Token = cmp.Or(os.Getenv("CONSUL_TOKEN"), CLI.ConsulToken)
	}
	consulClient, err := capi.NewClient(consulConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't connect to consul: %v", err))
		os.Exit(1)
	}

	consulKey := cmp.Or(os.Getenv("CONSUL_CONFIG_PATH"), CLI.ConsulConfigPath)
	kv, _, err := consulClient.KV().Get(consulKey, nil)
	if err != nil {
		log.Fatalf("Failed to get config from Consul: %v", err)
	}
	if kv == nil {
		log.Fatal("Config not found in Consul")
	}
	err = yaml.Unmarshal(kv.Value, &config)
	if err != nil {
		return config, fmt.Errorf("Failed to unmarshal config: %w", err)
	}
	return config, nil
}

func getTopicNumPartitions(topicName string, saramaConnect sarama.ClusterAdmin) (int, error) {
	topics := []string{topicName}
	numPartitions, err := saramaConnect.DescribeTopics(topics)
	if err != nil {
		return 0, err
	}
	currentNumPartitions := numPartitions
	numPartitions = nil

	for _, topicMetadata := range currentNumPartitions {
		if topicMetadata.Name == topicName {
			return len(topicMetadata.Partitions), nil
		}
	}
	return 0, nil
}

func main() {
	versionInfo := version.Get()
	slog.SetDefault(logger)
	_ = kong.Parse(&CLI,
		kong.Vars{
			"consul_enabled":     "false",
			"consul_url":         "http://127.0.0.1:8500",
			"consul_token":       "you-consul-acl-token",
			"consul_config_path": "kafka/topics",
			"config_path":        "./topics.yaml",
			"version":            versionInfo.GitTag,
		})

	if os.Getenv("CONSUL_ENABLED") == "true" || CLI.ConsulEnabled == "true" {
		provider = &ConsulConfigProvider{consulAddress: cmp.Or(os.Getenv("CONSUL_URL"), CLI.ConsulURL)}
		logger.Info("Using Consul config provider")
	} else {
		provider = &FileConfigProvider{configFile: cmp.Or(os.Getenv("CONFIG_PATH"), CLI.ConfigPath)}
		logger.Info("Using file config provider")
	}
	config, err := provider.GetConfig()
	if err != nil {
		logger.Error("Error getting config: %v\n", err)
		os.Exit(1)
	}

	brokers := config.Brokers
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_6_0_0
	admin, err := sarama.NewClusterAdmin(brokers, saramaConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't connect to kafka brokers: %v", err))
		os.Exit(1)
	}
	defer admin.Close()

	for _, topic := range config.Topics {
		if topic.NumPartitions == nil {
			topic.NumPartitions = &DefaultNumPartitions
		}
		if topic.ReplicationFactor == nil {
			topic.ReplicationFactor = &DefaultReplicationFactor
		}

		topicDetail := sarama.TopicDetail{
			NumPartitions:     *topic.NumPartitions,
			ReplicationFactor: *topic.ReplicationFactor,
			ConfigEntries:     topic.ConfigEntries,
		}

		currentNumPartitions, err := getTopicNumPartitions(topic.Name, admin)
		if err != nil {
			logger.Error("Topic creation failed", "topic", topic.Name, "error", err)
			os.Exit(1)
		}

		if topicDetail.NumPartitions >= int32(currentNumPartitions) || topicDetail.NumPartitions == DefaultNumPartitions {
			err = admin.CreateTopic(topic.Name, &topicDetail, false)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					err := admin.AlterConfig(sarama.TopicResource, topic.Name, topic.ConfigEntries, false)
					if err != nil {
						logger.Error(fmt.Sprintf("Update failed for '%s' topic. %v", topic.Name, err))
						os.Exit(1)
					} else {
						logger.Info(fmt.Sprintf("Successfully update parameters for '%s' topic.", topic.Name))
					}
				} else {
					logger.Error(fmt.Sprintf("Topic '%s' creation error. %v", topic.Name, err))
					os.Exit(1)
				}
			} else {
				logger.Info(fmt.Sprintf("Topic '%s' was successfully created.", topic.Name))
			}
		} else {
			logger.Error(fmt.Sprintf("New num.partition value for topic %s is less than current num.partitions value: %d < %d", topic.Name, topicDetail.NumPartitions, currentNumPartitions))
			os.Exit(1)
		}
	}
}
