package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/roppenlabs/dobby-service/internal/config"
	types "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ValidatorsTestSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	validator Validator
}

type TopicNameRequest interface {
	TopicName(string)
	Partitions(int)
	ReplicationFactor(int)
	RetentionMs(int)
	Get() *types.CreateTopicRequest
}
type TopicNameRequestImpl struct {
	req *types.CreateTopicRequest
}

func NewTopicNameRequest() *TopicNameRequestImpl {
	return &TopicNameRequestImpl{
		req: &types.CreateTopicRequest{
			TopicName:         "Test123",
			Partitions:        300,
			ReplicationFactor: 300,
			RetentionMs:       300,
		},
	}
}

func (t *TopicNameRequestImpl) TopicName(s string) *TopicNameRequestImpl {
	t.req.TopicName = s
	return t
}
func (t *TopicNameRequestImpl) Partitions(x int) *TopicNameRequestImpl {
	t.req.Partitions = x
	return t
}
func (t *TopicNameRequestImpl) ReplicationFactor(x int) *TopicNameRequestImpl {
	t.req.ReplicationFactor = x
	return t
}
func (t *TopicNameRequestImpl) RetentionMs(x int) *TopicNameRequestImpl {
	t.req.RetentionMs = x
	return t
}

func (t *TopicNameRequestImpl) Get() *types.CreateTopicRequest {
	return t.req
}

func topicError(code string, s string) *types.KafkaError {
	return &types.KafkaError{
		Error: types.Error{
			Code:           code,
			Message:        s,
			DisplayMessage: s,
		},
	}
}

func TestKafkaCreateTopicValidatorsSuite(t *testing.T) {
	suite.Run(t, &ValidatorsTestSuite{})
}

func (v *ValidatorsTestSuite) BeforeTest(suiteName, testName string) {
	v.ctrl = gomock.NewController(v.T())
	v.validator = NewValidator(
		&config.Config{
			Kafka: config.KafkaConfig{
				Validations: config.KafkaValidations{
					MinReplicationFactor:       100,
					MaxReplicationFactor:       500,
					MinNumPartitions:           100,
					MaxNumPartitions:           500,
					MinRetentionInMS:           100,
					MaxRetentionInMS:           500,
					TopicNameConventionPattern: "^[a-zA-Z][a-zA-Z0-9]*$",
				},
			},
		},
	)
}

func (v *ValidatorsTestSuite) AfterTest(suiteName, testName string) {
	v.ctrl.Finish()
}

func (v *ValidatorsTestSuite) TestCorrectTopicNameShouldReturnTrue() {
	req := NewTopicNameRequest().TopicName("Test123").Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.True(v.T(), ok)
	assert.Nil(v.T(), err)
}

func (v *ValidatorsTestSuite) TestTopicNameShouldNotBeEmpty() {
	req := NewTopicNameRequest().TopicName("").Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", topicNameCannotBeEmtpy))
}

func (v *ValidatorsTestSuite) TestTopicNameShouldFollowNamingConvention() {
	req := NewTopicNameRequest().TopicName("&!BaDD&!!").Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", topicNameShouldFollowNamingConvention))
}

func (v *ValidatorsTestSuite) TestWrongRegexConfigForNamingConventionShouldReturnSomethingWentWrongWith500ResponseCode() {
	v.validator = NewValidator(
		&config.Config{
			Kafka: config.KafkaConfig{
				Validations: config.KafkaValidations{
					MinReplicationFactor:       100,
					MaxReplicationFactor:       500,
					MinNumPartitions:           100,
					MaxNumPartitions:           500,
					MinRetentionInMS:           100,
					MaxRetentionInMS:           500,
					TopicNameConventionPattern: "^[a-zA",
				},
			},
		},
	)
	req := NewTopicNameRequest().TopicName("TestName123").Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("INTERNAL_SERVER_ERROR", somethingWentWrong))
}

func (v *ValidatorsTestSuite) TestNumPartitionsShouldReturnOKForCorrectNumPartitions() {
	req := NewTopicNameRequest().Partitions(300).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.True(v.T(), ok)
	assert.Nil(v.T(), err)
}

func (v *ValidatorsTestSuite) TestNumPartitionsShouldNotBeGreaterThanAMaximumValue() {
	req := NewTopicNameRequest().Partitions(501).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", numPartitionsTooHigh))
}

func (v *ValidatorsTestSuite) TestNumPartitionsShouldNotBeLesserThanAMinimumValue() {
	req := NewTopicNameRequest().Partitions(99).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", numPartitionsTooLow))
}

func (v *ValidatorsTestSuite) TestReplicationFactorShouldReturnOKForCorrectNumPartitions() {
	req := NewTopicNameRequest().ReplicationFactor(300).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.True(v.T(), ok)
	assert.Nil(v.T(), err)
}

func (v *ValidatorsTestSuite) TestReplicationFactorShouldNotBeGreaterThanAMaximumValue() {
	req := NewTopicNameRequest().ReplicationFactor(501).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", replicationFactorTooHigh))
}

func (v *ValidatorsTestSuite) TestReplicationFactorShouldNotBeLesserThanAMinimumValue() {
	req := NewTopicNameRequest().ReplicationFactor(99).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", replicationFactorTooLow))
}

func (v *ValidatorsTestSuite) TestRetentionInMsShouldReturnOKForCorrectNumPartitions() {
	req := NewTopicNameRequest().RetentionMs(300).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.True(v.T(), ok)
	assert.Nil(v.T(), err)
}

func (v *ValidatorsTestSuite) TestRetentionInMsShouldNotBeGreaterThanAMaximumValue() {
	req := NewTopicNameRequest().RetentionMs(501).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", retentionPeriodTooHigh))
}

func (v *ValidatorsTestSuite) TestRetentionInMsShouldNotBeLesserThanAMinimumValue() {
	req := NewTopicNameRequest().RetentionMs(99).Get()
	ok, err := v.validator.ValidateCreateTopicRequest(req)
	assert.False(v.T(), ok)
	assert.Equal(v.T(), err, topicError("BAD_REQUEST", retentionPeriodTooLow))
}
