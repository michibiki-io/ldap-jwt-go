package ldapc

type Bind struct {
	BindDN       string // for LDAP Server authentication.
	BindPassword string // BindDN password
	BaseDN       string // Base search path for users
}
