package config

type CtxKey string

const (
	KeyUserID = CtxKey("user_id")
	KeyOrgID  = CtxKey("org_id")
)
