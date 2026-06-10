package engine

type Type string

const (
	Noop     Type = "noop"
	Acl      Type = "acl"
	Rbac     Type = "rbac"
	Casbin   Type = "casbin"
	Opa      Type = "opa"
	Zanzibar Type = "zanzibar"
	Cedar    Type = "cedar"
	Cerbos   Type = "cerbos"
	AwsIam   Type = "awsiam"
)
