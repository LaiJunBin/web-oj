package judgetype

var (
	JUDGE         = declareJudgeType("JUDGE")
	JUDGE_PARTIAL = declareJudgeType("JUDGE_PARTIAL")
	REJUDGE       = declareJudgeType("REJUDGE")
)

var JudgeTypeMap = map[string]JudgeType{}

type JudgeType struct {
	value string
}

func (j JudgeType) String() string {
	return j.value
}

func declareJudgeType(judgeType string) JudgeType {
	j := JudgeType{
		value: judgeType,
	}
	JudgeTypeMap[judgeType] = j
	return j
}

func (j JudgeType) IsJudge() bool {
	return j == JUDGE
}

func (j JudgeType) IsJudgePartial() bool {
	return j == JUDGE_PARTIAL
}

func (j JudgeType) IsRejudge() bool {
	return j == REJUDGE
}
