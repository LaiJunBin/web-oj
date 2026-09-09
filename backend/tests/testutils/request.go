package testutils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Request(method, url string, data *string, headers map[string]string, cookies map[string]string) (resp *http.Response, err error) {
	defer func() {
		switch t := recover().(type) {
		case string:
			err = fmt.Errorf(t, "panic occurred")
		case error:
			err = t
		}
	}()

	var body *bytes.Buffer = bytes.NewBuffer([]byte(""))
	if data != nil {
		body = bytes.NewBuffer([]byte(*data))
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if req.Header.Get("Host") != "" {
		req.Host = req.Header.Get("Host")
	}

	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// projectFixtureName pulls <name> out of a "{<name> project}" placeholder,
// naming tests/data/<name>.zip.
func projectFixtureName(placeholder string) (string, bool) {
	if !strings.HasPrefix(placeholder, "{") || !strings.HasSuffix(placeholder, " project}") {
		return "", false
	}

	name := strings.TrimSuffix(strings.TrimPrefix(placeholder, "{"), " project}")
	if name == "" {
		return "", false
	}

	return name, true
}

// attachFile adds src to the multipart form under field. Errors are returned
// rather than printed: a fixture that fails to attach otherwise turns into a
// confusing assertion failure much later in the scenario.
func attachFile(writer *multipart.Writer, field string, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(field, filepath.Base(file.Name()))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	return err
}

func FormRequest(method, url string, data map[string]string, headers map[string]string, cookies map[string]string) (resp *http.Response, err error) {
	defer func() {
		switch t := recover().(type) {
		case string:
			err = fmt.Errorf(t, "panic occurred")
		case error:
			err = t
		}
	}()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	for k, v := range data {
		if k == "project" || k == "file" {
			// An empty value attaches nothing on purpose — that is how a
			// scenario asks the server what it does with a missing file.
			if v == "" {
				continue
			}

			// "{<name> project}" attaches tests/data/<name>.zip. Adding a
			// fixture used to mean adding another near-identical branch here,
			// so the name is resolved instead of enumerated. A value that
			// looks like a placeholder but does not resolve is a typo worth
			// failing on rather than silently sending no file.
			name, ok := projectFixtureName(v)
			if !ok {
				return nil, fmt.Errorf("unrecognised project fixture %q, expected \"{<name> project}\" or an empty value", v)
			}

			if err := attachFile(writer, k, path.Join(Root, "/data/", name+".zip")); err != nil {
				return nil, err
			}
		} else if k == "users" {
			if err := attachFile(writer, "users", path.Join(Root, "/data/", v[1:len(v)-1])); err != nil {
				return nil, err
			}
		} else {
			p, err := writer.CreateFormField(k)
			if err != nil {
				return nil, err
			}

			_, err = p.Write([]byte(v))
			if err != nil {
				return nil, err
			}
		}
	}

	writer.Close()

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if req.Header.Get("Host") != "" {
		req.Host = req.Header.Get("Host")
	}

	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}
