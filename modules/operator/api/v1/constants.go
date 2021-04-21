package v1

const (
	// cluster status
	ClusterOkStatus       = "Ready"
	ClusterNotReadyStatus = "NotReady"
	ClusterFailedStatus   = "Failed"
	ClusterUnknownStatus  = "Unknown"

	// backup status
	BackupSuccessStatus = "Success"
	BackupFailedStatus  = "Failed"
	BackupUnknownStatus = "Unknown"

	// restore status
	RestoreSuccessStatus = "Success"
	RestoreFailedStatus  = "Failed"
	RestoreRunningStatus = "Running"

	// alert status
	AlertCreatedStatus  = "Created"
	AlertAckedStatus    = "Acknowledged"
	AlertResolvedStatus = "Resolved"
	AlertUnknownStatus  = "Unknown"
)
