package tui

// AnnotationDoneMsg is emitted when all hunks have been annotated.
type AnnotationDoneMsg struct {
	Annotations []string
}

// QuizDoneMsg is emitted when the quiz is complete.
type QuizDoneMsg struct {
	Score  float64
	Passed bool
}

// CommitMsgReadyMsg is emitted when commit message generation completes.
type CommitMsgReadyMsg struct {
	Msg string
	Err error
}

// CommitConfirmedMsg is emitted when the user confirms the generated commit message.
type CommitConfirmedMsg struct {
	Message string
}
