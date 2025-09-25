package types

import "time"

type Config struct {
	Server struct {
		Port         string
		AuthToken    string
		TimeZone     string
		TimeLocation *time.Location
	}
	ServerChan struct {
		APIKey string
	}
}
