package kafkatypes

type KafkaError struct {
	Error Error `json:"error"`
}

type Error struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	DisplayMessage string `json:"displayMessage"`
}
