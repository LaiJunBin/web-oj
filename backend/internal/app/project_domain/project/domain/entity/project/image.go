package project

type Image struct {
	tag string
}

func NewImage(tag string) Image {
	return Image{tag: tag}
}

func (i Image) Tag() string {
	return i.tag
}

func (i Image) Equals(other Image) bool {
	return i.tag == other.tag
}

func (i Image) IsEmpty() bool {
	return i.tag == ""
}
