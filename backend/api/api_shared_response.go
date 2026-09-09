package api

type SharedAPIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type SharedAPIErrorResponse struct {
	StatusCode int      `json:"-"`
	Details    []string `json:"details"`
	Message    string   `json:"message"`
	Success    bool     `json:"success"`
}
