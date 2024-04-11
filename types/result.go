package types

type ResultMessage struct {
	SubmissionID string
	ProblemID    string
	UserID       string
	Result       map[string]string
}
