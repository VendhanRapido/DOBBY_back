package kafkatypes

type ListTopicsRequest struct {
	ClusterID string `json:"clusterId" binding:"required"`
}

type ListTopicsResponse struct {
	Topics []TopicInfo `json:"result"`
}

type TopicInfo struct {
	Name              string `json:"name"`
	Partitions        int    `json:"partition,omitempty"`
	ReplicationFactor int    `json:"replicationFactor,omitempty"`
	RetentionMS       string `json:"retentionMs,omitempty"`
}
