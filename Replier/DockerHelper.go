package replier

import (
	model "GO/Judge/Model"
	"archive/tar"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func CopyDirToContainer(ctx context.Context, srcDir, destDir string, cli *client.Client, id string) error {
	archivePath := filepath.Join(os.TempDir(), "archive.tar")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			return
		}
	}(archivePath)
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {
			return
		}
	}(archiveFile)
	tw := tar.NewWriter(archiveFile)
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
				}
			}(file)
			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	err = tw.Close()
	if err != nil {
		return err
	}
	archiveFile, err = os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {

		}
	}(archiveFile)

	err = cli.CopyToContainer(ctx, id, destDir, archiveFile, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}

func TarToTxt(reader io.ReadCloser, submission model.SubmissionMessage) {
	read := tar.NewReader(reader)
	for {
		header, err := read.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		if header.Typeflag == tar.TypeReg {
			file, err := os.Create(fmt.Sprintf("Submissions/%s/%s", submission.ProblemID+"/"+submission.UserID+"/"+strconv.FormatInt(submission.TimeStamp, 10), filepath.Base(header.Name)))
			if err != nil {
				fmt.Println(err)
			}
			_, err = io.Copy(file, read)
			if err != nil {
				fmt.Println(err)
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func KillContainer(cli *client.Client, ctx context.Context, containerID string) {
	err := cli.ContainerKill(ctx, containerID, "SIGKILL")
	if err != nil {
		log.Printf("%s: %s", "Failed to kill container", err)
	}
}
