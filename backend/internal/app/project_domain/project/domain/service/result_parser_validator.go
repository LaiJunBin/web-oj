package service

type ResultParserValidator interface {
	IsValid(parser string) bool
}
