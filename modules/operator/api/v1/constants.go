package v1

const (
	// cluster status
	ClusterOkStatus       = "Ready"
	ClusterNotReadyStatus = "NotReady"
	ClusterFailedStatus   = "Failed"
	ClusterUnknownStatus  = "Unknown"

	// backup status
	BackupRunningStatus = "Running"
	BackupSuccessStatus = "Success"
	BackupFailedStatus  = "Failed"
	BackupUnknownStatus = "Unknown"

	// alert status
	AlertCreatedStatus  = "Created"
	AlertAckedStatus    = "Acknowledged"
	AlertResolvedStatus = "Resolved"
	AlertUnknownStatus  = "Unknown"
)
