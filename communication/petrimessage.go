package communication

import (
	"bytes"
	"encoding/gob"

	"github.com/FLAGlab/DCoPN/petrinet"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/payload"
	"github.com/pkg/errors"
)

// CommandType enums for PetriNodes communication
type CommandType string

const (
	// TransitionCommand to query transitions
	TransitionCommand  CommandType = "transitions"
	// FireCommand to activate a fire event on the PetriNet
	FireCommand        CommandType = "fire"
	// PrintCommand to print the current state of the PetriNet
	PrintCommand       CommandType = "print"
	// VoteCommand to vote for a leader node
	VoteCommand        CommandType = "vote"
	// RequestVoteCommand to request votes for a leader node
	RequestVoteCommand CommandType = "requestvote"
)

type petriMessage struct {
	Command CommandType
	Address string
	Term int
	FromType NodeType
	VoteGranted string
	Transitions []*petrinet.Transition
}

func (petriMessage) Read(reader payload.Reader) (noise.Message, error) {
	byts, err := reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read msg")
	}
	var m petriMessage
	dec := gob.NewDecoder(bytes.NewReader(byts))
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode msg")
	}

	return m, nil
}

func (m petriMessage) Write() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		log.Info().Msgf("Got a fucking error: %v", err)
	}
	return payload.NewWriter(nil).WriteBytes(buf.Bytes()).Bytes()
}
