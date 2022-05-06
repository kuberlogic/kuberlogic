package cli

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"

	"github.com/spf13/cobra"
)

// Schema cli for Service

// register flags to command
func registerModelServiceFlags(depth int, cmdPrefix string, cmd *cobra.Command) error {

	if err := registerServiceAdvanced(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceCreatedAt(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceLimits(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceName(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceNs(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceReplicas(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceStatus(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceType(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceVersion(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	if err := registerServiceVolumeSize(depth, cmdPrefix, cmd); err != nil {
		return err
	}

	return nil
}

func registerServiceAdvanced(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	// warning: advanced Advanced map type is not supported by go-swagger cli yet

	return nil
}

func registerServiceCreatedAt(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	createdAtDescription := ``

	var createdAtFlagName string
	if cmdPrefix == "" {
		createdAtFlagName = "created_at"
	} else {
		createdAtFlagName = fmt.Sprintf("%v.created_at", cmdPrefix)
	}

	_ = cmd.PersistentFlags().String(createdAtFlagName, "", createdAtDescription)

	return nil
}

func registerServiceLimits(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	var limitsFlagName string
	if cmdPrefix == "" {
		limitsFlagName = "limits"
	} else {
		limitsFlagName = fmt.Sprintf("%v.limits", cmdPrefix)
	}

	if err := registerModelLimitsFlags(depth+1, limitsFlagName, cmd); err != nil {
		return err
	}

	return nil
}

func registerServiceName(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	nameDescription := `Required. `

	var nameFlagName string
	if cmdPrefix == "" {
		nameFlagName = "name"
	} else {
		nameFlagName = fmt.Sprintf("%v.name", cmdPrefix)
	}

	var nameFlagDefault string

	_ = cmd.PersistentFlags().String(nameFlagName, nameFlagDefault, nameDescription)

	return nil
}

func registerServiceNs(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	nsDescription := ``

	var nsFlagName string
	if cmdPrefix == "" {
		nsFlagName = "ns"
	} else {
		nsFlagName = fmt.Sprintf("%v.ns", cmdPrefix)
	}

	var nsFlagDefault string

	_ = cmd.PersistentFlags().String(nsFlagName, nsFlagDefault, nsDescription)

	return nil
}

func registerServiceReplicas(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	replicasDescription := ``

	var replicasFlagName string
	if cmdPrefix == "" {
		replicasFlagName = "replicas"
	} else {
		replicasFlagName = fmt.Sprintf("%v.replicas", cmdPrefix)
	}

	var replicasFlagDefault int64

	_ = cmd.PersistentFlags().Int64(replicasFlagName, replicasFlagDefault, replicasDescription)

	return nil
}

func registerServiceStatus(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	statusDescription := ``

	var statusFlagName string
	if cmdPrefix == "" {
		statusFlagName = "status"
	} else {
		statusFlagName = fmt.Sprintf("%v.status", cmdPrefix)
	}

	var statusFlagDefault string

	_ = cmd.PersistentFlags().String(statusFlagName, statusFlagDefault, statusDescription)

	return nil
}

func registerServiceType(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	typeDescription := `Required. `

	var typeFlagName string
	if cmdPrefix == "" {
		typeFlagName = "type"
	} else {
		typeFlagName = fmt.Sprintf("%v.type", cmdPrefix)
	}

	var typeFlagDefault string

	_ = cmd.PersistentFlags().String(typeFlagName, typeFlagDefault, typeDescription)

	return nil
}

func registerServiceVersion(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	versionDescription := ``

	var versionFlagName string
	if cmdPrefix == "" {
		versionFlagName = "version"
	} else {
		versionFlagName = fmt.Sprintf("%v.version", cmdPrefix)
	}

	var versionFlagDefault string

	_ = cmd.PersistentFlags().String(versionFlagName, versionFlagDefault, versionDescription)

	return nil
}

func registerServiceVolumeSize(depth int, cmdPrefix string, cmd *cobra.Command) error {
	if depth > maxDepth {
		return nil
	}

	volumeSizeDescription := ``

	var volumeSizeFlagName string
	if cmdPrefix == "" {
		volumeSizeFlagName = "volumeSize"
	} else {
		volumeSizeFlagName = fmt.Sprintf("%v.volumeSize", cmdPrefix)
	}

	var volumeSizeFlagDefault string

	_ = cmd.PersistentFlags().String(volumeSizeFlagName, volumeSizeFlagDefault, volumeSizeDescription)

	return nil
}

// retrieve flags from commands, and set value in model. Return true if any flag is passed by user to fill model field.
func retrieveModelServiceFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	retAdded := false

	err, advancedAdded := retrieveServiceAdvancedFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || advancedAdded

	err, createdAtAdded := retrieveServiceCreatedAtFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || createdAtAdded

	err, limitsAdded := retrieveServiceLimitsFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || limitsAdded

	err, nameAdded := retrieveServiceNameFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || nameAdded

	err, nsAdded := retrieveServiceNsFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || nsAdded

	err, replicasAdded := retrieveServiceReplicasFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || replicasAdded

	err, statusAdded := retrieveServiceStatusFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || statusAdded

	err, typeAdded := retrieveServiceTypeFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || typeAdded

	err, versionAdded := retrieveServiceVersionFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || versionAdded

	err, volumeSizeAdded := retrieveServiceVolumeSizeFlags(depth, m, cmdPrefix, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || volumeSizeAdded

	return nil, retAdded
}

func retrieveServiceAdvancedFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	advancedFlagName := fmt.Sprintf("%v.advanced", cmdPrefix)
	if cmd.Flags().Changed(advancedFlagName) {
		// warning: advanced map type Advanced is not supported by go-swagger cli yet
	}

	return nil, retAdded
}

func retrieveServiceCreatedAtFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	createdAtFlagName := fmt.Sprintf("%v.created_at", cmdPrefix)
	if cmd.Flags().Changed(createdAtFlagName) {

		var createdAtFlagName string
		if cmdPrefix == "" {
			createdAtFlagName = "created_at"
		} else {
			createdAtFlagName = fmt.Sprintf("%v.created_at", cmdPrefix)
		}

		createdAtFlagValueStr, err := cmd.Flags().GetString(createdAtFlagName)
		if err != nil {
			return err, false
		}
		var createdAtFlagValue strfmt.DateTime
		if err := createdAtFlagValue.UnmarshalText([]byte(createdAtFlagValueStr)); err != nil {
			return err, false
		}
		m.CreatedAt = createdAtFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceLimitsFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	limitsFlagName := fmt.Sprintf("%v.limits", cmdPrefix)
	if cmd.Flags().Changed(limitsFlagName) {
		// info: complex object limits Limits is retrieved outside this Changed() block
	}
	limitsFlagValue := m.Limits
	if swag.IsZero(limitsFlagValue) {
		limitsFlagValue = &models.Limits{}
	}

	err, limitsAdded := retrieveModelLimitsFlags(depth+1, limitsFlagValue, limitsFlagName, cmd)
	if err != nil {
		return err, false
	}
	retAdded = retAdded || limitsAdded
	if limitsAdded {
		m.Limits = limitsFlagValue
	}

	return nil, retAdded
}

func retrieveServiceNameFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	nameFlagName := fmt.Sprintf("%v.name", cmdPrefix)
	if cmd.Flags().Changed(nameFlagName) {

		var nameFlagName string
		if cmdPrefix == "" {
			nameFlagName = "name"
		} else {
			nameFlagName = fmt.Sprintf("%v.name", cmdPrefix)
		}

		nameFlagValue, err := cmd.Flags().GetString(nameFlagName)
		if err != nil {
			return err, false
		}
		m.Name = &nameFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceNsFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	nsFlagName := fmt.Sprintf("%v.ns", cmdPrefix)
	if cmd.Flags().Changed(nsFlagName) {

		var nsFlagName string
		if cmdPrefix == "" {
			nsFlagName = "ns"
		} else {
			nsFlagName = fmt.Sprintf("%v.ns", cmdPrefix)
		}

		nsFlagValue, err := cmd.Flags().GetString(nsFlagName)
		if err != nil {
			return err, false
		}
		m.Ns = nsFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceReplicasFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	replicasFlagName := fmt.Sprintf("%v.replicas", cmdPrefix)
	if cmd.Flags().Changed(replicasFlagName) {

		var replicasFlagName string
		if cmdPrefix == "" {
			replicasFlagName = "replicas"
		} else {
			replicasFlagName = fmt.Sprintf("%v.replicas", cmdPrefix)
		}

		replicasFlagValue, err := cmd.Flags().GetInt64(replicasFlagName)
		if err != nil {
			return err, false
		}
		m.Replicas = &replicasFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceStatusFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	statusFlagName := fmt.Sprintf("%v.status", cmdPrefix)
	if cmd.Flags().Changed(statusFlagName) {

		var statusFlagName string
		if cmdPrefix == "" {
			statusFlagName = "status"
		} else {
			statusFlagName = fmt.Sprintf("%v.status", cmdPrefix)
		}

		statusFlagValue, err := cmd.Flags().GetString(statusFlagName)
		if err != nil {
			return err, false
		}
		m.Status = statusFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceTypeFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	typeFlagName := fmt.Sprintf("%v.type", cmdPrefix)
	if cmd.Flags().Changed(typeFlagName) {

		var typeFlagName string
		if cmdPrefix == "" {
			typeFlagName = "type"
		} else {
			typeFlagName = fmt.Sprintf("%v.type", cmdPrefix)
		}

		typeFlagValue, err := cmd.Flags().GetString(typeFlagName)
		if err != nil {
			return err, false
		}
		m.Type = &typeFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceVersionFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	versionFlagName := fmt.Sprintf("%v.version", cmdPrefix)
	if cmd.Flags().Changed(versionFlagName) {

		var versionFlagName string
		if cmdPrefix == "" {
			versionFlagName = "version"
		} else {
			versionFlagName = fmt.Sprintf("%v.version", cmdPrefix)
		}

		versionFlagValue, err := cmd.Flags().GetString(versionFlagName)
		if err != nil {
			return err, false
		}
		m.Version = versionFlagValue

		retAdded = true
	}

	return nil, retAdded
}

func retrieveServiceVolumeSizeFlags(depth int, m *models.Service, cmdPrefix string, cmd *cobra.Command) (error, bool) {
	if depth > maxDepth {
		return nil, false
	}
	retAdded := false

	volumeSizeFlagName := fmt.Sprintf("%v.volumeSize", cmdPrefix)
	if cmd.Flags().Changed(volumeSizeFlagName) {

		var volumeSizeFlagName string
		if cmdPrefix == "" {
			volumeSizeFlagName = "volumeSize"
		} else {
			volumeSizeFlagName = fmt.Sprintf("%v.volumeSize", cmdPrefix)
		}

		volumeSizeFlagValue, err := cmd.Flags().GetString(volumeSizeFlagName)
		if err != nil {
			return err, false
		}
		m.VolumeSize = volumeSizeFlagValue

		retAdded = true
	}

	return nil, retAdded
}
