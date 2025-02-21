package vote

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func TestNodeVotesImpl_Good(t *testing.T) {
	n := &NodeVotesImpl{good: 5}
	assert.Equal(t, 5, n.Good())
}

func TestNodeVotesImpl_Bad(t *testing.T) {
	n := &NodeVotesImpl{bad: 3}
	assert.Equal(t, 3, n.Bad())
}

func TestNodeVotesImpl_Upvote(t *testing.T) {
	n := &NodeVotesImpl{good: 0}
	n.Upvote()
	assert.Equal(t, 1, n.Good())
}

func TestNodeVotesImpl_Downvote(t *testing.T) {
	n := &NodeVotesImpl{bad: 0}
	n.Downvote()
	assert.Equal(t, 1, n.Bad())
}

func TestNewNodeVotes(t *testing.T) {
	n := NewNodeVotes()
	assert.Equal(t, 0, n.Good())
	assert.Equal(t, 0, n.Bad())
}

func TestNodeVotesImpl_EncodeDecodeMsgpack(t *testing.T) {
	n := &NodeVotesImpl{good: 10, bad: 5}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err := n.EncodeMsgpack(enc)
	assert.NoError(t, err)

	decoded := &NodeVotesImpl{}
	dec := msgpack.NewDecoder(&buf)
	err = decoded.DecodeMsgpack(dec)
	assert.NoError(t, err)

	assert.Equal(t, n.Good(), decoded.Good())
	assert.Equal(t, n.Bad(), decoded.Bad())
}

func TestNodeVotesImpl_EncodeMsgpack_Error(t *testing.T) {
	n := &NodeVotesImpl{good: 10, bad: 5}
	enc := msgpack.NewEncoder(&errWriter{})

	err := n.EncodeMsgpack(enc)
	assert.Error(t, err)
}

func TestNodeVotesImpl_DecodeMsgpack_Error(t *testing.T) {
	n := &NodeVotesImpl{}
	dec := msgpack.NewDecoder(&errReader{})

	err := n.DecodeMsgpack(dec)
	assert.Error(t, err)

	// Test for incorrect data type during decode
	buf := bytes.NewBuffer([]byte{0xC0}) // msgpack nil
	dec = msgpack.NewDecoder(buf)
	err = n.DecodeMsgpack(dec)
	assert.Error(t, err, "Expected error when decoding incorrect data type")
}

// errWriter is a writer that always returns an error.
type errWriter struct{}

func (e *errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("mock write error") // Explicit error message
}

// errReader is a reader that always returns an error.
type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error") // Explicit error message
}
