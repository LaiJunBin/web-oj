package judgeorder

import "fmt"

var (
	JUDGE         = judgeType("JUDGE")
	JUDGE_PARTIAL = judgeType("JUDGE_PARTIAL")
	REJUDGE       = judgeType("REJUDGE")
)

type judgeType string

type JudgeOrder struct {
	value judgeType
}

func New(value string) (*JudgeOrder, error) {
	switch value {
	case "JUDGE":
		return &JudgeOrder{value: JUDGE}, nil
	case "JUDGE_PARTIAL":
		return &JudgeOrder{value: JUDGE_PARTIAL}, nil
	case "REJUDGE":
		return &JudgeOrder{value: REJUDGE}, nil
	default:
		return nil, fmt.Errorf("invalid judge order value: %s", value)
	}
}

func (j JudgeOrder) String() string {
	return string(j.value)
}
