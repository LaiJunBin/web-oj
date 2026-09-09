package project

type Description string

func NewDescription(description string) Description {
	return Description(description)
}

func (d Description) String() string {
	return string(d)
}

func (d Description) Equals(other Description) bool {
	return d == other
}

func (d Description) IsEmpty() bool {
	return d == ""
}
