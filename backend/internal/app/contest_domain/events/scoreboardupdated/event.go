package scoreboardupdated

type Event[T any] struct {
	contestProjectID int32
	data             T
}

func New[T any](contestProjectID int32, data T) *Event[T] {
	return &Event[T]{
		contestProjectID: contestProjectID,
		data:             data,
	}
}

func (e *Event[T]) Topic() string           { return "scoreboard_updated" }
func (e *Event[T]) ContestProjectID() int32 { return e.contestProjectID }
func (e *Event[T]) Data() T                 { return e.data }
