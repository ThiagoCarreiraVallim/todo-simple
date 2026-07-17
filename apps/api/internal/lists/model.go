package lists

import "time"

// List is the public shape of a list. The internal uuid primary key is never
// serialized — clients only ever see the capability slug.
type List struct {
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	Position    int        `json:"position"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

type ListWithTasks struct {
	List
	Tasks []Task `json:"tasks"`
}
