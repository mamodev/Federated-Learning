package fed

import (
	"context"
	"errors"
)

type Task string;
var (
	TrainTask Task = "train"
	PredictTask Task = "predict"
	EvaluateTask Task = "evaluate"
)

type WorkerProtocol string;
var (
	PHttpWorker WorkerProtocol = "http"
)

type SubscriberParams []Param;

type TaskSubscription struct {
	Protocols []WorkerProtocol `json:"protocols"`
	Type Task `json:"type"`
}

type Registry interface {
	RegisterSubscriber(group string, params SubscriberParams) (string, RegError)
	UnregisterSubscriber(group string, token string) RegError
	
	Subscribe(group string, token	string, subscriptions []TaskSubscription) RegError	
	Unsubscribe(group string, token string, task []Task) RegError
	UnsubscribeAll(group string, token string) RegError

	FindSubscribers(group string, task Task, protocols []WorkerProtocol, filters []ParamFilter, number int) ([]string, RegError)
	GetTask(group string, token string) (*TaskInstance, RegError)
	PushTask(group string, tokens []string, task *TaskInstance)	map[string]RegError
}


type RegistryServer interface {
	Start(context.Context) error
	Wait() error	
}

type RegError error;
var (
	ErrSubscriberNotFound RegError = errors.New("subscriber not found")
	ErrNotEnoughSubscribers RegError = errors.New("not enough subscribers found")
	ErrTaskNotFound RegError = errors.New("task not found")
	ErrTaskAlreadyExists RegError = errors.New("task already exists")
	ErrTaskNotSubscribed RegError = errors.New("task not subscribed")
	ErrTaskAlreadySubscribed RegError = errors.New("task already subscribed")
	ErrTokenGeneration RegError = errors.New("token generation failed")
	ErrQueue RegError = errors.New("operation on queue failed")
)

