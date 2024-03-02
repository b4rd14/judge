package Test

import (
	replier "GO/Judge/Replier"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAcceptedWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"Accepted","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err, conn, ch := replier.DeployRabbitMq("result")
	defer conn.Close()
	defer ch.Close()
	result := make(map[string]interface{})
	for msg := range msgs {
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
		break
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
	msgs, err, conn, ch := replier.DeployRabbitMq("result")
	defer conn.Close()
	defer ch.Close()
	result := make(map[string]interface{})
	for msg := range msgs {
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
		break
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
	msgs, err, conn, ch := replier.DeployRabbitMq("result")
	defer conn.Close()
	defer ch.Close()
	result := make(map[string]interface{})
	for msg := range msgs {
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
		break
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
	msgs, err, conn, ch := replier.DeployRabbitMq("result")
	defer conn.Close()
	defer ch.Close()
	result := make(map[string]interface{})
	for msg := range msgs {
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
		break
	}
	expected := map[string]interface{}{"TestCase1": "Accepted", "TestCase2": "Accepted", "TestCase3": "Accepted", "TestCase4": "Accepted", "TestCase5": "Runtime Error", "TestCase6": "Runtime Error", "TestCase7": "Runtime Error", "TestCase8": "Runtime Error", "TestCase9": "Runtime Error", "TestCase10": "Runtime Error"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}

func TestMemoryLimitWithRabbitMQ(t *testing.T) {
	SetupJudge()
	SetupServer()
	submission := []byte(`{"SubmissionID":"12","ProblemID":"12","UserID":"test","TimeStamp":"Memory","Type":"python","TestCaseNumber":10,"TimeLimit":1000000000,"MemoryLimit":256000000}`)
	_, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(submission))
	assert.Nil(t, err)
	msgs, err, conn, ch := replier.DeployRabbitMq("result")
	defer conn.Close()
	defer ch.Close()
	result := make(map[string]interface{})
	for msg := range msgs {
		msg.Ack(true)
		assert.Nil(t, json.Unmarshal(msg.Body, &result))
		break
	}
	expected := map[string]interface{}{"TestCase1": "Memory Limit Exceeded", "TestCase2": "Memory Limit Exceeded", "TestCase3": "Memory Limit Exceeded", "TestCase4": "Memory Limit Exceeded", "TestCase5": "Memory Limit Exceeded", "TestCase6": "Memory Limit Exceeded", "TestCase7": "Memory Limit Exceeded", "TestCase8": "Memory Limit Exceeded", "TestCase9": "Memory Limit Exceeded", "TestCase10": "Memory Limit Exceeded"}
	assert.Equal(t, "test", result["user_id"])
	assert.Equal(t, "12", result["problem_id"])
	assert.Equal(t, expected, result["results"])
}
