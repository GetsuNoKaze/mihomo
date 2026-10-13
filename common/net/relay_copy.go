package net

import (
	"errors"
	"io"
	"syscall"

	"github.com/metacubex/sing/common"
	"github.com/metacubex/sing/common/buf"
	"github.com/metacubex/sing/common/bufio"
	"github.com/metacubex/sing/common/network"
)

// smallReadSize is the buffer, headroom included, that the relay reads into
// while a connection looks idle: a connection that waits for its peer holds
// this much instead of a full pool buffer (buf.BufferSize, 32 KiB).
const smallReadSize = 2048

// minSmallRead keeps the small buffer usable when the destination asks for a
// large headroom.
const minSmallRead = 512

// copyConn is bufio.Copy with another fallback for sources that cannot wait
// for data without a buffer.
//
// bufio.Copy ends in bufio.CopyExtendedWithPool for a source that is neither a
// socket (splice, or a read waiter that takes a buffer only once the socket
// is readable) nor a read waiter of its own, such as a TLS or REALITY
// connection. That loop takes a 32 KiB pool buffer before every read, so a
// connection whose client is silent keeps one for as long as it stays open:
// about two thirds of the live heap of an idle proxied connection.
//
// copyConn reads such a source into a small buffer while the connection looks
// idle and into full pool buffers while data keeps coming: a read that fills
// the small buffer switches to full buffers, and a full buffer read that
// returns no more than the small buffer would have held switches back. A
// bulk transfer runs on full buffers as before, at the cost of one more read
// where it starts; a connection that waits after a short message waits on
// the small buffer.
//
// Everything else (cached data, Vision's replaceable reader and writer,
// splice, read waiters) is bufio.Copy's own code.
func copyConn(destination io.Writer, source io.Reader) (n int64, err error) {
	if source == nil {
		return 0, errors.New("nil reader")
	} else if destination == nil {
		return 0, errors.New("nil writer")
	}
	originSource := source
	var readCounters, writeCounters []network.CountFunc
	var _n int64
	possibly := bufio.MaxCopyExtendedOnceTimes
	for {
		source, readCounters = network.UnwrapCountReader(source, readCounters)
		destination, writeCounters = network.UnwrapCountWriter(destination, writeCounters)
		if cachedSrc, isCached := source.(network.CachedReader); isCached {
			cachedBuffer := cachedSrc.ReadCached()
			if cachedBuffer != nil {
				dataLen := cachedBuffer.Len()
				_, err = destination.Write(cachedBuffer.Bytes())
				cachedBuffer.Release()
				if err != nil {
					return
				}
				countAll(readCounters, writeCounters, dataLen)
				continue
			}
		}
		replaceableReader, isReaderPossiblyReplaceable := source.(network.ReaderPossiblyReplaceable)
		replaceableWriter, isWriterPossiblyReplaceable := destination.(network.WriterPossiblyReplaceable)
		// Same condition (and operator precedence) as bufio.Copy.
		if possibly != 0 &&
			(isReaderPossiblyReplaceable && replaceableReader.ReaderPossiblyReplaceable()) ||
			(isWriterPossiblyReplaceable && replaceableWriter.WriterPossiblyReplaceable()) {
			possibly--
			_n, err = bufio.CopyExtendedOnce(destination, source, readCounters, writeCounters)
			n += _n
			if err != nil {
				if n == _n { // first time
					err = network.ReportHandshakeFailure(originSource, err)
				}
				if errors.Is(err, io.EOF) {
					err = nil
				}
				return
			}
			continue
		}
		break
	}
	_, srcIsSyscall := source.(syscall.Conn)
	_, dstIsSyscall := destination.(syscall.Conn)
	if srcIsSyscall && dstIsSyscall {
		// Splice, or a socket read waiter when it cannot: no buffer while idle.
		_n, err = bufio.CopyWithCounters(destination, source, originSource, readCounters, writeCounters)
		n += _n
		return
	}
	_n, err = copyExtended(originSource, bufio.NewExtendedWriter(destination), bufio.NewExtendedReader(source), readCounters, writeCounters)
	n += _n
	return
}

// copyExtended is bufio.CopyExtended with copyAdaptive in place of
// bufio.CopyExtendedWithPool.
func copyExtended(originSource io.Reader, destination network.ExtendedWriter, source network.ExtendedReader, readCounters []network.CountFunc, writeCounters []network.CountFunc) (n int64, err error) {
	options := network.NewReadWaitOptions(source, destination)
	if readWaiter, isReadWaiter := bufio.CreateReadWaiter(source); isReadWaiter {
		needCopy := readWaiter.InitializeReadWaiter(options)
		if !needCopy || common.LowMemory {
			return copyWait(originSource, destination, readWaiter, readCounters, writeCounters)
		}
	}
	return copyAdaptive(originSource, destination, source, options, readCounters, writeCounters)
}

// copyWait is the read waiter loop of bufio.CopyExtended.
func copyWait(originSource io.Reader, destination network.ExtendedWriter, source network.ReadWaiter, readCounters []network.CountFunc, writeCounters []network.CountFunc) (n int64, err error) {
	var notFirstTime bool
	for {
		var buffer *buf.Buffer
		buffer, err = source.WaitReadBuffer()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}
		dataLen := buffer.Len()
		err = destination.WriteBuffer(buffer)
		if err != nil {
			buffer.Leak()
			if !notFirstTime {
				err = network.ReportHandshakeFailure(originSource, err)
			}
			return
		}
		n += int64(dataLen)
		countAll(readCounters, writeCounters, dataLen)
		notFirstTime = true
	}
}

// copyAdaptive is bufio.CopyExtendedWithPool that waits on a small buffer
// while the source is idle (see copyConn).
func copyAdaptive(originSource io.Reader, destination network.ExtendedWriter, source network.ExtendedReader, options network.ReadWaitOptions, readCounters []network.CountFunc, writeCounters []network.CountFunc) (n int64, err error) {
	small := smallReadSize - options.FrontHeadroom - options.RearHeadroom
	if small < minSmallRead {
		small = minSmallRead
	}
	if options.MTU > 0 && options.MTU <= small {
		// The full buffer is no larger than the small one.
		return bufio.CopyExtendedWithPool(originSource, destination, source, readCounters, writeCounters)
	}
	var (
		notFirstTime bool
		full         bool
	)
	for {
		var buffer *buf.Buffer
		if full {
			buffer = options.NewBuffer()
		} else {
			buffer = newSmallBuffer(options, small)
		}
		err = source.ReadBuffer(buffer)
		if err != nil {
			buffer.Release()
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}
		dataLen := buffer.Len()
		options.PostReturn(buffer)
		err = destination.WriteBuffer(buffer)
		if err != nil {
			buffer.Leak()
			if !notFirstTime {
				err = network.ReportHandshakeFailure(originSource, err)
			}
			return
		}
		n += int64(dataLen)
		countAll(readCounters, writeCounters, dataLen)
		notFirstTime = true
		// A small read that filled its buffer: more is probably waiting.
		// A full read no larger than a small one: the burst is over.
		full = dataLen >= small
	}
}

// newSmallBuffer is ReadWaitOptions.NewBuffer with room for size bytes of
// data.
func newSmallBuffer(options network.ReadWaitOptions, size int) *buf.Buffer {
	buffer := buf.NewSize(options.FrontHeadroom + size + options.RearHeadroom)
	if options.FrontHeadroom > 0 {
		buffer.Resize(options.FrontHeadroom, 0)
	}
	if options.RearHeadroom > 0 {
		buffer.Reserve(options.RearHeadroom)
	}
	return buffer
}

func countAll(readCounters, writeCounters []network.CountFunc, n int) {
	for _, counter := range readCounters {
		counter(int64(n))
	}
	for _, counter := range writeCounters {
		counter(int64(n))
	}
}
