package todo

import "time"

const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

func ValidPriority(p string) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}

	return false
}

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes"`
	Completed bool      `json:"completed"`
	Priority  string    `json:"priority"`
	DueDate   *string   `json:"due_date"`
	CreatedAt time.Time `json:"created_at"`
}

type Input struct {
	Title     string
	Notes     string
	Completed bool
	Priority  string
	DueDate   *string // "YYYY-MM-DD", or nil for no due date
}

type Filter struct {
	Status   string // "", "all", "active", "completed"
	Priority string // "", "low", "medium", "high"
	Search   string
	Sort     string // "", "created", "due", "priority", "title"
	Order    string // "", "asc", "desc"
}
