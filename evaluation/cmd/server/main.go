package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go-lib/lib/npz"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
)

func assert(cond bool, msg string) {
	if !cond {
		fail(msg)
	}
}

func assertNoErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)	
		// print error & stack trace
		fail(fmt.Sprintf("Error: %v\nFile: %s\nLine: %d\n", err, file, line))
	}
}

func fail(msg string) {
	os.Stderr.WriteString(msg)
	os.Exit(1)
}

var SHOULD_AGG_CURRENT_MODEL = false

var pending_models [][]byte = make([][]byte, 0)
var pending_models_lock = sync.Mutex{}

var curr_model_file []byte
var curr_model_file_lock = sync.RWMutex{}

func addPendingModel(model []byte) {
	pending_models_lock.Lock()
	defer pending_models_lock.Unlock()
	pending_models = append(pending_models, model)
	// if len(pending_models) >= MIN_FOR_AGGREGATION {
	// 	aggregation_chunks <- pending_models
	// 	pending_models = make([][]byte, 0)
	// }
}

func aggreateModels() {
	// wait for stdin "AGGREGATE" message

	for {
		var msg [9]byte
		_, err := os.Stdin.Read(msg[:])
		assertNoErr(err)
		assert(string(msg[:]) == "AGGREGATE", "Invalid message")

		pending_models_lock.Lock()
		models := make([][]byte, len(pending_models))
		copy(models, pending_models)
		pending_models = make([][]byte, 0)
		pending_models_lock.Unlock()
	
		assert(len(models) > 0, "No models to aggregate")
	
		var file_copy []byte
	
		if SHOULD_AGG_CURRENT_MODEL {
			curr_model_file_lock.RLock()
			file_copy = make([]byte, len(curr_model_file))	
			copy(file_copy, curr_model_file)
			curr_model_file_lock.RUnlock()
	
			models = append(models, file_copy)
		}
	
		// for i, model := range models {
		// 	fmt.Println("Model ", i, " size: ", len(model))
		// }
	
		dicts, err := npz.LoadNzpFromBuffers(models)
			
		assertNoErr(err)
	
		outBuffer := bytes.NewBuffer([]byte{})
		wc := &NopWriter{Writer: outBuffer}
		npzWriter, err := npz.NewNpzDictWriterFromWriter(wc, binary.LittleEndian)
		assertNoErr(err)
	
		err = npz.Aggragate(dicts, npzWriter, npz.Mean)
		assertNoErr(err)
	
		npzWriter.Close()
	
		checkDict, err := npz.LoadNpzFromBuffer(outBuffer.Bytes())
		assertNoErr(err)
		assert(npz.HaveSameShape(*checkDict, dicts[0]), "Aggregated model has different shape")
	
	
		// fmt.Println(len(outBuffer.Bytes()))
		// os.Stdout.Write([]byte{byte(len(curr_model_file))})
		// os.Stdout.Write(outBuffer.Bytes())
		// fmt.Println("Aggregated model size: ", len(outBuffer.Bytes()), " bytes")
	
		binary.Write(os.Stdout, binary.BigEndian, uint32(len(outBuffer.Bytes())))
		os.Stdout.Write(outBuffer.Bytes())
	
	
		curr_model_file_lock.Lock()
		curr_model_file = outBuffer.Bytes()
		curr_model_file_lock.Unlock()
	}


}

func main () {
	// get model path from arguments
	if len(os.Args) != 2 {
		fail("Usage: server <model_path>")
	}

	modelPath := os.Args[1]

	file, err := os.Open(modelPath)
	assertNoErr(err)	
	defer file.Close()

	// read all file content into curr_model_file
	curr_model_file, err = io.ReadAll(file)
	assertNoErr(err)

	ln, err := net.Listen("tcp", "0.0.0.0:8080")
	assertNoErr(err)

	go aggreateModels()

	// Accept incoming connections and handle them
	for {
			conn, err := ln.Accept()
			assertNoErr(err)

			// Handle the connection in a new goroutine
			go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// read first int32 for the length of the message
	var length uint32
	err := binary.Read(conn, binary.BigEndian, &length)
	assertNoErr(err)
	assert(length > 0, "Invalid message length")

	// read fist byte for request type
	var requestType byte
	err = binary.Read(conn, binary.BigEndian, &requestType)
	assertNoErr(err)

	// Requets type:
	// 1 - Get Model
	// 2 - Post Model
  assert(requestType == 1 || requestType == 2, fmt.Sprint("Invalid request type: ", requestType))
	switch requestType {
	case 1:
		assert(length == 1, "Invalid message length for request type 1")
		streamModel(conn)
	case 2:
		assert(length > 1, "Invalid message length for request type 2")
		writeModel(conn, length - 1)
	}
}

func streamModel(conn net.Conn) {
	curr_model_file_lock.RLock()
	defer curr_model_file_lock.RUnlock()

	err := binary.Write(conn, binary.BigEndian, uint32(len(curr_model_file)))
	assertNoErr(err)

	_, err = conn.Write(curr_model_file)
	assertNoErr(err)
}

func writeModel(conn net.Conn, length uint32) {
	// read the rest of the message
	model := make([]byte, length)
	_, err := io.ReadFull(conn, model)
	assertNoErr(err)

	// fmt.Println("Received model size: ", len(model), " bytes with expected size: ", length)

	addPendingModel(model)

	ack := []byte("OK")
	err = binary.Write(conn, binary.BigEndian, uint32(len(ack)))
	assertNoErr(err)

	_, err = conn.Write(ack)
	assertNoErr(err)
}

type NopWriter struct {
	Writer io.Writer
}

func (w *NopWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)
}

func (w *NopWriter) Close() error {
	return nil
}