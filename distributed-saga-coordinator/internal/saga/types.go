package saga

import "time"

type (
	// Step represents an action within a saga. CompensatingAction is optional for steps that can fail safely.
	Step struct {
		Name               string            `json:"name"`
		Action             string            `json:"action"`
		CompensatingAction string            `json:"compensating_action"`
		Timeout            time.Duration     `json:"timeout"`
		Requires           []string          `json:"requires"`
		Metadata           map[string]string `json:"metadata"`
	}

	// Saga is a collection of ordered steps with dependency constraints.
	Saga struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Steps       []Step            `json:"steps"`
		Metadata    map[string]string `json:"metadata"`
	}
)

// ExecutionRequest describes a runtime invocation of a saga.
type ExecutionRequest struct {
	SagaName string                 `json:"saga_name"`
	Payload  map[string]interface{} `json:"payload"`
	TraceID  string                 `json:"trace_id"`
}

// ExecutionResult captures the final state of a saga run.
type ExecutionResult struct {
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Status     string                 `json:"status"`
	Details    map[string]interface{} `json:"details"`
}
