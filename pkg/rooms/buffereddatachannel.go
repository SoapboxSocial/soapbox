package rooms

import (
	"io"

	"github.com/pion/webrtc/v3"
)

type BufferedDataChannel struct {
	channel  *webrtc.DataChannel
	msgQueue chan []byte
}

func NewBufferedDataChannel() *BufferedDataChannel {
	return &BufferedDataChannel{
		msgQueue: make(chan []byte, 500),
	}
}

func (b *BufferedDataChannel) Start(channel *webrtc.DataChannel) {
	b.channel = channel

	b.channel.OnOpen(func() {
		go b.handle()
	})

	b.channel.OnClose(func() {
		close(b.msgQueue)
	})
}

func (b *BufferedDataChannel) Write(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = io.EOF
		}
	}()

	b.msgQueue <- data

	return nil
}

func (b *BufferedDataChannel) handle() {
	for msg := range b.msgQueue {
		err := b.channel.Send(msg)
		if err != nil && (err.Error() == "Stream closed" || err == io.EOF) {
			close(b.msgQueue)
		}
	}
}
