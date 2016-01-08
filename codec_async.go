package link

import (
	"errors"
	"io"
	"sync"
)

var ErrBlocking = errors.New("operation blocking")

func Async(chanSize int, base CodecType) CodecType {
	return &asyncCodecType{base, chanSize}
}

type asyncCodecType struct {
	base     CodecType
	chanSize int
}

func (codecType *asyncCodecType) NewEncoder(w io.Writer) Encoder {
	encoder := &asyncEncoder{
		base:     codecType.base.NewEncoder(w),
		writer:   w,
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
	writer   io.Writer
	sendChan chan interface{}
	stopChan chan struct{}
	stopWait sync.WaitGroup
	stopOnce sync.Once
}

func (encoder *asyncEncoder) stop() {
	encoder.stopOnce.Do(func() {
		close(encoder.stopChan)
		encoder.stopWait.Wait()
		if closer, ok := encoder.writer.(io.Closer); ok {
			closer.Close()
		}
	})
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
	select {
	case encoder.sendChan <- msg:
	default:
		encoder.stop()
		return ErrBlocking
	}
	return nil
}

func (encoder *asyncEncoder) Dispose() {
	encoder.stop()
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
}
