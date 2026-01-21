package risk

// Decision 风控决策类型
type Decision int

const (
	// Allow 允许
	Allow Decision = iota + 1

	// Block 阻止
	Block

	// Reduce 降档（减少数量/杠杆）
	Reduce
)

func (d Decision) String() string {
	switch d {
	case Allow:
		return "ALLOW"
	case Block:
		return "BLOCK"
	case Reduce:
		return "REDUCE"
	default:
		return "UNKNOWN"
	}
}

// DecisionDetail 风控决策详情
type DecisionDetail struct {
	Decision Decision // 决策类型
	Reason   string   // 原因说明

	// Reduce 决策的建议参数
	SuggestedQuantity string // 建议数量（原始为 Money，序列化为字符串）
	SuggestedLeverage int    // 建议杠杆

	// 触发规则
	TriggeredRule string // 触发的规则名称
}

// NewAllow 创建允许决策
func NewAllow() DecisionDetail {
	return DecisionDetail{
		Decision: Allow,
		Reason:   "passed all risk checks",
	}
}

// NewBlock 创建阻止决策
func NewBlock(reason, rule string) DecisionDetail {
	return DecisionDetail{
		Decision:      Block,
		Reason:        reason,
		TriggeredRule: rule,
	}
}

// NewReduce 创建降档决策
func NewReduce(reason, rule string, suggestedQty string, suggestedLeverage int) DecisionDetail {
	return DecisionDetail{
		Decision:          Reduce,
		Reason:            reason,
		TriggeredRule:     rule,
		SuggestedQuantity: suggestedQty,
		SuggestedLeverage: suggestedLeverage,
	}
}

// IsAllowed 是否允许
func (d DecisionDetail) IsAllowed() bool {
	return d.Decision == Allow
}

// IsBlocked 是否阻止
func (d DecisionDetail) IsBlocked() bool {
	return d.Decision == Block
}

// ShouldReduce 是否需要降档
func (d DecisionDetail) ShouldReduce() bool {
	return d.Decision == Reduce
}
