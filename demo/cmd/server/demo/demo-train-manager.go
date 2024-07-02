package demo

import (
	"encoding/binary"
	"fmt"
	"go-lib/lib/fed"
	"go-lib/lib/npz"
	"net/url"
	"os"
)


func OnNextRoundFactory (reg fed.Registry) func (manager *fed.FedTrainTaskManager) {
	return func (manager *fed.FedTrainTaskManager) {

		clients := manager.GetClients()
		files, err := loadClientsNpzDict(clients)
		if err != nil {
			fmt.Println("Fatal error loading clients npz, training stopped", err)
			return
		}

		writer, err := getOutputWriter(manager.ModelFolder, manager.FinalModel)
		if err != nil {
			fmt.Println("Fatal error creating output writer, training stopped", err)
			return	
		}
		defer writer.Close()

		err = npz.Aggragate(files, writer, func (values []float64) float64 {
			return npz.Mean(values)
		})

		if err != nil {
			fmt.Println("Fatal error aggragating files, training stopped", err)
			return
		}
		
		removeClientFiles(clients, manager.ModelFolder)

		reg.PushTask("GROUP_TOKEN", clients, evalTask)
	}
}

func removeClientFiles (clients []string, modelFolder string) {
	for _, client := range clients {
		path, err := url.JoinPath(modelFolder, client + ".npz")
		if err == nil {
			os.Remove(path)	
		}
	}
}

func loadClientsNpzDict (clients []string) ([]npz.NpzDict, error) {
	files := make([]npz.NpzDict, 0, len(clients))

	for _, client := range clients {
		path, err := url.JoinPath("./model", client + ".npz")
		if err != nil {
			return nil, err
		}

		npz, err := npz.LoadNpz(path)
		if err != nil {
			return nil, err
		}

		files = append(files, *npz)
	}

	return files, nil
}

func getOutputWriter (folder, model string) (*npz.NpzDictWriter, error) {
	outputPath, err := url.JoinPath(folder, model)
	if err != nil {
		return nil, err
	}

	writer, err := npz.NewNpzDictWriter(outputPath, binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	return writer, nil
}