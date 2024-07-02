package demo

import (
	"context"
	"fmt"
	"go-lib/lib/fed"
	"os"
	"os/signal"
)

func StartDemo(nclients int) {
	initModelFile()

	var registry  fed.Registry = fed.NewMemRegistry(fed.NewMuxTaskQueueFactory(10))
	var server fed.RegistryServer = fed.NewHttpRegistryServer(registry, 8080)

	var pool fed.TaskManagerPool = fed.NewMuxTaskManagerPool()
	var coordinator_server fed.TaskCoordinator = fed.NewHttpTaskCoordinator(8081, pool)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	server.Start(ctx);
	coordinator_server.Start(ctx)

	oracle  := <-FindOracle(registry, ctx)
	clients := <-FindClientPoll(registry, ctx, nclients)

	if oracle != ""  && len(clients) == nclients {
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