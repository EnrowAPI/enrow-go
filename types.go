package enrow

// The request and response types mirror the API reference:
// https://docs.enrow.io/api-reference

// ── Email Finder ──

// EmailFindParams is the body of a single email search, and one item of a bulk
// search. A search needs FullName plus CompanyDomain or CompanyName, or a
// LinkedinURL (single searches only).
type EmailFindParams struct {
	CompanyDomain string `json:"company_domain,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	FullName      string `json:"fullname,omitempty"`
	// LinkedinURL searches from a LinkedIn profile URL instead of a name and a
	// company. It is only used when the request is not a complete name +
	// company search, and bulk searches do not take it.
	LinkedinURL string `json:"linkedin_url,omitempty"`
	// Custom is returned as-is in the result and in the webhook notification.
	Custom   map[string]any  `json:"custom,omitempty"`
	Settings *SearchSettings `json:"settings,omitempty"`
}

// SearchSettings are the settings of an email search, single or bulk.
type SearchSettings struct {
	// CountryCode is the ISO 3166 Alpha-2 code of the country of the search
	// (default "US"). Relevant when searching with a company name.
	CountryCode string `json:"country_code,omitempty"`
	// Webhook is notified with an HTTP POST once the search is finished.
	Webhook        string `json:"webhook,omitempty"`
	RetrieveGender bool   `json:"retrieve_gender,omitempty"`
	// RetrieveCompanyInfo returns additional company information in
	// Info.CompanyInfo. Currently only effective when CountryCode is "FR".
	RetrieveCompanyInfo bool `json:"retrieve_company_info,omitempty"`
}

// EmailInfo is the contact information of an email search result.
type EmailInfo struct {
	// CompanyDomain is set on single search results, Domain on bulk ones.
	CompanyDomain string `json:"company_domain"`
	Domain        string `json:"domain,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	FullName      string `json:"fullname,omitempty"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	// Gender ("male" or "female") is only set when RetrieveGender is true.
	Gender string `json:"gender"`
	// CompanyInfo is only set when RetrieveCompanyInfo is true and data is available.
	CompanyInfo map[string]any `json:"companyInfo,omitempty"`
}

// EmailFindResult is the state of a single email search. Qualification is
// "valid" (email found), "invalid" (no email found) or "ongoing" (still
// running).
type EmailFindResult struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Qualification string `json:"qualification"`
	// CreditsUsed is the launch cost when the search starts, then its final
	// cost once it has ended (refunds applied).
	CreditsUsed float64    `json:"credits_used"`
	Info        *EmailInfo `json:"info,omitempty"`
	Custom      any        `json:"custom,omitempty"`
	// LinkedinStatus and LinkedinURL are only set on a search from a LinkedIn
	// profile URL. LinkedinStatus is "found", "profile_not_found",
	// "no_current_company", "email_not_found" or "lookup_failed".
	LinkedinStatus string `json:"linkedin_status,omitempty"`
	LinkedinURL    string `json:"linkedin_url,omitempty"`
}

// EmailFindBulkParams is the body of a bulk email search (up to 5,000 searches).
type EmailFindBulkParams struct {
	Searches []EmailFindParams `json:"searches"`
	// Custom is returned as-is in GetBulk (General.Custom) and in the webhook.
	Custom   map[string]any  `json:"custom,omitempty"`
	Settings *SearchSettings `json:"settings,omitempty"`
}

// EmailFindBulkResponse answers FindBulk. Pass ID to GetBulk.
type EmailFindBulkResponse struct {
	ID          string  `json:"id"`
	Message     string  `json:"message"`
	CreditsUsed float64 `json:"credits_used"`
	// EstimatedDuration is in minutes, not guaranteed.
	EstimatedDuration int `json:"estimated_duration"`
}

// BulkGeneral is the batch metadata of a bulk result.
type BulkGeneral struct {
	ID string `json:"id"`
	// Status is "ongoing", "completed" or "failed".
	Status string `json:"status"`
	// Custom is the batch-level custom data sent with the batch.
	Custom any `json:"custom,omitempty"`
}

// EmailFindBulkResult is the state of a bulk email search. Results is empty
// until General.Status is "completed".
type EmailFindBulkResult struct {
	General BulkGeneral `json:"general"`
	Stats   struct {
		Finished    int `json:"finished"`
		Requested   int `json:"requested"`
		Valid       int `json:"valid"`
		CreditsCost struct {
			Initial  float64 `json:"initial"`
			Refunded float64 `json:"refunded"`
			Final    float64 `json:"final"`
		} `json:"credits_cost"`
	} `json:"stats"`
	Results []EmailFindBulkItem `json:"results"`
}

// EmailFindBulkItem is one result of a bulk email search.
type EmailFindBulkItem struct {
	// Index identifies the search in the batch.
	Index string `json:"index"`
	// Qualification is "valid" (email found) or "invalid" (no email found).
	Qualification string `json:"qualification"`
	Email         string `json:"email"`
	// Info is only set on valid results.
	Info *EmailInfo `json:"info,omitempty"`
	// RawParams are the search parameters as sent.
	RawParams map[string]any `json:"raw_params,omitempty"`
	Custom    any            `json:"custom,omitempty"`
}

// ── Email Verifier ──

// VerifySingleParams is the body of a single email verification.
type VerifySingleParams struct {
	Email string `json:"email"`
	// Custom is returned as-is in the result and in the webhook notification.
	Custom   map[string]any      `json:"custom,omitempty"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

// WebhookOnlySetting holds the webhook notified with an HTTP POST once the
// work is finished.
type WebhookOnlySetting struct {
	Webhook string `json:"webhook,omitempty"`
}

// VerifySingleResult is the state of a single email verification.
// Qualification is "valid" (deliverable), "invalid" (not deliverable) or
// "ongoing" (still running).
type VerifySingleResult struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Qualification string `json:"qualification"`
	// CreditsUsed is answered when the verification starts: Single sets it,
	// Get alone leaves it at 0.
	CreditsUsed float64 `json:"credits_used,omitempty"`
	Custom      any     `json:"custom,omitempty"`
}

// VerifyBulkParams is the body of a bulk email verification (up to 5,000
// addresses).
type VerifyBulkParams struct {
	// Emails are the addresses to verify, sent as "verifications".
	Emails []string `json:"verifications"`
	// Custom is returned as-is in GetBulk (General.Custom) and in the webhook.
	Custom   map[string]any      `json:"custom,omitempty"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

// VerifyBulkResponse answers Bulk. Pass ID to GetBulk.
type VerifyBulkResponse struct {
	ID          string  `json:"id"`
	Message     string  `json:"message"`
	CreditsUsed float64 `json:"credits_used"`
	// EstimatedDuration is in minutes, not guaranteed.
	EstimatedDuration int `json:"estimated_duration"`
}

// VerifyBulkResult is the state of a bulk email verification. Results is
// empty until General.Status is "completed".
type VerifyBulkResult struct {
	General BulkGeneral `json:"general"`
	Stats   struct {
		// Finished is only set while the batch is ongoing.
		Finished    int     `json:"finished"`
		Requested   int     `json:"requested"`
		Valid       int     `json:"valid"`
		CreditsCost float64 `json:"credits_cost"`
	} `json:"stats"`
	Results []VerifyBulkItem `json:"results"`
}

// VerifyBulkItem is one result of a bulk email verification.
type VerifyBulkItem struct {
	Email string `json:"email"`
	// Index identifies the verification in the batch.
	Index string `json:"index"`
	// Qualification is "valid" (deliverable) or "invalid" (not deliverable).
	Qualification string `json:"qualification"`
}

// ── Phone Finder ──

// PhoneFindParams is the body of a single phone search, and one item of a bulk
// search. A search needs a LinkedinURL (recommended, it prevails), or
// FirstName + LastName + (CompanyDomain or CompanyName).
type PhoneFindParams struct {
	LinkedinURL   string `json:"linkedin_url,omitempty"`
	FirstName     string `json:"firstname,omitempty"`
	LastName      string `json:"lastname,omitempty"`
	CompanyDomain string `json:"company_domain,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	// Custom is a value of yours (an internal ID, for instance) returned in
	// the result.
	Custom   string              `json:"custom,omitempty"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

// PhoneFindResponse is the answer of the API when a single phone search starts.
type PhoneFindResponse struct {
	ID          string  `json:"id"`
	Message     string  `json:"message"`
	CreditsUsed float64 `json:"credits_used"`
}

// PhoneFindResult is the state of a single phone search. Qualification is
// "found", "not_found" or "ongoing" (still running).
type PhoneFindResult struct {
	ID            string `json:"id"`
	Qualification string `json:"qualification"`
	Number        string `json:"number"`
	// Country is an ISO 3166-1 Alpha-2 code.
	Country string `json:"country"`
	Params  any    `json:"params,omitempty"`
	Custom  any    `json:"custom,omitempty"`
}

// PhoneFindBulkParams is the body of a bulk phone search (up to 3,000 searches).
type PhoneFindBulkParams struct {
	Searches []PhoneFindParams `json:"searches"`
	// Custom is returned as-is in GetBulk (General.Custom) and in the webhook.
	Custom   map[string]any      `json:"custom,omitempty"`
	Settings *WebhookOnlySetting `json:"settings,omitempty"`
}

// PhoneBulkResponse answers FindBulk. Pass ID to GetBulk.
type PhoneBulkResponse struct {
	ID          string  `json:"id"`
	Message     string  `json:"message"`
	CreditsUsed float64 `json:"credits_used"`
	// EstimatedDuration is in minutes, not guaranteed.
	EstimatedDuration int `json:"estimated_duration"`
}

// PhoneBulkResultItem is one result of a bulk phone search.
type PhoneBulkResultItem struct {
	// Index is the position of the search in the batch.
	Index int `json:"index"`
	// Qualification is "found" or "not_found".
	Qualification string `json:"qualification"`
	Number        string `json:"number"`
	Country       string `json:"country"`
	// Params are the search parameters as sent.
	Params map[string]any `json:"params,omitempty"`
}

// PhoneBulkResult is the state of a bulk phone search. Results is empty until
// General.Status is "completed".
type PhoneBulkResult struct {
	General BulkGeneral `json:"general"`
	Stats   struct {
		// Finished is only set while the batch is ongoing.
		Finished        int     `json:"finished"`
		Requested       int     `json:"requested"`
		Found           int     `json:"found"`
		CreditsCost     float64 `json:"credits_cost"`
		CreditsToRefund float64 `json:"credits_to_refund"`
	} `json:"stats"`
	Results []PhoneBulkResultItem `json:"results"`
}

// ── Account ──

type AccountInfo struct {
	Credits  float64  `json:"credits"`
	Webhooks []string `json:"webhooks"`
}
