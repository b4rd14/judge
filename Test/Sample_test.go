package Test

import (
	"GO/Judge/Replier"
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


