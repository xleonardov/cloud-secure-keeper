package dummy

import (
	"github.com/xleonardov/cloud-secure-keeper/domain"
)

type adapter struct{}

func (a *adapter) ToString() string {
	return "dummy"
}

func (a *adapter) CreateRules(rules []domain.Rule) domain.AdapterResult {
	return domain.AdapterResult{}
}

func (a *adapter) DeleteRules(rules []domain.Rule) domain.AdapterResult {
	return domain.AdapterResult{}
}
