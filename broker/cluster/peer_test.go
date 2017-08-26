package cluster

/*
import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	//"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/network/mock"
	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
)

func TestPeer_peerKey(t *testing.T) {
	node := "Peer1"
	expected := uint32(0x6e682264)

	assert.Equal(t, expected, peerKey(node))
}

func TestPeer_newPeer(t *testing.T) {
	conn := mock.NewConn()
	p := newPeer(conn.Client)
	defer p.Close()

	assert.NotNil(t, p)
	assert.Equal(t, conn.Client, p.socket)
	assert.Equal(t, snappy.NewBufferedWriter(conn.Client), p.writer)
	assert.Empty(t, p.frame)
	assert.NotNil(t, p.closing)
}

func TestPeerClose(t *testing.T) {
	conn := mock.NewConn()
	p := newPeer(conn.Client)

	var closed bool
	p.OnClosing = func(*Peer) { closed = true }
	p.Close()

	assert.NotNil(t, p.OnClosing)
	assert.True(t, closed)
}

func TestPeerSend(t *testing.T) {
	tests := []struct {
		ssid    []uint32
		channel string
		payload string
	}{
		{ssid: []uint32{1, 2}, channel: "a/b/", payload: "hi"},
		{ssid: []uint32{1, 2, 4}, channel: "a/b/c/", payload: "hello"},
	}

	for _, tc := range tests {
		peer := new(Peer)
		peer.frame = MessageFrame{}
		peer.Send(tc.ssid, []byte(tc.channel), []byte(tc.payload))

		assert.Contains(t, peer.frame, &Message{Ssid: tc.ssid, Channel: []byte(tc.channel), Payload: []byte(tc.payload)})
	}
}

func TestPeer_processSendQueue(t *testing.T) {
	tests := []struct {
		ssid       []uint32
		channel    string
		payload    string
		encodedLen int
	}{
		{ssid: []uint32{1, 2}, channel: "a/b/", payload: "hi", encodedLen: 58},
		{ssid: []uint32{1, 2, 4}, channel: "a/b/c/", payload: "hello", encodedLen: 64},
	}

	for _, tc := range tests {
		buffer := bytes.NewBuffer(nil)
		peer := &Peer{
			writer: snappy.NewBufferedWriter(buffer),
			frame: MessageFrame{
				&Message{Ssid: tc.ssid, Channel: []byte(tc.channel), Payload: []byte(tc.payload)},
			},
		}

		peer.processSendQueue()
		assert.Equal(t, tc.encodedLen, len(buffer.Bytes()))
	}
}

func TestPeer_processSendQueueFail(t *testing.T) {
	peer := &Peer{
		writer: snappy.NewWriter(nil),
		frame: MessageFrame{
			&Message{Ssid: []uint32{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("test")},
		},
	}

	assert.NotPanics(t, func() {
		peer.processSendQueue()
	})
}

func TestPeerHandshake(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		err      bool
	}{
		{name: "", expected: 0},
		{name: "test", expected: 46},
		{name: "longerpeername", expected: 57},
	}

	for _, tc := range tests {
		buffer := bytes.NewBuffer(nil)
		peer := &Peer{
			writer: snappy.NewBufferedWriter(buffer),
		}

		err := peer.Handshake(tc.name, nil)

		assert.Equal(t, tc.err, err != nil)
		assert.Equal(t, tc.expected, buffer.Len())
	}
}

func testPrintBytes(b []byte) {
	str := make([]string, 0)
	for _, b := range b {
		str = append(str, strconv.Itoa(int(b)))
	}

	fmt.Printf("%v\n", strings.Join(str, ","))
}

func TestPeerProcess(t *testing.T) {
	tests := []struct {
		msg1     []byte
		msg2     []byte
		accept   bool
		recv     bool
		err      bool
		expected *Message
	}{
		{msg1: []byte{}, msg2: []byte{}, err: true},
		{
			msg1:   []byte{255, 6, 0, 0, 115, 78, 97, 80, 112, 89, 1, 24, 0, 0, 144, 239, 34, 255, 118, 180, 1, 3, 75, 101, 121, 68, 180, 2, 4, 78, 111, 100, 101, 72, 116, 101, 115, 116},
			accept: false,
			err:    true,
		},
		{
			msg1: []byte{255, 6, 0, 0, 115, 78, 97, 80, 112, 89, 1, 24, 0, 0, 144, 239, 34, 255, 118, 180, 1, 3, 75, 101, 121, 68, 180, 2, 4, 78, 111, 100, 101, 72, 116, 101, 115, 116},
			msg2: []byte{1, 49, 0, 0, 45, 0, 110, 233, 101, 119, 180, 1, 7, 67, 104, 97, 110, 110, 101, 108, 90, 97, 47, 98, 47, 99, 47, 180, 2, 7, 80, 97, 121, 108, 111, 97, 100, 88,
				116, 101, 115, 116, 180, 3, 4, 83, 115, 105, 100, 103, 144, 145, 146},
			accept:   true,
			expected: &Message{Ssid: []uint32{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("test")},
			err:      true,
			recv:     true,
		},
	}

	for _, tc := range tests {
		count := 0
		output := bytes.NewBuffer(nil)
		input := bytes.NewBuffer(nil)
		peer := &Peer{
			closing: make(chan bool),
			socket:  mock.NewConn().Client,
			reader:  snappy.NewReader(input),
			writer:  snappy.NewBufferedWriter(output),
			OnHandshake: func(_ *Peer, _ HandshakeEvent) (err error) {
				if !tc.accept {
					err = errors.New("not accepted")
				}
				return
			},
			OnMessage: func(m *Message) {
				count++
				assert.Equal(t, tc.expected, m)
			},
		}

		input.Write(tc.msg1)
		input.Write(tc.msg2)

		err := peer.Process()
		assert.Equal(t, tc.err, err != nil, err.Error())
		if tc.recv {
			assert.Equal(t, 1, count)
		}
	}
}
*/
