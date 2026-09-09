package outputformater

import "errors"

var (
	Formaters = map[string]OutputFormater{
		TextFormater.String(): TextFormater,
		HtmlFormater.String(): HtmlFormater,
	}

	ErrorInvalidOutputFormat = errors.New("invalid output format")
)

type OutputFormater interface {
	String() string
	Equals(other OutputFormater) bool
	FilterOutput(output string) string
}

type outputFormater struct {
	value        string
	filterOutput func(output string) string
}

func NewOutputFormater(value string, filterOutput func(output string) string) OutputFormater {
	return &outputFormater{value: value, filterOutput: filterOutput}
}

func (o *outputFormater) String() string {
	return o.value
}

func (o *outputFormater) Equals(other OutputFormater) bool {
	return o.String() == other.String()
}

func (o *outputFormater) FilterOutput(output string) string {
	return o.filterOutput(output)
}
