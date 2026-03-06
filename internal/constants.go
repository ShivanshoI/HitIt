package internal

const (
	AppName   = "Hit-IT"
	APIPrefix = "/api"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userId"
	TeamIDKey ContextKey = "teamId"
)
