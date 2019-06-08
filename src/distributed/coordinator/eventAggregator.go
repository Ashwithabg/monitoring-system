package coordinator

import "time"

type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

type EventAggregator struct {
	listeners map[string][]func(EventData)
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(data EventData)),
	}

	return &ea
}

func (ea *EventAggregator) AddListener(name string, f func(EventData)) {
	ea.listeners[name] = append(ea.listeners[name], f)
}

func (ea *EventAggregator) PublishEvent(name string, data EventData) {
	if ea.listeners[name] != nil {
		for _, cbFunc := range ea.listeners[name] {
			cbFunc(data)
		}
	}
}
