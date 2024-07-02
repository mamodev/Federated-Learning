package npz

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"strings"
)

type NpyDict struct {
	Shape []int `json:"shape"`
	Descr string `json:"descr"`
	FortranOrder bool `json:"fortran_order"`
}


func parseNpyDict (dict string) (*NpyDict, error) {

	dict = strings.ReplaceAll(dict, " ", "")
	dict = strings.ReplaceAll(dict, "'", "\"")
	dict = strings.ReplaceAll(dict, "None", "null")
	dict = strings.ReplaceAll(dict, "True", "true")
	dict = strings.ReplaceAll(dict, "False", "false")
	dict = strings.ReplaceAll(dict, "(", "[")
	dict = strings.ReplaceAll(dict, ")", "]")
	dict = strings.ReplaceAll(dict, ",}", "}")
	dict = strings.ReplaceAll(dict, ",]", "]")

	var dictObj NpyDict
	err := json.Unmarshal([]byte(dict), &dictObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}

	return &dictObj, nil
}

func (dict NpyDict) ToBytes(offset int) []byte {
	shapeStr := "("
	for i, s := range dict.Shape {
		shapeStr += fmt.Sprintf("%d", s)
		if i < len(dict.Shape) - 1 {
			shapeStr += ", "
		}

		if len(dict.Shape) == 1 {
			shapeStr += ","
		}
	}

	shapeStr += ")"

	boolean := "False"
	if dict.FortranOrder {
		boolean = "True"
	}

	rawDict := []byte(fmt.Sprintf("{'descr': '%s', 'fortran_order': %s, 'shape': %s, }", dict.Descr, boolean, shapeStr))

	// check if is aligned to 64 
	// The plus 1 is for the newLine
	padding := 64 - ((offset + len(rawDict) + 1) % 64)

	// fmt.Println("TOTAL LEN: ", len(rawDict), offset, 1, (offset + len(rawDict) + 1) % 64)

	endBuff := make([]byte, padding + 1)
	
	for i := 0; i < len(endBuff) -1; i++ {
		endBuff[i] = ' '
	} 

	endBuff[len(endBuff) - 1] = '\n'

	return append(rawDict, endBuff...)
}


type ZipReader struct {
	OnClose func() error
	*zip.Reader
}

func (z *ZipReader) Close() error  {
	if z.OnClose != nil {
		return z.OnClose()
	}

	return nil
}


type NpzDict struct {
	Arrays map[string]NpyArray
	// Path string
	GetZipReader func() (*ZipReader, error)
}


func (npzDict *NpzDict) GetArrayReader (key string) (*NpyArrayReader, error) {
	_, ok := npzDict.Arrays[key]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	preader, err := npzDict.GetZipReader()
	if err != nil {
		return nil, fmt.Errorf("error opening .npz file %v", err)
	}

	for _, file := range preader.File {
		if file.Name == key {
		
			reader, err := file.Open()
			if err != nil {
				preader.Close()
				return nil, fmt.Errorf("error opening file inside .npz file %v", err)
			}

			r, err :=  NewNpyArrayReader(reader)
			if err != nil {
				preader.Close()
				reader.Close()
				return nil, fmt.Errorf("error creating NpyArrayReader %v", err)
			}

			r.onClose = func() error {
				return preader.Close()
			}

			return r, nil
		}
	}

	return nil, fmt.Errorf("key not found insize .npz file")
}

func (npzDict *NpzDict) HasSameShape (target *NpzDict) bool {
	if len(npzDict.Arrays) != len(target.Arrays) {
		return false
	}

	for key, value := range npzDict.Arrays {
		targetValue, ok := target.Arrays[key]
		if !ok {
			return false
		}

		if len(value.Shape) != len(targetValue.Shape) {
			return false
		}

		for i := range value.Shape {
			if value.Shape[i] != targetValue.Shape[i] {
				return false
			}
		}
	}

	return true
}

func HaveSameShape(npzDicts ...NpzDict) bool {
	if len(npzDicts) == 0 {
		return true
	}

	for i := 1; i < len(npzDicts); i++ {
		if !npzDicts[0].HasSameShape(&npzDicts[i]) {
			return false
		}
	}

	return true
}