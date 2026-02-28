package abi

import (
	"encoding/binary"
	"math"
)

type Encoder struct {
	buf []byte
}

func NewEncoder(capHint int) *Encoder {
	return &Encoder{buf: make([]byte, 0, capHint)}
}

func (e *Encoder) U8(v uint8) {
	e.buf = append(e.buf, v)
}

func (e *Encoder) Bool(v bool) {
	if v {
		e.U8(1)
		return
	}
	e.U8(0)
}

func (e *Encoder) I32(v int32) {
	e.U32(uint32(v))
}

func (e *Encoder) U32(v uint32) {
	start := len(e.buf)
	e.buf = append(e.buf, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(e.buf[start:], v)
}

func (e *Encoder) I64(v int64) {
	e.U64(uint64(v))
}

func (e *Encoder) U64(v uint64) {
	start := len(e.buf)
	e.buf = append(e.buf, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint64(e.buf[start:], v)
}

func (e *Encoder) F64(v float64) {
	e.U64(math.Float64bits(v))
}

func (e *Encoder) Bytes(v []byte) {
	e.U32(uint32(len(v)))
	e.buf = append(e.buf, v...)
}

func (e *Encoder) String(v string) {
	e.Bytes([]byte(v))
}

func (e *Encoder) Data() []byte {
	return e.buf
}

type Decoder struct {
	data []byte
	off  int
	err  bool
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

func (d *Decoder) Ok() bool {
	return !d.err
}

func (d *Decoder) U8() uint8 {
	if d.off+1 > len(d.data) {
		d.err = true
		return 0
	}
	v := d.data[d.off]
	d.off++
	return v
}

func (d *Decoder) Bool() bool {
	return d.U8() == 1
}

func (d *Decoder) I32() int32 {
	return int32(d.U32())
}

func (d *Decoder) U32() uint32 {
	if d.off+4 > len(d.data) {
		d.err = true
		return 0
	}
	v := binary.LittleEndian.Uint32(d.data[d.off:])
	d.off += 4
	return v
}

func (d *Decoder) I64() int64 {
	return int64(d.U64())
}

func (d *Decoder) U64() uint64 {
	if d.off+8 > len(d.data) {
		d.err = true
		return 0
	}
	v := binary.LittleEndian.Uint64(d.data[d.off:])
	d.off += 8
	return v
}

func (d *Decoder) F64() float64 {
	return math.Float64frombits(d.U64())
}

func (d *Decoder) Bytes() []byte {
	sz := int(d.U32())
	if d.off+sz > len(d.data) {
		d.err = true
		return nil
	}
	v := d.data[d.off : d.off+sz]
	d.off += sz
	return v
}

func (d *Decoder) String() string {
	return string(d.Bytes())
}
