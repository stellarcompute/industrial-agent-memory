package memory

import "time"

type ScopeType string

const (
	ScopeEquipment ScopeType = "equipment"
	ScopeLine      ScopeType = "line"
	ScopeProcess   ScopeType = "process"
	ScopeAlarm     ScopeType = "alarm"
	ScopeUser      ScopeType = "user"
	ScopeSite      ScopeType = "site"
)

type MemoryKind string

const (
	KindIncident     MemoryKind = "incident"
	KindMaintenance  MemoryKind = "maintenance"
	KindParameter    MemoryKind = "parameter"
	KindProcedure    MemoryKind = "procedure"
	KindPreference   MemoryKind = "preference"
	KindObservation  MemoryKind = "observation"
	KindDecision     MemoryKind = "decision"
	KindKnownFailure MemoryKind = "known_failure"
)

type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

type Outcome string

const (
	OutcomeUnknown   Outcome = "unknown"
	OutcomeEffective Outcome = "effective"
	OutcomeRejected  Outcome = "rejected"
	OutcomeExpired   Outcome = "expired"
)

type Scope struct {
	Type ScopeType `json:"type"`
	ID   string    `json:"id"`
	Name string    `json:"name,omitempty"`
}

type Evidence struct {
	Source    string    `json:"source"`
	Reference string    `json:"reference,omitempty"`
	Observed  time.Time `json:"observed"`
	Note      string    `json:"note,omitempty"`
}

type Memory struct {
	ID         string            `json:"id"`
	Kind       MemoryKind        `json:"kind"`
	Scope      Scope             `json:"scope"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Tags       []string          `json:"tags,omitempty"`
	Signals    map[string]string `json:"signals,omitempty"`
	Evidence   []Evidence        `json:"evidence,omitempty"`
	Confidence Confidence        `json:"confidence"`
	Outcome    Outcome           `json:"outcome"`
	Importance float64           `json:"importance"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	ExpiresAt  *time.Time        `json:"expiresAt,omitempty"`
}

type Observation struct {
	Kind       MemoryKind
	Scope      Scope
	Title      string
	Content    string
	Tags       []string
	Signals    map[string]string
	Evidence   []Evidence
	Confidence Confidence
	Importance float64
	TTL        time.Duration
}

type RecallQuery struct {
	Scope       Scope
	Text        string
	Tags        []string
	Signals     map[string]string
	Kinds       []MemoryKind
	Limit       int
	IncludeWeak bool
	Now         time.Time
}

type RecallResult struct {
	Memory Memory  `json:"memory"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type ConsolidationCandidate struct {
	PrimaryID   string   `json:"primaryId"`
	DuplicateID string   `json:"duplicateId"`
	Reasons     []string `json:"reasons"`
	Score       float64  `json:"score"`
}
