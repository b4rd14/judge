package Test

import (
	replier "GO/Judge/Replier"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	check := replier.CheckRunTime("TestRunTime.txt")
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
	_, err := replier.DeployRabbitMq()
	assert.Nil(t, err)
}
