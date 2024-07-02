package npz

import (
	"archive/zip"
	"bytes"
	"fmt"
)

func LoadNpz(path string) (*NpzDict, error) {
	return LoadNpzFromReader(func () (*ZipReader, error) {
		// return zip.OpenReader(path)

		reader, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}

		return &ZipReader{
			Reader: &reader.Reader,
			OnClose: func() error {
				return reader.Close()
			},
		}, nil

	})
}

func LoadNpzFromReader(getReader func() (*ZipReader, error)) (*NpzDict, error) {

	reader, err := getReader()
	if err != nil {
		return nil, err
	}

	npzDict := NpzDict{
		Arrays: make(map[string]NpyArray),
		GetZipReader: getReader,
	}

	for _, file := range reader.File {
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}

		r, err := NewNpyArrayReader(reader)
		if err != nil {
			reader.Close()
			return nil, fmt.Errorf("error creating NpyArrayReader, %v", err)
		}

		npzDict.Arrays[file.Name] =  r.Array
		r.Close()	
	}

	return &npzDict, nil
}

func LoadNpzFromBuffer(data []byte) (*NpzDict, error) {
	return LoadNpzFromReader(func () (*ZipReader, error) {

		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}

		return &ZipReader{
			Reader: reader,
		}, nil
	})
}

func LoadNzpFromBuffers(buffers [][]byte) ([]NpzDict, error) {
	npzDicts := make([]NpzDict, 0, len(buffers))
	for i, buffer := range buffers {
		npzDict, err := LoadNpzFromBuffer(buffer)
		if err != nil {
			return nil, fmt.Errorf("error loading %d th buffer, %v", i, err)
		}

		npzDicts = append(npzDicts, *npzDict)
	}

	return npzDicts, nil
}