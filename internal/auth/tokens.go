package auth

import "time"

type Token struct {
	Subject string
	Expiry  time.Time
}
