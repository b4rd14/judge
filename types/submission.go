package types

import "time"

type SubmissionMessage struct {
	SubmissionID   string
	ProblemID      string
	UserID         string
	TimeStamp      string
	Type           string
	TestCaseNumber int
	TimeLimit      time.Duration
	MemoryLimit    int64
}
