package fed

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"sync"
)

type TaskInstance struct {
	Token string                `json:"token"`
	Type Task                   `json:"type"`
	Host string                 `json:"host"`			
	Port int                    `json:"port"`											
	Protocols []WorkerProtocol  `json:"protocols"`
}

type Payload struct {
	MimeType string 
	Data     io.ReadCloser
}

type TaskManager interface {
	GetPayload(client string) (Payload, error)
	IsValidResponseType(t string) bool
	RegisterResponse(client string, response io.ReadCloser) error
}

type TaskManagerPool interface {
	GetTaskManager(task string) (TaskManager, error)
	RemoveTaskManager(task string) error
	AddTaskManager(task string, manager TaskManager) error
}

type TaskCoordinator interface {
	Start(ctx context.Context) error
	Wait() error
}

// Mem Task Manager Pool
type MuxTaskManagerPool struct {
	Managers map[string]TaskManager
	mux sync.Mutex
}

func NewMuxTaskManagerPool() *MuxTaskManagerPool {
	return &MuxTaskManagerPool{
		Managers: make(map[string]TaskManager),
	}
}

func (p *MuxTaskManagerPool) GetTaskManager(task string) (TaskManager, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	manager, ok := p.Managers[task]
	if !ok {
		return nil, errors.New("task not found")
	}

	return manager, nil
}

func (p *MuxTaskManagerPool) RemoveTaskManager(task string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	_, ok := p.Managers[task]
	if !ok {
		return errors.New("task not found")
	}

	delete(p.Managers, task)
	return nil
}

func (p *MuxTaskManagerPool) AddTaskManager(task string, manager TaskManager) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	_, ok := p.Managers[task]
	if ok {
		return errors.New("task already exists")
	}

	p.Managers[task] = manager
	return nil
}

// ----------------- Implementation of task managers -----------------
type SimpleTaskManager struct {
	GetPayloadFunc func(client string) (Payload, error)
	IsValidResponseTypeFunc func(mime string) bool
	RegisterResponseFunc func(client string, response io.ReadCloser) error
}

func (t *SimpleTaskManager) GetPayload(client string) (Payload, error) {
	return t.GetPayloadFunc(client)
}

func (t *SimpleTaskManager) IsValidResponseType(mime string) bool {
	if t.IsValidResponseTypeFunc == nil {
		return true
	}
	
	return t.IsValidResponseTypeFunc(mime)
}

func (t *SimpleTaskManager) RegisterResponse(client string, response io.ReadCloser) error {
	return t.RegisterResponseFunc(client, response)
}


type FedTrainTaskManagerConfig struct {
	ModelFolder string
	InitialModel string
	FinalModel string
	Clients []string
	OnNextRound func(manager *FedTrainTaskManager) 
}

type FedTrainTaskManager struct {
	ModelFolder string
	InitialModel string
	FinalModel string

	OnNextRound func(manager *FedTrainTaskManager) 
	Responded map[string]bool
	mux sync.Mutex
}	


func NewFedTrainTaskManager(config *FedTrainTaskManagerConfig) *FedTrainTaskManager {
	responded :=  make(map[string]bool)

	for _, client := range config.Clients {
		responded[client] = false
	}
	
	return &FedTrainTaskManager{
		ModelFolder: config.ModelFolder,
		InitialModel: config.InitialModel,
		FinalModel: config.FinalModel,

		OnNextRound: config.OnNextRound,
		Responded: responded,
	}
}

func (t *FedTrainTaskManager) GetClients () []string {
	clients := make([]string, 0, len(t.Responded))
	for client := range t.Responded {
		clients = append(clients, client)
	}

	return clients
}


func (t *FedTrainTaskManager) IsValidResponseType(mime string) bool {
	return true
}

func (t *FedTrainTaskManager) RegisterResponse(client string, response io.ReadCloser) error {
	t.mux.Lock()
	defer t.mux.Unlock()

	alreadyResponded, ok := t.Responded[client]
	if !ok {
		return errors.New("client not allowed to respond")
	}

	if alreadyResponded {
		return errors.New("client already responded")
	}

	t.Responded[client] = true


	path, err := url.JoinPath(t.ModelFolder, client + ".npz")
	if err != nil {
		return err	
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, response)
	if err != nil {
		return err
	}	


	for _, responded := range t.Responded {
		if !responded {
			return nil
		}
	}

	t.OnNextRound(t)
	return nil
}

func (t *FedTrainTaskManager) GetPayload(client string) (Payload, error) {
	t.mux.Lock()	
	defer t.mux.Unlock()


	path, err := url.JoinPath(t.ModelFolder, t.InitialModel)
	if err != nil {
		return Payload{}, err
	}

	_, ok := t.Responded[client]
	if !ok {
		return Payload{}, errors.New("client not allowed to get payload")
	}

	file, err := os.Open(path)
	if err != nil {
		return Payload{}, err
	}

	return Payload{
		MimeType: "application/octet-stream",
		Data: file,
	}, nil
}
