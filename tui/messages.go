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
