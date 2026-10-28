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
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Notes      string    `json:"notes"`
	Completed  bool      `json:"completed"`
	Priority   string    `json:"priority"`
	DueDate    *string   `json:"due_date"`
	CategoryID *int      `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Input struct {
	Title      string
	Notes      string
	Completed  bool
	Priority   string
	DueDate    *string
	CategoryID *int
}

type Filter struct {
	Status        string
	Priority      string
	Search        string
	Sort          string
	Order         string
	CategoryID    *int
	Uncategorized bool
}
