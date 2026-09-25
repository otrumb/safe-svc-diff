package snapshot

type Document struct {
	SchemaVersion     string        `json:"schemaVersion"`
	ProjectionVersion string        `json:"projectionVersion"`
	Capture           Capture       `json:"capture"`
	Transactions      []Transaction `json:"transactions"`
}

type Capture struct {
	Safe             string  `json:"safe"`
	Complete         bool    `json:"complete"`
	CompletedAt      *string `json:"completedAt"`
	IncompleteReason *string `json:"incompleteReason"`
	PagesFetched     int     `json:"pagesFetched"`
	RecordsFetched   int     `json:"recordsFetched"`
	AdvertisedCount  *int    `json:"advertisedCount"`
}

type Transaction struct {
	SafeTxHash         string   `json:"safeTxHash"`
	Safe               string   `json:"safe"`
	ConfirmationOwners []string `json:"confirmationOwners"`
}
