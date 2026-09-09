package imagedescription

type Description struct {
	value string
}

func New(value string) Description {
	return Description{value: value}
}

func (d Description) String() string {
	return d.value
}

func (d Description) Equals(other Description) bool {
	return d.value == other.value
}
