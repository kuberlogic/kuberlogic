package service_operator

import (
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/operator/modules/operator/service-operator/mysql"
	"github.com/kuberlogic/operator/modules/operator/service-operator/postgresql"
	"github.com/pkg/errors"
)

func GetOperator(t string) (interfaces.OperatorInterface, error) {
	var operators = map[string]interfaces.OperatorInterface{
		"postgresql": &postgresql.Postgres{},
		"mysql":      &mysql.Mysql{},
	}

	value, ok := operators[t]
	if !ok {
		return nil, errors.Errorf("Service Operator %s is not supported", t)
	}
	return value, nil
}
