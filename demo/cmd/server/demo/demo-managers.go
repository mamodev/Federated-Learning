package demo

import (
	"fmt"
	"go-lib/lib/fed"
)

func start_managers(clients []string, oracle string, pool fed.TaskManagerPool, registry fed.Registry) {
	evalTaskManager := NewEvalManager(func () {	
		err := registry.PushTask("GROUP_TOKEN", clients, trainTask)
		if err != nil {
			fmt.Println("Error pushing task", err, " training stopped")
		}
	})

	taskManager := fed.NewFedTrainTaskManager(&fed.FedTrainTaskManagerConfig{
		ModelFolder: "./model",
		InitialModel: "init.npz",
		FinalModel: "final.npz",
		Clients: clients,
		OnNextRound: OnNextRoundFactory(registry),
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