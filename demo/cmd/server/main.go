package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go-lib/lib/npz"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func FindClientPoll (reg Registry, ctx context.Context, clients int) <-chan[]string {

	result := make(chan []string)

	var TargetClients int = clients

	go func () {
		for {
			select {
			case <-ctx.Done():
				result<-[]string{}
			case <-time.After(1 * time.Second): 
				clients, err := reg.FindSubscribers("GROUP_TOKEN", TrainTask, []WorkerProtocol{"http"}, []ParamFilter{
					{
						Name: "type",
						Type: PText,
						Value: "worker",
						Operator: OpEq,
					},
				}, TargetClients)

				if err != nil {
					// fmt.Println(err)
					continue
				}

				if len(clients) != TargetClients {
					continue
				}

				result<-clients
				close(result)
				return
			}
		}
	}()

	return result
}

func FindOracle (reg Registry, ctx context.Context) <-chan string {
	result := make(chan string)

	go func () {
		for {
			select {
			case <-ctx.Done():
				result<-""
			case <-time.After(1 * time.Second): 
				clients, err := reg.FindSubscribers("GROUP_TOKEN", TrainTask, []WorkerProtocol{"http"}, []ParamFilter{
					{
						Name: "type",
						Type: PText,
						Value: "oracle",
						Operator: OpEq,
					},
				}, 1)
				
				if err != nil {
					// fmt.Println("Looking for oracle: ", err)
					continue
				}

				result<-clients[0]
				close(result)
				return
			}
		}
	}()

	return result
}


var trainTask *TaskInstance = &TaskInstance{
	Token: "TRAIN_TOKEN",
	Type: TrainTask,
	Host: "localhost",
	Port: 8081,
	Protocols: []WorkerProtocol{PHttpWorker},
}

var evalTask *TaskInstance = &TaskInstance{
	Token: "EVAL_TOKEN",
	Type: EvaluateTask,
	Host: "localhost",
	Port: 8081,
	Protocols: []WorkerProtocol{PHttpWorker},
}

func main() {	

	// get number of clients from arguments
	if len(os.Args) != 2 {
		fmt.Println("Usage: server <number of clients>")
		return
	}

	// get number of clients
	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error parsing number of clients", err)
		return
	}

	os.Remove("./model/init.npz")
	
	original, err := os.Open("./model/original.npz")
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}


	file, err := os.Create("./model/init.npz")
	if err != nil {
		fmt.Println("Error creating file", err)
		return
	}
	_, err = io.Copy(file,original)
	if err != nil {
		fmt.Println("Error copying file", err)
		return
	}
	
	file.Close()
	original.Close()

	var registry Registry = NewMemRegistry(NewMuxTaskQueueFactory(10))
	var server RegistryServer =  NewHttpRegistryServer(registry, 8080)

	var pool TaskManagerPool = NewMuxTaskManagerPool()
	var coordinator_server TaskCoordinator = NewHttpTaskCoordinator(8081, pool)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	server.Start(ctx);
	coordinator_server.Start(ctx)


	fmt.Println("Server started")
	oracle := <-FindOracle(registry, ctx)
	
	if oracle != "" {
		fmt.Println("Found oracle", oracle)
	}

	fmt.Println("Waiting for clients")

	clients := <-FindClientPoll(registry, ctx, numClients)

	if len(clients) != 0 {
		fmt.Println("Clients founded, starting training")

		var round int = 0
		var prevAccuracy float64 = -1

		var evalTaskManager TaskManager = &SimpleTaskManager{
			GetPayloadFunc: func (client string) (Payload, error) {
				path := "./model/final.npz"
				if round == 0 {
					path = "./model/init.npz"
				}

				file, err := os.Open(path)
				if err != nil {
					return Payload{}, err
				}

				return Payload{
					MimeType: "application/octet-stream",
					Data: file,
				}, nil
			},

			RegisterResponseFunc: func (client string, response io.ReadCloser) error {
				defer response.Close()

				type EvalBody struct {
					Accuracy float64 `json:"accuracy"`
					Total int `json:"total"`
					Correct int `json:"correct"`
					Loss float64 `json:"loss"`
				}

				var body EvalBody
				err := json.NewDecoder(response).Decode(&body)
				if err != nil {
					return err
				}

				if round == 0 {
					prevAccuracy = body.Accuracy
				}

				accIncrementPrc := (body.Accuracy - prevAccuracy) / prevAccuracy * 100		
				prevAccuracy = body.Accuracy


				sign := "+"
				if accIncrementPrc < 0 {
					sign = ""
				}

				fmt.Printf("[R-%d] Acc: %.2f (%s%.2f), Loss: %.2f\n", round, body.Accuracy, sign, accIncrementPrc, body.Loss)

				if round != 0 {
					os.Remove("./model/init.npz")
					os.Rename("./model/final.npz", "./model/init.npz")
				}
			
				registry.PushTask("GROUP_TOKEN", clients, trainTask)
				round++;
				return nil
			},
		}

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

	coordinator_server.Wait()
	server.Wait()	
	fmt.Println("Server stopped")
}