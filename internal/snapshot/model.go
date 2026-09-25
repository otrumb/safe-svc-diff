package snapshot

type Document struct {
	SchemaVersion     string        `json:"schemaVersion"`
	ProjectionVersion string        `json:"projectionVersion"`
	Producer          Producer      `json:"producer"`
	Capture           Capture       `json:"capture"`
	Transactions      []Transaction `json:"transactions"`
}
type Producer struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type Query struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Capture struct {
	SourceOrigin     string  `json:"sourceOrigin"`
	EndpointPath     string  `json:"endpointPath"`
	Safe             string  `json:"safe"`
	Query            []Query `json:"query"`
	StartedAt        string  `json:"startedAt"`
	CompletedAt      *string `json:"completedAt"`
	Complete         bool    `json:"complete"`
	IncompleteReason *string `json:"incompleteReason"`
	PagesFetched     int     `json:"pagesFetched"`
	RecordsFetched   int     `json:"recordsFetched"`
	AdvertisedCount  *int    `json:"advertisedCount"`
	MaxPages         int     `json:"maxPages"`
	MaxRecords       int     `json:"maxRecords"`
	MaxPageBytes     int     `json:"maxPageBytes"`
	MaxTotalBytes    int     `json:"maxTotalBytes"`
}
type Transaction struct {
	SafeTxHash            string   `json:"safeTxHash"`
	Safe                  string   `json:"safe"`
	To                    string   `json:"to"`
	Value                 string   `json:"value"`
	DataSHA256            *string  `json:"dataSha256"`
	DataLength            *int     `json:"dataLength"`
	Operation             int      `json:"operation"`
	GasToken              string   `json:"gasToken"`
	SafeTxGas             string   `json:"safeTxGas"`
	BaseGas               string   `json:"baseGas"`
	GasPrice              string   `json:"gasPrice"`
	RefundReceiver        *string  `json:"refundReceiver"`
	Nonce                 string   `json:"nonce"`
	ExecutionDate         *string  `json:"executionDate"`
	SubmissionDate        string   `json:"submissionDate"`
	Modified              string   `json:"modified"`
	BlockNumber           *int     `json:"blockNumber"`
	TransactionHash       *string  `json:"transactionHash"`
	Proposer              *string  `json:"proposer"`
	ProposedByDelegate    *string  `json:"proposedByDelegate"`
	Executor              *string  `json:"executor"`
	IsExecuted            bool     `json:"isExecuted"`
	IsSuccessful          *bool    `json:"isSuccessful"`
	EthGasPrice           *string  `json:"ethGasPrice"`
	MaxFeePerGas          *string  `json:"maxFeePerGas"`
	MaxPriorityFeePerGas  *string  `json:"maxPriorityFeePerGas"`
	GasUsed               *int     `json:"gasUsed"`
	Fee                   *string  `json:"fee"`
	Payment               *string  `json:"payment"`
	ConfirmationsRequired *int     `json:"confirmationsRequired"`
	ConfirmationOwners    []string `json:"confirmationOwners"`
	Trusted               bool     `json:"trusted"`
}
