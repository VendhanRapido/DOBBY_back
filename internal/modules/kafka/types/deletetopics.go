package kafkatypes

type DeleteTopicResponse struct {
	Message string `json:"message"`
}

type DeleteTopicRequest struct {
	ClusterID string `json:"clusterId" binding:"required"`
	TopicName string `json:"topicName" binding:"required"`
}
