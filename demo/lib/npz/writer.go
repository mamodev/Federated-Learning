package npz

import (
	"archive/zip"
	"encoding/binary"
	"os"
)

type NpzDictWriter struct {
	Path string
	Endian binary.ByteOrder	
	writer *zip.Writer
}

func NewNpzDictWriter (path string, endian binary.ByteOrder) (*NpzDictWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	zipWriter := zip.NewWriter(file)

	return &NpzDictWriter{
		Path: path,
		Endian: endian,
		writer: zipWriter,
	}, nil
}

func (npzDictWriter *NpzDictWriter) GetArrayWriter (key string, array NpyArray) (*NpyArrayWriter, error) {
	w, err := npzDictWriter.writer.Create(key)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, []byte("\x93NUMPY"))
	if err != nil {
		return nil, err
	}

	version := []byte{1, 0}
	err = binary.Write(w, binary.LittleEndian, version)
	if err != nil {
		return nil, err
	}

	dtype := "<" + array.Dtype
	if npzDictWriter.Endian == binary.BigEndian {
		dtype = ">" + array.Dtype
	}	

	npzDict := NpyDict{
		Shape: array.Shape,
		Descr: dtype,
		FortranOrder: array.ColumnMajor,
	}

	header := npzDict.ToBytes(6 + 2 + 2);

	err = binary.Write(w, binary.LittleEndian, uint16(len(header)))
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, []byte(header))	
	if err != nil {
		return nil, err
	}

	return &NpyArrayWriter{
		dtype: array.Dtype,
		endian: array.Endian,
		w: w,
	}, nil
}

func (npzDictWriter *NpzDictWriter) Close() error {
	err := npzDictWriter.writer.Close()
	if err != nil {
		return err
	}

	return nil
}