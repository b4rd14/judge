package replier

import (
	"archive/tar"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"path/filepath"
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
		log.Fatalf("%s: %s", "Failed to open file", err)
	}
	defer func(out1 *os.File) {
		err := out1.Close()
		if err != nil {

		}
	}(out1)

	out2, err := os.Open(output2)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file", err)
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

func TarToTxt(reader io.ReadCloser, ID string) {
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
			file, err := os.Create(fmt.Sprintf("Submissions/%s/%s", ID, filepath.Base(header.Name)))
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

func calculatePrecision(truePositives, falsePositives int) float64 {
	if truePositives+falsePositives == 0 {
		return 0.0
	}

	return float64(truePositives) / float64(truePositives+falsePositives)
}

func calculateRecall(truePositives, falseNegatives int) float64 {
	if truePositives+falseNegatives == 0 {
		return 0.0
	}

	return float64(truePositives) / float64(truePositives+falseNegatives)
}

func calculateFScore(precision, recall float64) float64 {
	if precision+recall == 0 {
		return 0.0
	}

	return 2 * ((precision * recall) / (precision + recall))
}

func AnalyzeCSV() {

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
