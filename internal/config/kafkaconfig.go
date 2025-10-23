package config

type KafkaConfig struct {
	ClusterInfo map[string]string `yaml:"clusterInfo"`
	Validations KafkaValidations  `yaml:"validations"`
	TimeoutInMS KafkaTimeouts     `yaml:"timeoutInMS"`
}

type KafkaValidations struct {
	TopicNameConventionPattern string `yaml:"topicNameConventionPattern"`
	MinReplicationFactor       int    `yaml:"minReplicationFactor"`
	MaxReplicationFactor       int    `yaml:"maxReplicationFactor"`
	MinNumPartitions           int    `yaml:"minNumPartitions"`
	MaxNumPartitions           int    `yaml:"maxNumPartitions"`
	MinRetentionInMS           int    `yaml:"minRetentionInMS"`
	MaxRetentionInMS           int    `yaml:"maxRetentionInMS"`
}

type KafkaTimeouts struct {
	CreateTopic int `yaml:"createTopic"`
	DeleteTopic int `yaml:"deleteTopic"`
	GetMetadata int `yaml:"getMetadata"`
}

func (c *Config) GetKafkaClusterInfo() map[string]string {
	return c.Kafka.ClusterInfo
}

func (c *Config) GetKafkaValidationKeys() KafkaValidations {
	return c.Kafka.Validations
}

func (c *Config) GetCreateTopicTimeout() int {
	return c.Kafka.TimeoutInMS.CreateTopic
}

func (c *Config) GetDeleteTopicTimeout() int {
	return c.Kafka.TimeoutInMS.DeleteTopic
}

func (c *Config) GetMetadataTimeoutFromConfig() int {
	return c.Kafka.TimeoutInMS.GetMetadata
}
