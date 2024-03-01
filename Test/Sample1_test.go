package Test

import (
	replier "GO/Judge/Replier"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

var (
	judgeServerDo sync.Once
	serverDo      sync.Once
)

func SetupJudge() {
	judgeServerDo.Do(func() {
		go replier.Reply()
		time.Sleep(1000 * time.Millisecond)
	})
}
func SetupServer() {
	serverDo.Do(func() {
		go StartServer()
		time.Sleep(300 * time.Millisecond)
	})
}

func TestEnv(t *testing.T) {
	env := replier.NewEnv()
	assert.NotNil(t, env)
}

func TestMINIOConnection(t *testing.T) {
	_, err := replier.NewMinIoClient()
	assert.Nil(t, err)
}

func TestRabbitMQConnection(t *testing.T) {
	_, err := replier.NewRabbitMQConnection()
	assert.Nil(t, err)
}

func TestDockerClient(t *testing.T) {
	_, err := replier.NewDockerClint()
	assert.Nil(t, err)
}

func TestCheckRuntime(t *testing.T) {
	check := replier.CheckRunTimeError("TestRunTime.txt")
	assert.Equal(t, check, true)
}

func TestCompareOutputAccepted(t *testing.T) {
	result := replier.CompareOutputs("output1.txt", "output2.txt")
	assert.Equal(t, result, "Accepted")
}

func TestCompareOutputWrong(t *testing.T) {
	result := replier.CompareOutputs("output1.txt", "output3.txt")
	assert.Equal(t, result, "Wrong Answer")
}

func TestDownload(t *testing.T) {
	client, _ := replier.NewMinIoClient()
	err := replier.Download(context.Background(), client, "submissions", "test", "Submissions")
	assert.Nil(t, err)
	assert.FileExists(t, "Submissions/test/test.py")
	assert.DirExists(t, "Submissions/test")

}

func TestDeployRabbitMq(t *testing.T) {
	_, err := replier.DeployRabbitMq("submit")
	assert.Nil(t, err)
}

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

func TestAcceptedWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"Accepted","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err := replier.DeployRabbitMq("result")
	result := make(map[string]interface{})
	select {
	case msg := <-msgs:
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
	}
	expected := map[string]interface{}{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Accepted", "TestCase6": "Accepted", "TestCase7": "Accepted", "TestCase8": "Accepted", "TestCase9": "Accepted", "TestCase10": "Accepted"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}

func TestWrongWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"Wrong","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err := replier.DeployRabbitMq("result")
	result := make(map[string]interface{})
	select {
	case msg := <-msgs:
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))

	}
	expected := map[string]interface{}{"TestCase1": "Wrong Answer", "TestCase2": "Wrong Answer", "TestCase3": "Wrong Answer", "TestCase4": "Wrong Answer", "TestCase5": "Wrong Answer", "TestCase6": "Wrong Answer", "TestCase7": "Wrong Answer", "TestCase8": "Wrong Answer", "TestCase9": "Wrong Answer", "TestCase10": "Wrong Answer"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}

func TestTimeLimitWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"TimeLimit","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err := replier.DeployRabbitMq("result")
	result := make(map[string]interface{})
	select {
	case msg := <-msgs:
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))

	}
	expected := map[string]interface{}{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Time Limit Exceeded", "TestCase6": "Time Limit Exceeded", "TestCase7": "Time Limit Exceeded", "TestCase8": "Time Limit Exceeded", "TestCase9": "Time Limit Exceeded", "TestCase10": "Time Limit Exceeded"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}

func TestRunTimeWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"Runtime","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err := replier.DeployRabbitMq("result")
	result := make(map[string]interface{})
	select {
	case msg := <-msgs:
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))

	}
	expected := map[string]interface{}{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Runtime Error", "TestCase6": "Runtime Error", "TestCase7": "Runtime Error", "TestCase8": "Runtime Error", "TestCase9": "Runtime Error", "TestCase10": "Runtime Error"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}
