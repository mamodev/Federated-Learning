package npz

import (
	"encoding/binary"
	"fmt"
	"io"
)

type NpyArray struct {
	Shape []int
	Dtype string
	ColumnMajor bool
	shapeSize int
	Endian binary.ByteOrder
}

func (npyArray NpyArray) GetShapeSize() int {
	return npyArray.shapeSize
}

type NpyArrayReader struct {
	Array NpyArray
	Version int
	rawReader io.ReadCloser
	readed int
	onClose func() error
}


func NewNpyArrayReader(reader io.ReadCloser) (*NpyArrayReader, error) {
	magicNumber := make([]byte, 6)
	_, err := reader.Read(magicNumber)
	if err != nil {
		return nil, err
	}

	if string(magicNumber) != "\x93NUMPY" {
		return nil, fmt.Errorf("invalid magic number")
	}

	version := make([]byte, 2)
	err = binary.Read(reader, binary.LittleEndian, &version)
	if err != nil {
		return nil, fmt.Errorf("error reading version from file")
	}

	if version[0] != 1 && version[1] != 2 && version[1] != 3 {
		return nil, fmt.Errorf("invalid version number")
	}

	var headerLength uint
	if version[0] == 1 {
		var hl uint16
		err = binary.Read(reader, binary.LittleEndian, &hl)
		if err != nil {
			return nil, fmt.Errorf("error reading header length field version 1")
		}
		headerLength = uint(hl)
	} else {
		var hl uint32 
		err = binary.Read(reader, binary.LittleEndian, &hl)
		if err != nil {
			return nil, fmt.Errorf("error reading header length field version 2 or 3")
		}
		headerLength = uint(hl)
	}

	headerBuff := make([]byte, headerLength)
	err = binary.Read(reader, binary.LittleEndian, headerBuff)
	if err != nil {
		return nil, fmt.Errorf("error reading header from file")
	}

	// check version if version 3 header is UTF-8 else ASCII (basically latin1)
	header := string(headerBuff)

	options, err := parseNpyDict(header)
	if err != nil {
		return nil, fmt.Errorf("error parsing header: %v", err)
	}

	if len(options.Descr) < 2 {
		return nil, fmt.Errorf("dtype description is too short")
	}

	if options.Descr[0] != '<' && options.Descr[0] != '>' {
		return nil, fmt.Errorf("invalid dtype description")
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if options.Descr[0] == '>' {
		endian = binary.BigEndian
	}

	if len(options.Shape) == 0 {
		return nil, fmt.Errorf("shape is empty")
	}	
	
	size := 1
	for _, s := range options.Shape {
		size *= s
	}

	return &NpyArrayReader{
		Version: int(version[0]),
		Array: NpyArray{
			Shape: options.Shape,
			Dtype: options.Descr[1:],
			ColumnMajor: options.FortranOrder,
			shapeSize: size,
			Endian: endian,
		},
		rawReader: reader,
		readed: 0,
	}, nil
}

func (npyArrayReader *NpyArrayReader) HasNext () bool {
	return npyArrayReader.readed < npyArrayReader.Array.GetShapeSize()
}	

func (npyArrayReader *NpyArrayReader) Read () (float64, error) {
	if !npyArrayReader.HasNext() {
		return 0, fmt.Errorf("end of file")
	}

	val, err := readDtype(npyArrayReader.rawReader, npyArrayReader.Array.Dtype, npyArrayReader.Array.Endian)
	if err != nil {
		return 0, err
	}

	npyArrayReader.readed++
	return parseDtypeToFloat64(val)
}

func (npyArrayReader *NpyArrayReader) Close() error {
	err :=  npyArrayReader.rawReader.Close()
	if err != nil {
		return err
	}

	if npyArrayReader.onClose != nil {
		err = npyArrayReader.onClose()
		if err != nil {
			return err
		}
	}

	return nil
}


type NpyArrayWriter struct { 
	dtype string
	endian binary.ByteOrder
	w io.Writer
}

func (npyAW NpyArrayWriter) Write(val float64) error {
	return writeDtype(npyAW.w, npyAW.dtype, val, npyAW.endian)
}