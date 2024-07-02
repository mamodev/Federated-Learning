package demo

import (
	"context"
	"go-lib/lib/fed"
	"time"
)

/*
This file contains the functions that are used to query the registry for clients and oracles.
Is for DEMO purposes only.
*/

func FindClientPoll (reg fed.Registry, ctx context.Context, clients int) <-chan[]string {

	result := make(chan []string)

	var TargetClients int = clients

	go func () {
		for {
			select {
			case <-ctx.Done():
				result<-[]string{}
			case <-time.After(1 * time.Second): 
				clients, err := reg.FindSubscribers("GROUP_TOKEN", fed.TrainTask, []fed.WorkerProtocol{"http"}, []fed.ParamFilter{
					{
						Name: "type",
						Type: fed.PText,
						Value: "worker",
						Operator: fed.OpEq,
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

func FindOracle (reg fed.Registry, ctx context.Context) <-chan string {
	result := make(chan string)

	go func () {
		for {
			select {
			case <-ctx.Done():
				result<-""
			case <-time.After(1 * time.Second): 
				clients, err := reg.FindSubscribers("GROUP_TOKEN", fed.TrainTask, []fed.WorkerProtocol{"http"}, []fed.ParamFilter{
					{
						Name: "type",
						Type: fed.PText,
						Value: "oracle",
						Operator: fed.OpEq,
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
