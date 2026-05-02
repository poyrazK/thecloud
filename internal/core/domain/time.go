package domain

import "time"

// PtrTime returns a pointer to the given time.
func PtrTime(t time.Time) *time.Time { return &t }