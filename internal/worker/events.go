package worker

import "time"

type EventType string

const (
	EventTaskCreated  EventType = "task.created"
	EventTaskAssigned EventType = "task.assigned"
	EventTaskUpdated  EventType = "task.updated"
	EventCommentAdded EventType = "comment.added"
	EventMemberAdded  EventType = "project.member_added"
)

type Event struct {
	Type      EventType
	UserID    int64
	ActorID   int64
	ProjectID int64
	TaskID    int64
	Message   string
	OccurredAt time.Time
}
