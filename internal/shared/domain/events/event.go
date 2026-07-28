package events

import "time"

// DomainEvent represents a domain event
type DomainEvent interface {
	// OccurredOn returns when the event occurred
	OccurredOn() time.Time
	// EventType returns the type of the event
	EventType() string
	// AggregateID returns the ID of the aggregate that raised the event
	AggregateID() string
}

// BaseEvent provides common fields for domain events
type BaseEvent struct {
	occurredOn  time.Time
	eventType   string
	aggregateID string
}

func NewBaseEvent(eventType, aggregateID string) *BaseEvent {
	return &BaseEvent{
		occurredOn:  time.Now(),
		eventType:   eventType,
		aggregateID: aggregateID,
	}
}

func (e *BaseEvent) OccurredOn() time.Time {
	return e.occurredOn
}

func (e *BaseEvent) EventType() string {
	return e.eventType
}

func (e *BaseEvent) AggregateID() string {
	return e.aggregateID
}
