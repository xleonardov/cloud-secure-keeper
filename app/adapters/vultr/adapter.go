package vultr

import (
	"github.com/JamesClonk/vultr/lib"
	"github.com/xleonardov/cloud-secure-keeper/domain"
)

// Adapter is a Vultr API implementation of the domain.Adapter interface
type Adapter struct {
	client           *lib.Client
	firewallGroupID  string
	ruleNumbersIndex map[string]int
}

type ruleFunction func(rule domain.Rule) error

// NewVultrAdapter is a constructor for Adapter
func NewVultrAdapter(apiKey string, firewallGroupID string) *Adapter {
	vultrClient := lib.NewClient(apiKey, nil)

	a := new(Adapter)
	a.client = vultrClient
	a.firewallGroupID = firewallGroupID
	a.ruleNumbersIndex = make(map[string]int)

	return a
}

// ToString satisfies the domain.Adapter interface
func (a *Adapter) ToString() string {
	return "vultr"
}

func (a *Adapter) executeForEachRule(rules []domain.Rule, function ruleFunction) domain.AdapterResult {
	for _, rule := range rules {
		err := function(rule)
		if err == nil {
			continue
		}

		return domain.AdapterResult{Error: err}
	}

	return domain.AdapterResult{}
}

// CreateRules satisfies the domain.Adapter interface
func (a *Adapter) CreateRules(rules []domain.Rule) domain.AdapterResult {
	return a.executeForEachRule(rules, a.createRule)
}

// DeleteRules satisfies the domain.Adapter interface
func (a *Adapter) DeleteRules(rules []domain.Rule) domain.AdapterResult {
	return a.executeForEachRule(rules, a.deleteRule)
}

func (a *Adapter) createRule(rule domain.Rule) (err error) {
	_, keyExists := a.ruleNumbersIndex[rule.String()]
	if keyExists {
		return // Block subsequent rule requests util it's removed by the timeout
	}

	ruleNumber, err := a.client.CreateFirewallRule(a.firewallGroupID, rule.Protocol.String(), rule.Port.String(), &rule.IPNet, "")
	if err == nil {
		a.ruleNumbersIndex[rule.String()] = ruleNumber
	}

	return
}

func (a *Adapter) deleteRule(rule domain.Rule) (err error) {
	ruleNumber, keyExists := a.ruleNumbersIndex[rule.String()]
	if !keyExists {
		return
	}

	delete(a.ruleNumbersIndex, rule.String())
	return a.client.DeleteFirewallRule(ruleNumber, a.firewallGroupID)
}
