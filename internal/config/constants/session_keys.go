package constants

type key int

const (
	SessionIsAuthorized key = iota
	SessionSkipAuthorization
	SessionID
	SessionIPAddress
	SessionUser
	SessionUserCompanyName
	SessionUserRole
	SessionUserHasStaffRole
	SessionUserID
	SessionUserUUID
	SessionUserTimezone
	SessionUserName
	SessionUserLexicalName
	SessionUserFirstName
	SessionUserLastName
	SessionUserTenantID
	SessionUserTenantName
)
