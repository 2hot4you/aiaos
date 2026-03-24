package queue

import "context"

// Producer is the task queue producer interface (mock for Phase 1)
type Producer interface {
	Publish(ctx context.Context, taskType string, payload interface{}) (string, error)
}

// MockProducer is a no-op producer for Phase 1
type MockProducer struct{}

func NewMockProducer() *MockProducer {
	return &MockProducer{}
}

func (p *MockProducer) Publish(_ context.Context, _ string, _ interface{}) (string, error) {
	return "mock-task-id", nil
}
