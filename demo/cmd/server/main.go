package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"go-lib/lib/npz"
	"net/url"
	"os"
	"os/signal"
	"strconv"
)

func readArgs() int{
	if len(os.Args) != 2 {
		panic("Usage: server <number of clients>")
	}

	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("Error parsing number of clients")
	}

	return numClients
}

func start_managers(clients []string, oracle string, pool TaskManagerPool, registry Registry) {
		evalTaskManager := NewEvalManager(func () {	
			err := registry.PushTask("GROUP_TOKEN", clients, trainTask)
			if err != nil {
				fmt.Println("Error pushing task", err, " training stopped")
			}
		})

		taskManager := NewFedTrainTaskManager(&FedTrainTaskManagerConfig{
			ModelFolder: "./model",
			InitialModel: "init.npz",
			FinalModel: "final.npz",
			Clients: clients,
			onNextRound: func(manager *FedTrainTaskManager) {
				files := make([]*npz.NpzDict, 0, len(clients))

				sucess := false
				defer func () {
					if sucess {
						registry.PushTask("GROUP_TOKEN", []string{oracle}, evalTask)
					}
				}()

				defer func () {
					for _, client := range clients {
						manager.Responded[client] = false
						path, err := url.JoinPath(manager.ModelFolder, client + ".npz")
						if err == nil {
							os.Remove(path)
						}
					}
				}()

				for _, client := range clients {
					path, err := url.JoinPath(manager.ModelFolder, client + ".npz")
					if err != nil {
						fmt.Println("Error joining path", err)
						return
					}

					npz, err := npz.LoadNpz(path)
					if err != nil {
						fmt.Println("Error loading npz", err)
						return
					}

					files = append(files, npz)
				}

				outputPath, err := url.JoinPath(manager.ModelFolder, manager.FinalModel)
				if err != nil {
					fmt.Println("Error joining path", err)
					return
				}

				writer, err := npz.NewNpzDictWriter(outputPath, binary.LittleEndian)
				if err != nil {
					fmt.Println("Error creating writer", err)
					return
				}

				defer writer.Close()

				for _, npz := range files {
					if !npz.HasSameShape(files[0]) {
						fmt.Println("Shapes do not match")
						return
					}
				}

				shape := files[0].Arrays

				for key := range shape {
					w, err := writer.GetArrayWriter(key, shape[key])
					if err != nil {
						fmt.Println("Error getting array writer", err)
						return
					}

					readers := make([]*npz.NpyArrayReader, 0, len(files))
					defer func () {
						for _, reader := range readers {
							reader.Close()
						}
					}()


					for _, npz := range files {
						reader, err := npz.GetArrayReader(key)
						if err != nil {
							fmt.Println("Error getting array reader", err)
							return
						}

						readers = append(readers, reader)
					}

					for range(shape[key].GetShapeSize()) {
						var sum float64 = 0
						for _, reader := range readers {
							val, err := reader.Read()
							if err != nil {
								fmt.Println("Error reading value", err)
								return
							}

							sum += val
						}

						sum /= float64(len(readers))

						w.Write(sum)
					}
				}

				sucess = true
			},
		})

		pool.AddTaskManager("TRAIN_TOKEN", taskManager)
		pool.AddTaskManager("EVAL_TOKEN", evalTaskManager)

		fmt.Println("Created manager for train task and eval task")

		err := registry.PushTask("GROUP_TOKEN", []string{oracle}, evalTask)
		if err != nil {
			fmt.Println("Error pushing task", err)
		} 

		fmt.Println("Pushed task to clients")
}

func main() {	
	numClients:= readArgs()
	initModelFile()

	var registry Registry = NewMemRegistry(NewMuxTaskQueueFactory(10))
	var server RegistryServer =  NewHttpRegistryServer(registry, 8080)

	var pool TaskManagerPool = NewMuxTaskManagerPool()
	var coordinator_server TaskCoordinator = NewHttpTaskCoordinator(8081, pool)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	server.Start(ctx);
	coordinator_server.Start(ctx)

	oracle  := <-FindOracle(registry, ctx)
	clients := <-FindClientPoll(registry, ctx, numClients)

	if oracle != ""  && len(clients) == numClients {
		fmt.Println("Clients founded, starting training")

	}

	if len(clients) != 0 {
		fmt.Println("Clients founded, starting training")
		start_managers(clients, oracle, pool, registry)
	}

	coordinator_server.Wait()
	server.Wait()	
	fmt.Println("Server stopped")
}