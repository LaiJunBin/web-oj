package outputformater

var (
	HtmlFormater = NewOutputFormater("html", func(output string) string {
		return output
	})
)
