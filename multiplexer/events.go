package multiplexer

type BeamError error
type BeamMessage struct {
	Label string
	Data  []byte
}
type BeamNew string
