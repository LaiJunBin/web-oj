package outputformater

var (
	TextFormater = NewOutputFormater("text", func(output string) string {
		return output
	})
)
