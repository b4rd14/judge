package Test

import (
	replier "GO/Judge/Replier"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	serverDo sync.Once
)

func SetupJudge() {
	go replier.Reply()
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
	conn, err := replier.NewRabbitMQConnection()
	assert.Nil(t, err)
	ch, err := conn.NewChannel()
	assert.Nil(t, err)
	_, err = ch.ReadQueue("results")
	defer func(ch *replier.RabbitChannel) {
		err := ch.Close()
		if err != nil {
			return
		}
	}(ch)
	assert.Nil(t, err)
}
