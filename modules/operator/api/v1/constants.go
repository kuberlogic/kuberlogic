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
	RestoreUnknownStatus = "Unknown"

	Group                = "kuberlogic.com"
	apiAnnotationsGroup  = "api." + Group
	alertEmailAnnotation = apiAnnotationsGroup + "/" + "alert-email"
)
