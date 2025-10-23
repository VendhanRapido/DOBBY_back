package kafkatypes

type CreateTopicRequest struct {
	ClusterID         string `json:"clusterId" binding:"required"`
	TopicName         string `json:"topicName" binding:"required"`
	Partitions        int    `json:"numPartitions" binding:"required"`
	ReplicationFactor int    `json:"replicationFactor" binding:"required"`
	RetentionMs       int    `json:"retentionMs" binding:"required"`
}

type CreateTopicResponse struct {
	Message string `json:"message"`
}
