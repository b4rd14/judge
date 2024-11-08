package Test

import (
	replier "GO/Judge/Replier"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccepted(t *testing.T) {

	msg := amqp.Delivery{
		ContentType: "application/json",
		Body:        []byte(`{"ProblemID":"12","UserID":"test","TimeStamp":"Accepted","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`),
	}
	client, _ := replier.NewDockerClint()
	minioClient, _ := replier.NewMinIoClient()
	outputs, err := MockSendToJudge(msg, minioClient, client)
	fmt.Println(outputs)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)

	expected := map[string]string{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Accepted", "TestCase6": "Accepted", "TestCase7": "Accepted", "TestCase8": "Accepted", "TestCase9": "Accepted", "TestCase10": "Accepted"}
	assert.Equal(t, outputs, expected)
}

func TestWrongAnswer(t *testing.T) {

	msg := amqp.Delivery{
		ContentType: "application/json",
		Body:        []byte(`{"ProblemID":"12","UserID":"test","TimeStamp":"Wrong","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`),
	}
	client, _ := replier.NewDockerClint()
	minioClient, _ := replier.NewMinIoClient()
	outputs, err := MockSendToJudge(msg, minioClient, client)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)

	expected := map[string]string{"TestCase1": "Wrong Answer", "TestCase2": "Wrong Answer", "TestCase3": "Wrong Answer", "TestCase4": "Wrong Answer", "TestCase5": "Wrong Answer", "TestCase6": "Wrong Answer", "TestCase7": "Wrong Answer", "TestCase8": "Wrong Answer", "TestCase9": "Wrong Answer", "TestCase10": "Wrong Answer"}
	assert.Equal(t, expected, outputs)
}

func TestRuntimeError(t *testing.T) {
	msg := amqp.Delivery{
		ContentType: "application/json",
		Body:        []byte(`{"ProblemID":"12","UserID":"test","TimeStamp":"Runtime","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`),
	}
	client, _ := replier.NewDockerClint()
	minioClient, _ := replier.NewMinIoClient()
	outputs, err := MockSendToJudge(msg, minioClient, client)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	expected := map[string]string{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Runtime Error", "TestCase6": "Runtime Error", "TestCase7": "Runtime Error", "TestCase8": "Runtime Error", "TestCase9": "Runtime Error", "TestCase10": "Runtime Error"}
	assert.Equal(t, expected, outputs)
}

func TestTimeLimitError(t *testing.T) {
	msg := amqp.Delivery{
		ContentType: "application/json",
		Body:        []byte(`{"ProblemID":"12","UserID":"test","TimeStamp":"TimeLimit","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`),
	}
	client, _ := replier.NewDockerClint()
	minioClient, _ := replier.NewMinIoClient()
	outputs, err := MockSendToJudge(msg, minioClient, client)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	expected := map[string]string{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Time Limit Exceeded", "TestCase6": "Time Limit Exceeded", "TestCase7": "Time Limit Exceeded", "TestCase8": "Time Limit Exceeded", "TestCase9": "Time Limit Exceeded", "TestCase10": "Time Limit Exceeded"}
	assert.Equal(t, expected, outputs)

}

func TestMemoryLimitError(t *testing.T) {
	msg := amqp.Delivery{
		ContentType: "application/json",
		Body:        []byte(`{"ProblemID":"12","UserID":"test","TimeStamp":"Memory","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`),
	}
	client, _ := replier.NewDockerClint()
	minioClient, _ := replier.NewMinIoClient()
	outputs, err := MockSendToJudge(msg, minioClient, client)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	expected := map[string]string{"TestCase1": "Memory Limit Exceeded", "TestCase2": "Memory Limit Exceeded", "TestCase3": "Memory Limit Exceeded", "TestCase4": "Memory Limit Exceeded", "TestCase5": "Memory Limit Exceeded", "TestCase6": "Memory Limit Exceeded", "TestCase7": "Memory Limit Exceeded", "TestCase8": "Memory Limit Exceeded", "TestCase9": "Memory Limit Exceeded", "TestCase10": "Memory Limit Exceeded"}
	assert.Equal(t, expected, outputs)
}
