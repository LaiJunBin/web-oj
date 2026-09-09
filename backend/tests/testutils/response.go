package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

func ResponseMatchStatusCode(resp *http.Response, code int) (err error) {
	if code != resp.StatusCode {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("expected response code to be: %d, but actual is: %d", code, resp.StatusCode)
		}

		return fmt.Errorf("expected response code to be: %d, but actual is: %d, response body: %s", code, resp.StatusCode, string(b))
	}
	return nil
}

func ResponseMatchJSON(resp *http.Response, body *godog.DocString) (err error) {
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return MatchJSON(b, body.Content)
}

// MatchJSON compares two JSON documents with the suite's wildcards, without
// requiring an HTTP response. Bundle scenarios assert on files taken out of a
// downloaded zip, which have no response of their own but deserve the same
// "{任意字}" and 1e+308 placeholders every other assertion in the suite uses.
func MatchJSON(actualJSON []byte, expectedJSON string) (err error) {
	var expected, actual interface{}

	if err = json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
		return
	}

	if err = json.Unmarshal(actualJSON, &actual); err != nil {
		return
	}

	options := cmp.Options{
		cmp.Comparer(func(x, y string) bool {
			if len(x) > 2 && x[0] == '{' && x[len(x)-1] == '}' {
				return x == y || y != ""
			}

			if len(y) > 2 && y[0] == '{' && y[len(y)-1] == '}' {
				return x == y || x != ""
			}

			return x == y
		}),
		cmp.Comparer(func(x, y float64) bool {
			if x == y {
				return true
			}

			if x == 1e+308 {
				return y > 0
			}

			if x == -1e+308 {
				return y < 0
			}

			if y == 1e+308 {
				return x > 0
			}

			if y == -1e+308 {
				return x < 0
			}

			return x == y
		}),
	}

	if diff := cmp.Diff(expected, actual, options); diff != "" {
		return fmt.Errorf("response JSON does not match expected, \n%s", diff)
	}

	return nil
}
