package notion

type BlockURL struct {
	URL string `json:"url"`
}

type RecordProperties struct {
	ApplyURL BlockURL `json:"Apply URL"`
}

type Page struct {
	Properties RecordProperties `json:"properties"`
}

type lookupRecordResponse struct {
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}
