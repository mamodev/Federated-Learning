package fed

import (
	"sync"
	"time"
)

type Subscriptions map[Task][]WorkerProtocol
type Subscriber struct {
	Token string
	Params map[string]Param
	LatestActivity int64
	Subscribed Subscriptions
	Queue TaskQueue
}

func (s *Subscriber) MatchFilters(filters []ParamFilter) bool {
	for _, filter := range filters {
		param, ok := s.Params[filter.Name]
		if !ok {
			return false
		}

		if !param.MatchFilter(filter) {
			return false
		}
	}

	return true
}


type MemRegistry struct {
	Subscribers map[string]*Subscriber
	queueFactory TaskQueueFactory
	lock sync.RWMutex
}

func (r *MemRegistry) getKey(group, token string) string {
	return group + "-" + token
}

func NewMemRegistry(queueFactory TaskQueueFactory) *MemRegistry {
	return &MemRegistry{
		Subscribers: make(map[string]*Subscriber),
		queueFactory: queueFactory,
	}
}

func (r *MemRegistry) RegisterSubscriber(group string, params SubscriberParams) (string, RegError) {
	token, err :=  generateRandomKey(16);
	if err != nil {
		return "", ErrTokenGeneration
	}	

	var pmap = make(map[string]Param)
	for _, p := range params {
		pmap[p.Name] = p
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.Subscribers[r.getKey(group, token)] = &Subscriber{
		Token: token,
		Params: pmap,
		LatestActivity: time.Now().Unix(),
		Subscribed: make(Subscriptions),
		Queue: r.queueFactory(),
	}

	return token, nil
}

func (r *MemRegistry) UnregisterSubscriber(group string, token string) RegError {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.Subscribers, token)
	return nil
}


func (r *MemRegistry) Subscribe(group string, token string, subscriptions []TaskSubscription) RegError {
	r.lock.Lock()
	defer r.lock.Unlock()

	subscriber, ok := r.Subscribers[r.getKey(group, token)]
	if !ok {
		return ErrSubscriberNotFound
	}

	for _, sub  := range subscriptions {
		_, ok := subscriber.Subscribed[sub.Type]
		if ok { 
			return ErrTaskAlreadySubscribed
		}
	}
	
	for _, sub  := range subscriptions {
		subscriber.Subscribed[sub.Type] = sub.Protocols
	}

	return nil
}

func (r *MemRegistry) Unsubscribe(group string, token string, tasks []Task) RegError {
	r.lock.Lock()
	defer r.lock.Unlock()

	subscriber, ok := r.Subscribers[r.getKey(group, token)]
	if !ok {
		return ErrSubscriberNotFound
	}

	for _, task := range tasks {
		delete(subscriber.Subscribed, task)
	}

	return nil
}

func (r *MemRegistry) UnsubscribeAll(group string, token string) RegError {
	r.lock.Lock()
	defer r.lock.Unlock()

	subscriber, ok := r.Subscribers[r.getKey(group, token)]
	if !ok {
		return ErrSubscriberNotFound
	}

	subscriber.Subscribed = make(Subscriptions)
	return nil
}

func (r *MemRegistry) FindSubscribers(group string, task Task, protocols []WorkerProtocol, filters []ParamFilter, number int) ([]string, RegError) {
	r.lock.RLock()	
	defer r.lock.RUnlock()	

	var tokens []string
	for token, subscriber := range r.Subscribers {
		groupToken := token[:len(group)]
		if groupToken != group {
			continue
		}

		if _, ok := subscriber.Subscribed[task]; !ok {
			continue
		}

		if !subscriber.MatchFilters(filters) {
			continue
		}
		
		tokens = append(tokens, token[len(group)+1:])

		if len(tokens) == number {
			break
		}
	}

	if len(tokens) == 0 {
		return nil, ErrSubscriberNotFound
	}

	return tokens, nil
}

func (r *MemRegistry) GetTask(group string, token string) (*TaskInstance, RegError) {
	r.lock.RLock()	
	defer r.lock.RUnlock()	

	sub, ok := r.Subscribers[r.getKey(group, token)]
	if !ok {
		return nil, ErrSubscriberNotFound
	}

	task, err := sub.Queue.Dequeue()
	if err != nil {
		return nil, ErrQueue
	}

	if task == nil {
		return nil, nil
	}

	copy := *task
	return &copy, nil
}


func (r *MemRegistry) PushTask(group string, tokens []string, task *TaskInstance) map[string]RegError {
	r.lock.Lock()
	defer r.lock.Unlock()

	errors := make(map[string]RegError)
	for _, token := range tokens {
		sub, ok := r.Subscribers[r.getKey(group, token)]
		if !ok {
			errors[token] = ErrSubscriberNotFound
			continue
		}

		_, ok = sub.Subscribed[task.Type]
		if !ok {
			errors[token] = ErrTaskNotSubscribed
			continue
		}
	}

	for _, token := range tokens {
		sub := r.Subscribers[r.getKey(group, token)]

		err := sub.Queue.Enqueue(task)
		if err != nil {
			return map[string]RegError{token: ErrQueue}
		}
	}

	return nil
}
