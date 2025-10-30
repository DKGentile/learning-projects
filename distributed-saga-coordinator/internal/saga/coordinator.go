package saga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	// ErrSagaNotFound indicates the requested saga is not registered.
	ErrSagaNotFound = errors.New("saga not found")
	// ErrInvalidConfig is returned when the configuration file cannot be parsed.
	ErrInvalidConfig = errors.New("invalid saga configuration")
)

// Coordinator is responsible for managing saga registrations and orchestrating their execution.
type Coordinator struct {
	mu          sync.RWMutex
	sagas       map[string]Saga
	dispatcher  Dispatcher
	eventLogger EventLogger
}

// Dispatcher represents the contract an execution backend must implement.
type Dispatcher interface {
	Execute(ctx context.Context, saga Saga, payload map[string]interface{}) (*ExecutionResult, error)
}

// EventLogger allows pluggable observability sinks (stdout, Kafka, tracing backends, etc.).
type EventLogger interface {
	OnSagaStarted(ctx context.Context, request ExecutionRequest)
	OnSagaCompleted(ctx context.Context, request ExecutionRequest, result ExecutionResult)
	OnSagaFailed(ctx context.Context, request ExecutionRequest, err error)
}

// NewCoordinatorFromFile loads sagas from a JSON configuration file and returns a ready-to-use coordinator.
func NewCoordinatorFromFile(configPath string) (*Coordinator, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var payload struct {
		Sagas []Saga `json:"sagas"`
	}
	if err := json.Unmarshal(file, &payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	coord := &Coordinator{
		sagas: make(map[string]Saga, len(payload.Sagas)),
		// default dispatcher simply refuses to execute until the user wires a real engine
		dispatcher:  NewNoopDispatcher(),
		eventLogger: NewStdoutLogger(),
	}

	for _, s := range payload.Sagas {
		if err := coord.registerSaga(s); err != nil {
			return nil, err
		}
	}

	return coord, nil
}

func (c *Coordinator) registerSaga(s Saga) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if s.Name == "" {
		return fmt.Errorf("saga name cannot be empty")
	}

	if _, exists := c.sagas[s.Name]; exists {
		return fmt.Errorf("duplicate saga name %q", s.Name)
	}

	if len(s.Steps) == 0 {
		return fmt.Errorf("saga %q must declare at least one step", s.Name)
	}

	c.sagas[s.Name] = s
	return nil
}

// Execute runs a saga by delegating to the configured dispatcher.
func (c *Coordinator) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	c.mu.RLock()
	s, ok := c.sagas[req.SagaName]
	c.mu.RUnlock()
	if !ok {
		return nil, ErrSagaNotFound
	}

	c.eventLogger.OnSagaStarted(ctx, req)

	result, err := c.dispatcher.Execute(ctx, s, req.Payload)
	if err != nil {
		c.eventLogger.OnSagaFailed(ctx, req, err)
		return nil, err
	}

	c.eventLogger.OnSagaCompleted(ctx, req, *result)
	return result, nil
}

// AttachDispatcher swaps the execution backend.
func (c *Coordinator) AttachDispatcher(d Dispatcher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dispatcher = d
}

// AttachEventLogger swaps the logging sink.
func (c *Coordinator) AttachEventLogger(logger EventLogger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventLogger = logger
}

// NewNoopDispatcher returns a Dispatcher that simply returns a not implemented error.
func NewNoopDispatcher() Dispatcher {
	return DispatcherFunc(func(ctx context.Context, saga Saga, payload map[string]interface{}) (*ExecutionResult, error) {
		return nil, fmt.Errorf("dispatcher not implemented for saga %q", saga.Name)
	})
}

// DispatcherFunc allows using plain functions as dispatchers.
type DispatcherFunc func(ctx context.Context, saga Saga, payload map[string]interface{}) (*ExecutionResult, error)

// Execute implements Dispatcher for DispatcherFunc.
func (f DispatcherFunc) Execute(ctx context.Context, saga Saga, payload map[string]interface{}) (*ExecutionResult, error) {
	return f(ctx, saga, payload)
}

// NewStdoutLogger writes saga lifecycle events to stdout.
func NewStdoutLogger() EventLogger {
	return stdoutLogger{}
}

type stdoutLogger struct{}

func (stdoutLogger) OnSagaStarted(ctx context.Context, req ExecutionRequest) {
	fmt.Printf("[%s] saga %s started (trace=%s)\n", time.Now().Format(time.RFC3339), req.SagaName, req.TraceID)
}

func (stdoutLogger) OnSagaCompleted(ctx context.Context, req ExecutionRequest, result ExecutionResult) {
	fmt.Printf("[%s] saga %s completed in %s\n", time.Now().Format(time.RFC3339), req.SagaName, result.FinishedAt.Sub(result.StartedAt))
}

func (stdoutLogger) OnSagaFailed(ctx context.Context, req ExecutionRequest, err error) {
	fmt.Printf("[%s] saga %s failed: %v\n", time.Now().Format(time.RFC3339), req.SagaName, err)
}
