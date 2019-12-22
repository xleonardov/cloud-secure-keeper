package vpc

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/nstapelbroek/gatekeeper/domain"
)

type adapter struct {
	client           *ec2.Client
	NetworkAclId     string
	startRuleNumber  int64
	ruleStepsTaken   int
	ruleStepSize     int
	ruleNumbersIndex map[string]int64
}

func NewAWSNetworkACLAdapter(client *ec2.Client, networkAclId string, startRuleNumber int64) *adapter {
	return &adapter{
		client:           client,
		NetworkAclId:     networkAclId,
		startRuleNumber:  startRuleNumber,
		ruleStepsTaken:   0,
		ruleStepSize:     10,
		ruleNumbersIndex: make(map[string]int64),
	}
}

func (a *adapter) ToString() string {
	return "aws-network-acl"
}

func (a *adapter) getNextRuleNumber() *int64 {
	a.ruleStepsTaken = a.ruleStepsTaken + 1
	return aws.Int64(a.startRuleNumber + int64(a.ruleStepSize*a.ruleStepsTaken))
}

func (a *adapter) getProtocolNumber(protocol domain.Protocol) *string {
	if protocol == domain.TCP {
		return aws.String("6")
	}

	if protocol == domain.UDP {
		return aws.String("17")
	}

	if protocol == domain.ICMP {
		return aws.String("1")
	}

	// Fallback to all protocols
	return aws.String("-1")
}

func (a *adapter) buildCreateAclEntryRequest(rule domain.Rule) *ec2.CreateNetworkAclEntryInput {
	input := ec2.CreateNetworkAclEntryInput{
		Egress:       aws.Bool(rule.Direction.IsOutbound()),
		NetworkAclId: aws.String(a.NetworkAclId),
		PortRange:    &ec2.PortRange{From: aws.Int64(int64(rule.Port.BeginPort)), To: aws.Int64(int64(rule.Port.EndPort))},
		Protocol:     a.getProtocolNumber(rule.Protocol),
		RuleAction:   "allow",
		RuleNumber:   a.getNextRuleNumber(),
	}

	if rule.IPNet.IP.To4() == nil {
		input.Ipv6CidrBlock = aws.String(rule.IPNet.String())
	} else {
		input.CidrBlock = aws.String(rule.IPNet.String())
	}

	return &input
}

func (a *adapter) CreateRules(rules []domain.Rule) (result domain.AdapterResult) {
	for _, rule := range rules {
		input := a.buildCreateAclEntryRequest(rule)
		req := a.client.CreateNetworkAclEntryRequest(input)
		_, _ = req.Send(context.TODO()) // TODO error handling
		a.ruleNumbersIndex[rule.String()] = *input.RuleNumber
	}

	return domain.AdapterResult{}
}

func (a *adapter) DeleteRules(rules []domain.Rule) (result domain.AdapterResult) {
	for _, rule := range rules {
		ruleNumber, keyExists := a.ruleNumbersIndex[rule.String()]
		if !keyExists {
			return
		}

		input := ec2.DeleteNetworkAclEntryInput{
			Egress:       aws.Bool(rule.Direction.IsOutbound()),
			NetworkAclId: aws.String(a.NetworkAclId),
			RuleNumber:   &ruleNumber,
		}

		req := a.client.DeleteNetworkAclEntryRequest(&input)
		_, _ = req.Send(context.TODO()) // TODO error handling
		delete(a.ruleNumbersIndex, rule.String())
	}

	return domain.AdapterResult{}
}
