package domain

import "time"

// BaseEntity represents common fields for all entities
type BaseEntity struct {
	ID        string    `bun:"id,pk" json:"id"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
