package rulespb

import (
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"

	"github.com/cortexproject/cortex/pkg/cortexpb" //lint:ignore faillint allowed to import other protobuf
)

// ToProto transforms a formatted prometheus rulegroup to a rule group protobuf
func ToProto(user string, namespace string, rl rulefmt.RuleGroup) *RuleGroupDesc {
	var queryOffset *time.Duration
	if rl.QueryOffset != nil {
		offset := time.Duration(*rl.QueryOffset)
		queryOffset = &offset
	}
	rg := RuleGroupDesc{
		Name:        rl.Name,
		Namespace:   namespace,
		Interval:    time.Duration(rl.Interval),
		Rules:       formattedRuleToProto(rl.Rules),
		User:        user,
		Limit:       int64(rl.Limit),
		QueryOffset: queryOffset,
		Labels:      cortexpb.FromLabelsToLabelAdapters(labels.FromMap(rl.Labels)),
	}
	return &rg
}

func formattedRuleToProto(rls []rulefmt.RuleNode) []*RuleDesc {
	rules := make([]*RuleDesc, len(rls))
	for i := range rls {
		rules[i] = &RuleDesc{
			Expr:          rls[i].Expr.Value,
			Record:        rls[i].Record.Value,
			Alert:         rls[i].Alert.Value,
			For:           time.Duration(rls[i].For),
			KeepFiringFor: time.Duration(rls[i].KeepFiringFor),
			Labels:        cortexpb.FromLabelsToLabelAdapters(labels.FromMap(rls[i].Labels)),
			Annotations:   cortexpb.FromLabelsToLabelAdapters(labels.FromMap(rls[i].Annotations)),
		}
	}

	return rules
}

// FromProto generates a rulefmt RuleGroup
func FromProto(rg *RuleGroupDesc) rulefmt.RuleGroup {
	var queryOffset *model.Duration
	if rg.QueryOffset != nil {
		offset := model.Duration(*rg.QueryOffset)
		queryOffset = &offset
	}
	formattedRuleGroup := rulefmt.RuleGroup{
		Name:        rg.GetName(),
		Interval:    model.Duration(rg.Interval),
		Rules:       make([]rulefmt.RuleNode, len(rg.GetRules())),
		Limit:       int(rg.Limit),
		QueryOffset: queryOffset,
		Labels:      cortexpb.FromLabelAdaptersToLabels(rg.Labels).Map(),
	}

	for i, rl := range rg.GetRules() {
		exprNode := yaml.Node{}
		exprNode.SetString(rl.GetExpr())

		newRule := rulefmt.RuleNode{
			Expr:          exprNode,
			Labels:        cortexpb.FromLabelAdaptersToLabels(rl.Labels).Map(),
			Annotations:   cortexpb.FromLabelAdaptersToLabels(rl.Annotations).Map(),
			For:           model.Duration(rl.GetFor()),
			KeepFiringFor: model.Duration(rl.GetKeepFiringFor()),
		}

		if rl.GetRecord() != "" {
			recordNode := yaml.Node{}
			recordNode.SetString(rl.GetRecord())
			newRule.Record = recordNode
		} else {
			alertNode := yaml.Node{}
			alertNode.SetString(rl.GetAlert())
			newRule.Alert = alertNode
		}

		formattedRuleGroup.Rules[i] = newRule
	}

	return formattedRuleGroup
}
