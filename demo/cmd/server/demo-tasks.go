package main

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
