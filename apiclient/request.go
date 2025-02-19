package apiclient

type Frame struct {
	TestId   string
	Function string
	Repeat   int
}

func (f Frame) Requests() {
}

type Tape struct {
	Frames []Frame
	cursor int
}

func NewTape() Tape {
	return Tape{Frames: []Frame{}, cursor: 0}
}

func (t *Tape) Add(testId string, function string, repeat int) {
	r := Frame{TestId: testId, Function: function, Repeat: repeat}
	t.Frames = append(t.Frames, r)
}

func (t *Tape) Reset() {
	t.cursor = 0
}
