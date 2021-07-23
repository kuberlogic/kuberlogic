package store

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
)

type PermissionStore struct{}

const (
	ReadOnlyPrivileges = "read"
	FullPrivileges     = "all"
)

func NewPermissionStore() *PermissionStore {
	return &PermissionStore{}
}

func (ps *PermissionStore) DbToModel(dbPermissions []interfaces.Permission) []*models.Permission {
	var permissions []*models.Permission
	for _, perm := range dbPermissions {

		var type_ string
		if perm.Privilege == interfaces.FullPrivilege {
			type_ = FullPrivileges
		} else if perm.Privilege == interfaces.ReadOnlyPrivilege {
			type_ = ReadOnlyPrivileges
		}

		db := perm.Database
		permissions = append(permissions, &models.Permission{
			Database: &models.Database{
				Name: &db,
			},
			Type: type_,
		})
	}
	return permissions
}

func (ps *PermissionStore) ModelToDb(modelPermissions []*models.Permission) []interfaces.Permission {
	var permissions []interfaces.Permission
	for _, perm := range modelPermissions {

		var type_ interfaces.PrivilegeType
		if perm.Type == FullPrivileges {
			type_ = interfaces.FullPrivilege
		} else if perm.Type == ReadOnlyPrivileges {
			type_ = interfaces.ReadOnlyPrivilege
		}

		permissions = append(permissions, interfaces.Permission{
			Database:  *perm.Database.Name,
			Privilege: type_,
		})
	}
	return permissions
}
