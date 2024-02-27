package replier

import (
	"io"
	"log"
	"os"
	"strings"
)

func CheckRunTime(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("%s: %s", "Failed to open file", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	out := make([]byte, 4096)
	_, err = file.Read(out)
	if err != nil {
		log.Printf("%s: %s , %s", "Failed to read from file", err, filename)
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
		log.Printf("%s: %s", "Failed to read from file", err)
	}

	out2Bytes, err := io.ReadAll(out2)
	if err != nil {
		log.Printf("%s: %s", "Failed to read from file", err)
	}
	if strings.TrimSpace(string(out1Bytes)) == strings.TrimSpace(string(out2Bytes)) {
		return "Accepted"
	} else {
		return "Wrong Answer"
	}
}
