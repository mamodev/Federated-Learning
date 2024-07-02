package demo

import (
	"io"
	"os"
)

func initModelFile() {
	original, err := os.Open("./model/original.npz")
	if err != nil {
		panic("Error opening file")
	}

	os.Remove("./model/init.npz")
	file, err := os.Create("./model/init.npz")
	if err != nil {
		panic("Error creating file")
	}
	_, err = io.Copy(file,original)
	if err != nil {
		panic("Error copying file")
	}
	
	file.Close()
	original.Close()

}