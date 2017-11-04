package protocol

import (
	"encoding/binary"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

var (
	ErrVariableLengthDigit = errors.New("VarLen is too big")
)

type VarInt int32
type VarLong int64

func Read(r io.Reader, v interface{}) (n uint, err error) {
	switch v := v.(type) {
	case *string:
		stringLen, n, err := ReadVarInt(r)
		if err != nil {
			return 0, err
		}
		s := make([]byte, stringLen)
		r.Read(s)
		*v = string(s)
		return n, nil
	case *int64:
		p := make([]byte, 8)
		r.Read(p)
		*v = int64(binary.BigEndian.Uint64(p))
		return 8, nil
	case *VarInt:
		i, n, err := ReadVarInt(r)
		if err != nil {
			return 0, err
		}
		*v = VarInt(i)
		return n, nil
	case *uint16:
		p := make([]byte, 2)
		r.Read(p)
		*v = binary.BigEndian.Uint16(p)
		return 2, nil
	case *VarLong:
		i, n, err := ReadVarLong(r)
		if err != nil {
			return 0, err
		}
		*v = VarLong(i)
		return n, nil
	default:
		t := reflect.ValueOf(v)
		e := t.Elem()
		if e.Kind() == reflect.Struct {
			for i := 0; i < e.NumField(); i++ {
				field := e.Field(i)
				m, err := Read(r, field.Addr().Interface())
				if err != nil {
					return n, err
				}
				n += m
			}
			return n, nil
		} else {
			return 0, errors.Errorf("Data type %T not supported for reading", v)
		}
	}
}

func Write(w io.Writer, v interface{}) (n int, err error) {
	switch v := v.(type) {
	case VarInt:
		n, err = writeVarNumber(w, int64(v), 5)
	case VarLong:
		n, err = writeVarNumber(w, int64(v), 10)
	case int64:
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v))
		n, err = w.Write(b)
		if err != nil {
			return 0, errors.Wrapf(err, "Can't write in64 %d", v)
		}
	case string:
		vl, err := writeVarNumber(w, int64(len(v)), 5)
		if err != nil {
			return 0, errors.Wrapf(err, "Can't write string %s len", v)
		}
		b := []byte(v)
		_, err = w.Write(b)
		if err != nil {
			return 0, errors.Wrapf(err, "Can't write string %s bytes", v)
		}
		return vl + len(v), nil
	default:
		e := reflect.ValueOf(v)
		if e.Kind() == reflect.Struct {
			for i := 0; i < e.NumField(); i++ {
				field := e.Field(i)
				m, err := Write(w, field.Interface())
				if err != nil {
					return n, err
				}
				n += m
			}
			return n, nil
		} else {
			return 0, errors.Errorf("Data type %T not supported for writting", v)
		}
	}
	return
}

func readVarNumber(r io.Reader, maxBytes uint) (i int64, n uint, err error) {
	bs := make([]byte, 1)
	for {
		_, err := r.Read(bs)
		if err != nil {
			return 0, 0, err
		}
		b := bs[0]

		val := b & 0x7F
		i |= int64(val) << (7 * n)

		n++

		if n > maxBytes {
			err = ErrVariableLengthDigit
			return 0, 0, err
		}

		if (b & 0x80) == 0 {
			break
		}
	}
	return
}

func writeVarNumber(w io.Writer, number int64, maxBytes uint) (n int, err error) {
	for {
		temp := byte(number & 0x7F)
		number >>= 7
		if number != 0 {
			temp |= 0x80
		}
		b := []byte{temp}
		_, err := w.Write(b)
		if err != nil {
			return n, errors.Wrapf(err, "Error while writing varNumber %d, at %d bytes", number, maxBytes)
		}
		n++
		if number == 0 {
			break
		}
	}
	return
}

func ReadVarInt(r io.Reader) (i int32, n uint, err error) {
	res, n, err := readVarNumber(r, 5)
	i = int32(res)
	return
}

func ReadVarLong(r io.Reader) (i int64, n uint, err error) {
	return readVarNumber(r, 10)
}

type NullWritter struct{}

func (NullWritter) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func SizeOfSerializedData(v interface{}) int {
	nw := NullWritter{}
	n, _ := Write(nw, v)
	return n
}
