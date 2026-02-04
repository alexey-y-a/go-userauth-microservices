package clickhouse

import "time"

type AuthEvent struct {
    UserID    string
    IP        string
    EventType string
    OccuredAt time.Time
}