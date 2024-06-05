package npz

import (
	"encoding/binary"
	"fmt"
	"io"
)

func writeDtype(writer io.Writer, dtype string, val float64, endian binary.ByteOrder) error {
	switch dtype {
	case "i1":
		v := int8(val)
		return binary.Write(writer, endian, v)
	case "i2":
		v := int16(val)
		return binary.Write(writer, endian, v)
	case "i4":
		v := int32(val)
		return binary.Write(writer, endian, v)
	case "i8":
		v := int64(val)
		return binary.Write(writer, endian, v)
	case "u1":
		v := uint8(val)
		return binary.Write(writer, endian, v)
	case "u2":
		v := uint16(val)
		return binary.Write(writer, endian, v)
	case "u4":
		v := uint32(val)
		return binary.Write(writer, endian, v)
	case "u8":
		v := uint64(val)
		return binary.Write(writer, endian, v)
	case "f4":
		v := float32(val)
		return binary.Write(writer, endian, v)
	case "f8":
		v := float64(val)
		return binary.Write(writer, endian, v)
	default:
		return fmt.Errorf("invalid dtype")
	}
}



func readDtype(reader io.Reader, dtype string, endian binary.ByteOrder) (interface{}, error) {
	var val any
	var err error

	switch dtype {
	case "i1":
		var v int8
		err = binary.Read(reader, endian, &v)
		val = v
	case "i2":
		var v int16
		err = binary.Read(reader, endian, &v)
		val = v
	case "i4":
		var v int32
		err = binary.Read(reader, endian, &v)
		val = v
	case "i8":
		var v int64
		err = binary.Read(reader, endian, &v)
		val = v
	case "u1":
		var v uint8
		err = binary.Read(reader, endian, &v)
		val = v
	case "u2":
		var v uint16
		err = binary.Read(reader, endian, &v)
		val = v
	case "u4":
		var v uint32
		err = binary.Read(reader, endian, &v)
		val = v
	case "u8":
		var v uint64
		err = binary.Read(reader, endian, &v)
		val = v
	case "f4":
		var v float32
		err = binary.Read(reader, endian, &v)
		val = v
	case "f8":
		var v float64
		err = binary.Read(reader, endian, &v)
		val = v
	default:
		err = fmt.Errorf("invalid dtype")
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}

func parseDtypeToFloat64 (val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("invalid type")
}
}