package models

import (
	"time"
)

type Message struct {
	Text      string
	Tag       string
	CreatedAt time.Time
}

type Messages []*Message
