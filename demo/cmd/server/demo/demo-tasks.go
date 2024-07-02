package demo

import (
	"go-lib/lib/fed"
)

var trainTask *fed.TaskInstance = &fed.TaskInstance{
	Token: "TRAIN_TOKEN",
	Type: fed.TrainTask,
	Host: "localhost",
	Port: 8081,
	Protocols: []fed.WorkerProtocol{fed.PHttpWorker},
}

var evalTask *fed.TaskInstance = &fed.TaskInstance{
	Token: "EVAL_TOKEN",
	Type: fed.EvaluateTask,
	Host: "localhost",
	Port: 8081,
	Protocols: []fed.WorkerProtocol{fed.PHttpWorker},
}
