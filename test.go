package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
)

func copyFilesToContainer(containerID, sourcePath, targetPath string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	tarReader, err := client.CreateTarReader(context.Background(), types.CopyToContainerOptions{
		Content: data,
	})
	if err != nil {
		return err
	}
	defer tarReader.Close()

	err = cli.CopyToContainer(context.Background(), containerID, targetPath, tarReader, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}
