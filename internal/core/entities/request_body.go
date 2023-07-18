package entities

type WriteRequestBody struct {
	ResourceId string      `json:"resource_id"`
	Force      string      `json:"force"`
	Method     string      `json:"method"`
	Records    []GateCount `json:"records"`
}

type ReadRecordBody struct {
	ResourceId string `json:"resource_id"`
	Limit      int    `json:"limit"`
	Sort       string `json:"sort"`
}
