package replier

import (
	model "GO/Judge/Model"
	"archive/tar"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

		}
	}(archivePath)
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {

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

func CheckRunTime(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	out := make([]byte, 4096)
	_, err = file.Read(out)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}
	if strings.Contains(string(out), "Traceback (most recent call last):") {
		return true
	}
	return false
}

func CompareOutputs(output1 string, output2 string) string {
	out1, err := os.Open(output1)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file1", err)
	}
	defer func(out1 *os.File) {
		err := out1.Close()
		if err != nil {

		}
	}(out1)

	out2, err := os.Open(output2)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file2", err)
	}
	defer func(out2 *os.File) {
		err := out2.Close()
		if err != nil {

		}
	}(out2)

	out1Bytes, err := io.ReadAll(out1)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}

	out2Bytes, err := io.ReadAll(out2)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}
	if strings.TrimSpace(string(out1Bytes)) == strings.TrimSpace(string(out2Bytes)) {
		return "Accepted"
	} else {
		return "Wrong Answer"
	}
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
		log.Fatalf("%s: %s", "Failed to kill container", err)
	}
}

func RemoveDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to remove directory", err)
	}
}

func NewEnv() *model.ENV {
	env := model.ENV{}
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}
	err = viper.Unmarshal(&env)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}
	return &env
}

func DeployRabbitMq() (*amqp.Connection, <-chan amqp.Delivery, error) {
	conn, err := amqp.Dial("amqp://rabbitmq:DC6VaBq5WsR1pFG3gIAtJnA5euaNauyI@b9cf01c2-2518-496d-8cc0-f3c729bef2d7.hsvc.ir:31995")
	if err != nil {
		log.Printf("%s: %s", "Failed to connect to RabbitMQ", err)
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("%s: %s", "Failed to open a channel", err)
		return nil, nil, err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Printf("%s: %s", "Failed to close channel", err)
		}
	}(ch)
	if err != nil {
		log.Printf("%s: %s", "Failed to declare a queue", err)
		return nil, nil, err
	}
	msgs, err := ch.Consume("submit", "", false, false, false, false, nil)
	if err != nil {
		log.Printf("%s: %s", "Failed to register a consumer", err)
		return nil, nil, err
	}

	return conn, msgs, nil
}

func PythonJudge(msg amqp.Delivery, cli *client.Client, submission model.SubmissionMessage) {
	outputs, cli, resp, err := Run(cli, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
		msg.Ack(true)
		return
	}
	outputs = CheckTestCases(cli, resp.ID, outputs, submission)
	log.Println(outputs)
	RemoveDir("Submissions/" + submission.ProblemID + "/")
	msg.Ack(true)
}
