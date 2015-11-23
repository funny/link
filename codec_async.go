package link

import (
	"errors"
	"io"
	"sync"
)

var ErrBlocking = errors.New("operation blocking")

func Async(base CodecType, chanSize int) CodecType {
	return &asyncCodecType{base, chanSize}
}

type AsyncMsg struct {
	Msg interface{}
}

type asyncCodecType struct {
	base     CodecType
	chanSize int
}

func (codecType *asyncCodecType) NewEncoder(w io.Writer) Encoder {
	encoder := &asyncEncoder{
		base:     codecType.base.NewEncoder(w),
		stopChan: make(chan struct{}),
		sendChan: make(chan interface{}, codecType.chanSize),
	}
	encoder.start()
	return encoder
}

func (codecType *asyncCodecType) NewDecoder(r io.Reader) Decoder {
	return codecType.base.NewDecoder(r)
}

type asyncEncoder struct {
	base     Encoder
	sendChan chan interface{}
	stopChan chan struct{}
	stopWait sync.WaitGroup
}

func (encoder *asyncEncoder) start() {
	var wait sync.WaitGroup
	wait.Add(1)
	encoder.stopWait.Add(1)
	go func() {
		wait.Done()
		defer encoder.stopWait.Done()
		for {
			select {
			case msg := <-encoder.sendChan:
				encoder.base.Encode(msg)
			case <-encoder.stopChan:
				return
			}
		}
	}()
	wait.Wait()
}

func (encoder *asyncEncoder) Encode(msg interface{}) error {
	if async, ok := msg.(AsyncMsg); ok {
		select {
		case encoder.sendChan <- async.Msg:
		default:
			return ErrBlocking
		}
		return nil
	}
	return encoder.Encode(msg)
}

func (encoder *asyncEncoder) Dispose() {
	close(encoder.stopChan)
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
	encoder.stopWait.Wait()
}
