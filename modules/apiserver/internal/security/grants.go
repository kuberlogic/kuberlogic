package security

// this file contains all the possible grants for api server actions

const (
	BackupConfigCreateSecGrant = "service:backup-config:add"
	BackupConfigDeleteSecGrant = "service:backup-config:delete"
	BackupConfigEditSecGrant   = "service:backup-config:edit"
	BackupConfigGetSecGrant    = "service:backup-config:get"
	BackupListSecGrant         = "service:backup:list"
	DatabaseCreateSecGrant     = "service:database:add"
	DatabaseDeleteSecGrant     = "service:database:delete"
	DatabaseListSecGrant       = "service:database:list"
	DatabaseRestoreSecGrant    = "service:backup-restore:create"
	LogsGetSecGrant            = "service:logs"
	RestoreListSecGrant        = "service:restore:list"
	ServiceAddSecGrant         = "service:add"
	ServiceDeleteSecGrant      = "service:delete"
	ServiceEditSecGrant        = "service:edit"
	ServiceGetSecGrant         = "service:get"
	ServiceListSecGrant        = "service:list"
	UserCreateSecGrant         = "service:user:add"
	UserDeleteSecGrant         = "service:user:delete"
	UserEditSecGrant           = "service:user:edit"
	UserListSecGrant           = "service:user:list"
)

var (
	ServiceGrants = []string{
		BackupConfigCreateSecGrant,
		BackupConfigDeleteSecGrant,
		BackupConfigEditSecGrant,
		BackupConfigGetSecGrant,
		BackupListSecGrant,
		DatabaseCreateSecGrant,
		DatabaseDeleteSecGrant,
		DatabaseListSecGrant,
		DatabaseRestoreSecGrant,
		LogsGetSecGrant,
		RestoreListSecGrant,
		ServiceAddSecGrant,
		ServiceDeleteSecGrant,
		ServiceEditSecGrant,
		ServiceGetSecGrant,
		ServiceListSecGrant,
		UserCreateSecGrant,
		UserDeleteSecGrant,
		UserEditSecGrant,
		UserListSecGrant,
	}
)
