package apiclient

type Request struct {
	TestId   string
	Function string
	Repeat   int
}

type Tape struct {
	Requests []Request
	cursor   int
}

func NewTape() Tape {
	return Tape{Requests: []Request{}, cursor: 0}
}

func (t *Tape) Add(testId string, function string, repeat int) {
	r := Request{TestId: testId, Function: function, Repeat: repeat}
	t.Requests = append(t.Requests, r)
}

func (t *Tape) NextRequest() Request {
	if t.cursor >= len(t.Requests) {
		return Request{}
	}

	r := t.Requests[t.cursor]
	t.cursor++

	return r
}

func (t *Tape) Reset() {
	t.cursor = 0
}
