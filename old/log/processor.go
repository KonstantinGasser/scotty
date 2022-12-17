package log

import "github.com/KonstantinGasser/scotty/models/base"

// Processor implements the base.Streamer interface
type Processor struct {
	// processed is the channel on which
	// processed log events are pushed onto
	processed chan base.Event
}

func NewProcessor() *Processor {
	return &Processor{
		processed: make(chan base.Event),
	}
}

func (pro Processor) Stream() <-chan base.Event {
	return pro.processed
}

func (pro Processor) process(line Line) {

	pro.processed <- line
}

func (pro Processor) Write(p []byte) (int, error) {

	pro.process(Line{Raw: string(p)})

	return len(p), nil
}

type Line struct {
	Raw string
}

func (l Line) View() string { return l.Raw }
func (l Line) Index() uint8 { return 0 }
