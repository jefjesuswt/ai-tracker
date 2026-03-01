package report

type ObsidianFrontmatter struct {
	Type       string   `yaml:"type"`
	Date       string   `yaml:"date"`
	Month      string   `yaml:"month"`
	Cost       float64  `yaml:"cost"`
	Tokens     int      `yaml:"tokens"`
	Requests   int      `yaml:"requests"`
	TopModels  []string `yaml:"top_models"`
}

type Stats struct {
	ObsidianFrontmatter
	FormattedDate string
	LLMSummary string
	ModelBreakdown []ModelDetail
}

type ModelDetail struct {
	Name           string  `json:"name"`
	Reqs           int     `json:"reqs"`
	ModelCost      float64 `json:"model_cost"`
	ReqPercentage  int     `json:"req_percentage"`
	CostPercentage int     `json:"cost_percentage"`
}
