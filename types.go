package enrow

type Company struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// ── Email Finder ──

type EmailFindParams struct {
	CompanyDomain string          `json:"company_domain,omitempty"`
	CompanyName   string          `json:"company_name,omitempty"`
	FirstName     string          `json:"first_name,omitempty"`
	LastName      string          `json:"last_name,omitempty"`
	Settings      *SearchSettings `json:"settings,omitempty"`
}

type SearchSettings struct {
	CountryCode string `json:"country_code,omitempty"`
	Webhook     string `json:"webhook,omitempty"`
}

type EmailFindResult struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	Email         string  `json:"email"`
	Qualification string  `json:"qualification"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Company       Company `json:"company"`
	Verified      bool    `json:"verified"`
	CreditsUsed   float64 `json:"credits_used"`
}

type EmailFindBulkParams struct {
	Searches []EmailFindParams `json:"searches"`
	Settings *SearchSettings   `json:"settings,omitempty"`
}

type EmailFindBulkResponse struct {
	BatchID     string  `json:"batch_id"`
	Total       int     `json:"total"`
	Status      string  `json:"status"`
	CreditsUsed float64 `json:"credits_used"`
}

type EmailFindBulkResult struct {
	BatchID     string            `json:"batch_id"`
	Status      string            `json:"status"`
	Total       int               `json:"total"`
	Completed   int               `json:"completed"`
	CreditsUsed float64           `json:"credits_used"`
	Results     []EmailFindResult `json:"results"`
}

// ── Email Verifier ──

type VerifySingleParams struct {
	Email    string              `json:"email"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

type WebhookOnlySetting struct {
	Webhook string `json:"webhook,omitempty"`
}

type VerifyChecks struct {
	SyntaxValid   bool `json:"syntax_valid"`
	DomainValid   bool `json:"domain_valid"`
	MxFound       bool `json:"mx_found"`
	SmtpValid     bool `json:"smtp_valid"`
	MailboxExists bool `json:"mailbox_exists"`
}

type VerifyMetadata struct {
	IsDisposable bool   `json:"is_disposable"`
	IsRole       bool   `json:"is_role"`
	IsFree       bool   `json:"is_free"`
	IsCatchAll   bool   `json:"is_catch_all"`
	Provider     string `json:"provider"`
}

type VerifySingleResult struct {
	ID            string         `json:"id"`
	Email         string         `json:"email"`
	Qualification string         `json:"qualification"`
	IsDeliverable bool           `json:"is_deliverable"`
	Checks        VerifyChecks   `json:"checks"`
	Metadata      VerifyMetadata `json:"metadata"`
	CreditsUsed   float64        `json:"credits_used"`
}

type VerifyBulkParams struct {
	Emails   []string            `json:"emails"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

type VerifyBulkResponse struct {
	BatchID     string  `json:"batch_id"`
	Total       int     `json:"total"`
	Status      string  `json:"status"`
	CreditsUsed float64 `json:"credits_used"`
}

type VerifyBulkResult struct {
	BatchID     string               `json:"batch_id"`
	Status      string               `json:"status"`
	Total       int                  `json:"total"`
	Completed   int                  `json:"completed"`
	CreditsUsed float64              `json:"credits_used"`
	Results     []VerifySingleResult `json:"results"`
}

// ── Phone Finder ──

type PhoneFindParams struct {
	LinkedinURL   string              `json:"linkedin_url,omitempty"`
	FirstName     string              `json:"first_name,omitempty"`
	LastName      string              `json:"last_name,omitempty"`
	CompanyDomain string              `json:"company_domain,omitempty"`
	CompanyName   string              `json:"company_name,omitempty"`
	Settings      *WebhookOnlySetting `json:"settings,omitempty"`
}

type PhoneFindResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type PhoneFindResult struct {
	ID            string `json:"id"`
	Qualification string `json:"qualification"`
	Phone         string `json:"phone"`
	CountryCode   string `json:"country_code"`
}

type PhoneFindBulkParams struct {
	Searches []PhoneFindParams   `json:"searches"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

type PhoneBulkResponse struct {
	BatchID string `json:"batch_id"`
	Total   int    `json:"total"`
	Status  string `json:"status"`
}

type PhoneBulkResultItem struct {
	Index         int    `json:"index"`
	Qualification string `json:"qualification"`
	Phone         string `json:"phone"`
	CountryCode   string `json:"country_code"`
}

type PhoneBulkResult struct {
	BatchID string                `json:"batch_id"`
	Status  string                `json:"status"`
	Total   int                   `json:"total"`
	Results []PhoneBulkResultItem `json:"results"`
}

// ── Reverse Email ──

type ReverseEmailParams struct {
	Email    string              `json:"email"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

type ReverseEmailResult struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Email       string  `json:"email"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Company     Company `json:"company"`
	LinkedinURL string  `json:"linkedin_url"`
	CreditsUsed float64 `json:"credits_used"`
}

type ReverseEmailBulkParams struct {
	Emails   []struct{ Email string `json:"email"` } `json:"emails"`
	Settings *WebhookOnlySetting                     `json:"settings,omitempty"`
}

type ReverseEmailBulkResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Total  int    `json:"total"`
}

type ReverseEmailBulkResult struct {
	ID          string                     `json:"id"`
	Status      string                     `json:"status"`
	Total       int                        `json:"total"`
	Completed   int                        `json:"completed"`
	CreditsUsed float64                    `json:"credits_used"`
	Results     []ReverseEmailBulkItemResult `json:"results"`
}

type ReverseEmailBulkItemResult struct {
	Email       string  `json:"email"`
	Status      string  `json:"status"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Company     Company `json:"company"`
	LinkedinURL string  `json:"linkedin_url"`
	Index       int     `json:"index"`
}

// ── Account ──

type AccountInfo struct {
	Credits  float64  `json:"credits"`
	Webhooks []string `json:"webhooks"`
}
