package types

import (
	"github.com/gocql/gocql"
	"time"
)

type ReplyBodyType struct {
	ServerId   string
	TargetId   string
	SentAt     string
	SentAtTime time.Time
	MessageId  gocql.UUID
	Content    string
	Edited     bool
	Reactions  map[gocql.UUID]string
	SentBy     gocql.UUID
}
type Message struct {
	ServerId   string
	TargetId   string
	SentAt     string
	SentAtTime int64
	MessageId  gocql.UUID
	Content    string
	Edited     bool
	Reactions  map[gocql.UUID]string
	Reply      string
	SentBy     gocql.UUID
	ReplyBody  *ReplyBodyType
}

type EditMessage struct {
	ServerId string
	TargetId string
	SentAt string
	SentAtTime int64
	MessageId gocql.UUID
	Content string
}

type DeleteMessage struct {
	TargetId     string  `json:"TargetId"`
	ServerId     *gocql.UUID `json:"ServerId"`
	MessageId    gocql.UUID  `json:"MessageId"`
	SentAt     string      `json:"SentAt"`
	SentAtTime int64       `json:"SentAtTime"`
}