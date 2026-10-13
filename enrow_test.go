package enrow

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

// fakeAPI is a test server that records every request and answers with the
// replies queued for its route ("METHOD /path"), the last one repeating.
type fakeAPI struct {
	mu       sync.Mutex
	replies  map[string][]reply
	requests []recorded
}

type reply struct {
	status int
	body   string
}

type recorded struct {
	route       string
	id          string // the id query parameter
	apiKey      string
	contentType string
	body        []byte
}

func newFakeAPI(t *testing.T) (*fakeAPI, *Client) {
	t.Helper()
	api := &fakeAPI{replies: map[string][]reply{}}
	server := httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(server.Close)
	return api, New("test_key", WithBaseURL(server.URL))
}

func (f *fakeAPI) on(route string, status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[route] = append(f.replies[route], reply{status, body})
}

func (f *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	route := r.Method + " " + r.URL.Path

	f.mu.Lock()
	f.requests = append(f.requests, recorded{
		route:       route,
		id:          r.URL.Query().Get("id"),
		apiKey:      r.Header.Get("x-api-key"),
		contentType: r.Header.Get("Content-Type"),
		body:        body,
	})
	rep := reply{500, `{"message":"no reply queued for ` + route + `"}`}
	if queue := f.replies[route]; len(queue) > 0 {
		rep = queue[0]
		if len(queue) > 1 {
			f.replies[route] = queue[1:]
		}
	}
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rep.status)
	_, _ = io.WriteString(w, rep.body)
}

// calls returns the requests received on route.
func (f *fakeAPI) calls(route string) []recorded {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []recorded
	for _, r := range f.requests {
		if r.route == route {
			out = append(out, r)
		}
	}
	return out
}

// only returns the one request received on route.
func (f *fakeAPI) only(t *testing.T, route string) recorded {
	t.Helper()
	calls := f.calls(route)
	if len(calls) != 1 {
		t.Fatalf("%s: got %d requests, want 1", route, len(calls))
	}
	return calls[0]
}

// assertJSON fails unless the request body holds the same JSON value as want.
func assertJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("request body is not JSON: %v (%s)", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad expected JSON: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("request body:\n got %s\nwant %s", got, want)
	}
}

func assertIDs(t *testing.T, calls []recorded, id string) {
	t.Helper()
	if len(calls) == 0 {
		t.Fatal("no GET request")
	}
	for _, c := range calls {
		if c.id != id {
			t.Fatalf("%s sent id=%q, want %q", c.route, c.id, id)
		}
	}
}

var fastPoll = &PollOptions{WaitForResult: true, PollInterval: time.Millisecond, Timeout: 2 * time.Second}

// ── Email Finder ──

func TestEmailFindSendsTheDocumentedBody(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/find/single", 200, `{"message":"Single search operating","id":"search_1","credits_used":1}`)

	result, err := client.Email.Find(EmailFindParams{
		FullName:      "Dwight Schrute",
		CompanyDomain: "dundermifflin.com",
		Custom:        map[string]any{"crm_id": "lead_001"},
		Settings: &SearchSettings{
			Webhook:             "https://your-app.com/webhooks/enrow",
			CountryCode:         "FR",
			RetrieveGender:      true,
			RetrieveCompanyInfo: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := api.only(t, "POST /email/find/single")
	if req.apiKey != "test_key" || req.contentType != "application/json" {
		t.Fatalf("headers: x-api-key=%q content-type=%q", req.apiKey, req.contentType)
	}
	assertJSON(t, req.body, `{
		"fullname": "Dwight Schrute",
		"company_domain": "dundermifflin.com",
		"custom": {"crm_id": "lead_001"},
		"settings": {
			"webhook": "https://your-app.com/webhooks/enrow",
			"country_code": "FR",
			"retrieve_gender": true,
			"retrieve_company_info": true
		}
	}`)
	if result.ID != "search_1" || result.Qualification != "ongoing" || result.CreditsUsed != 1 {
		t.Fatalf("result: %+v", result)
	}
	if n := len(api.calls("GET /email/find/single")); n != 0 {
		t.Fatalf("Find without PollOptions polled %d times", n)
	}
}

func TestEmailFindFromLinkedinURLWaitsForTheResult(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/find/single", 200, `{"message":"Single search operating","id":"search_2","credits_used":1}`)
	api.on("GET /email/find/single", 202, `{"qualification":"ongoing"}`)
	api.on("GET /email/find/single", 200, `{
		"email": "dwight.schrute@dundermifflin.com",
		"custom": {"crm_id": "lead_001"},
		"info": {"company_domain": "dundermifflin.com", "company_name": "Dunder Mifflin", "firstname": "Dwight", "lastname": "Schrute"},
		"linkedin_status": "found",
		"linkedin_url": "https://www.linkedin.com/in/dwight-schrute",
		"qualification": "valid",
		"credits_used": 1
	}`)

	result, err := client.Email.Find(EmailFindParams{
		LinkedinURL: "https://www.linkedin.com/in/dwight-schrute",
		Custom:      map[string]any{"crm_id": "lead_001"},
	}, fastPoll)
	if err != nil {
		t.Fatal(err)
	}

	assertJSON(t, api.only(t, "POST /email/find/single").body, `{
		"linkedin_url": "https://www.linkedin.com/in/dwight-schrute",
		"custom": {"crm_id": "lead_001"}
	}`)
	gets := api.calls("GET /email/find/single")
	if len(gets) != 2 {
		t.Fatalf("polled %d times, want 2", len(gets))
	}
	assertIDs(t, gets, "search_2")
	if result.ID != "search_2" || result.Qualification != "valid" ||
		result.Email != "dwight.schrute@dundermifflin.com" ||
		result.LinkedinStatus != "found" || result.LinkedinURL != "https://www.linkedin.com/in/dwight-schrute" ||
		result.Info == nil || result.Info.CompanyName != "Dunder Mifflin" {
		t.Fatalf("result: %+v (info %+v)", result, result.Info)
	}
}

func TestEmailFindWaitReturnsTheOngoingSearchOnTimeout(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/find/single", 200, `{"message":"Single search operating","id":"search_3","credits_used":1}`)
	api.on("GET /email/find/single", 202, `{"qualification":"ongoing"}`)

	start := time.Now()
	result, err := client.Email.Find(EmailFindParams{FullName: "Jim Halpert", CompanyDomain: "dundermifflin.com"},
		&PollOptions{WaitForResult: true, PollInterval: 5 * time.Millisecond, Timeout: 30 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < 30*time.Millisecond {
		t.Fatalf("returned after %v, before the timeout", elapsed)
	}
	if result.ID != "search_3" || result.Qualification != "ongoing" || result.CreditsUsed != 1 {
		t.Fatalf("result: %+v", result)
	}
	if n := len(api.calls("GET /email/find/single")); n < 2 {
		t.Fatalf("polled %d times", n)
	}
}

func TestEmailFindWaitKeepsTheIDWhenAGetFails(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/find/single", 200, `{"message":"Single search operating","id":"search_4","credits_used":1}`)
	api.on("GET /email/find/single", 500, `{"message":"Could not retrieve single search results"}`)

	result, err := client.Email.Find(EmailFindParams{FullName: "Jim Halpert", CompanyDomain: "dundermifflin.com"}, fastPoll)
	var apiErr *EnrowError
	if !errors.As(err, &apiErr) || apiErr.Status != 500 {
		t.Fatalf("err: %v", err)
	}
	if result == nil || result.ID != "search_4" {
		t.Fatalf("result: %+v", result)
	}
}

func TestEmailGet(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("GET /email/find/single", 200, `{
		"email": "dwight.schrute@dundermifflin.com",
		"custom": {"external_id": "do4eIF8jF40x!", "partition_number": 2},
		"info": {"company_domain": "dundermifflin.com", "fullname": "Dwight Schrute", "firstname": "dwight", "lastname": "schrute", "gender": "male", "companyInfo": {"siren": "123456789"}},
		"qualification": "valid",
		"credits_used": 1
	}`)

	result, err := client.Email.Get("search_5")
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, api.calls("GET /email/find/single"), "search_5")
	if result.ID != "search_5" || result.Email != "dwight.schrute@dundermifflin.com" || result.Qualification != "valid" ||
		result.Info == nil || result.Info.FullName != "Dwight Schrute" || result.Info.Gender != "male" ||
		result.Info.CompanyInfo["siren"] != "123456789" {
		t.Fatalf("result: %+v (info %+v)", result, result.Info)
	}
}

func TestEmailBulkUsesTheReturnedID(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/find/bulk", 201, `{"message":"Bulk search operating","id":"bulk_1","credits_used":2,"estimated_duration":1}`)
	api.on("GET /email/find/bulk", 200, `{
		"general": {"id": "bulk_1", "status": "completed", "custom": {"campaign": "q2"}},
		"stats": {"finished": 2, "requested": 2, "valid": 1, "credits_cost": {"initial": 2, "refunded": 1, "final": 1}},
		"results": [
			{"index": "0", "qualification": "valid", "email": "dwight.schrute@dundermifflin.com",
			 "info": {"domain": "dundermifflin.com", "firstname": "Dwight", "lastname": "Schrute", "companyInfo": {"name": "Dunder Mifflin"}},
			 "raw_params": {"company_domain": "dundermifflin.com", "fullname": "Dwight Schrute"}, "custom": "lead_001"},
			{"index": "1", "qualification": "invalid", "email": null,
			 "raw_params": {"company_name": "Dunder Mifflin", "fullname": "Ryan Howard"}, "custom": "lead_002"}
		]
	}`)

	batch, err := client.Email.FindBulk(EmailFindBulkParams{
		Searches: []EmailFindParams{
			{FullName: "Dwight Schrute", CompanyDomain: "dundermifflin.com", Custom: map[string]any{"lead": "001"}},
			{FullName: "Ryan Howard", CompanyName: "Dunder Mifflin"},
		},
		Custom:   map[string]any{"campaign": "q2"},
		Settings: &SearchSettings{Webhook: "https://your-app.com/webhooks/enrow", RetrieveGender: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, api.only(t, "POST /email/find/bulk").body, `{
		"searches": [
			{"fullname": "Dwight Schrute", "company_domain": "dundermifflin.com", "custom": {"lead": "001"}},
			{"fullname": "Ryan Howard", "company_name": "Dunder Mifflin"}
		],
		"custom": {"campaign": "q2"},
		"settings": {"webhook": "https://your-app.com/webhooks/enrow", "retrieve_gender": true}
	}`)
	if batch.ID != "bulk_1" || batch.CreditsUsed != 2 || batch.EstimatedDuration != 1 {
		t.Fatalf("batch: %+v", batch)
	}

	results, err := client.Email.GetBulk(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, api.calls("GET /email/find/bulk"), "bulk_1")
	if results.General.Status != "completed" || results.Stats.Valid != 1 ||
		results.Stats.CreditsCost.Final != 1 || len(results.Results) != 2 {
		t.Fatalf("results: %+v", results)
	}
	found := results.Results[0]
	if found.Index != "0" || found.Email != "dwight.schrute@dundermifflin.com" || found.Custom != "lead_001" ||
		found.Info == nil || found.Info.Domain != "dundermifflin.com" || found.Info.CompanyInfo["name"] != "Dunder Mifflin" {
		t.Fatalf("first result: %+v (info %+v)", found, found.Info)
	}
	if missed := results.Results[1]; missed.Qualification != "invalid" || missed.Email != "" || missed.Info != nil {
		t.Fatalf("second result: %+v", missed)
	}
}

// ── Email Verifier ──

func TestVerifySingleReturnsTheIDAndWaitsForTheVerdict(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/verify/single", 200, `{"message":"Single verification operating","id":"verif_1","credits_used":0.25,
		"credits":{"amount":0.25,"source":"sub","split":{"fromPaygo":0,"fromSub":0.25}}}`)
	api.on("GET /email/verify/single", 202, `{"qualification":"ongoing"}`)
	api.on("GET /email/verify/single", 200, `{"email":"pam.beesly@dundermifflin.com","custom":{"external_id":"do4eIF8jF40x!"},"qualification":"valid"}`)

	params := VerifySingleParams{
		Email:  "pam.beesly@dundermifflin.com",
		Custom: map[string]any{"external_id": "do4eIF8jF40x!"},
	}
	started, err := client.Verify.Single(params)
	if err != nil {
		t.Fatal(err)
	}
	if started.ID != "verif_1" || started.Qualification != "ongoing" || started.CreditsUsed != 0.25 {
		t.Fatalf("started: %+v", started)
	}

	result, err := client.Verify.Single(params, fastPoll)
	if err != nil {
		t.Fatal(err)
	}
	for _, req := range api.calls("POST /email/verify/single") {
		assertJSON(t, req.body, `{"email": "pam.beesly@dundermifflin.com", "custom": {"external_id": "do4eIF8jF40x!"}}`)
	}
	assertIDs(t, api.calls("GET /email/verify/single"), "verif_1")
	if result.ID != "verif_1" || result.Qualification != "valid" || result.CreditsUsed != 0.25 ||
		result.Email != "pam.beesly@dundermifflin.com" {
		t.Fatalf("result: %+v", result)
	}
}

func TestVerifyBulkSendsVerifications(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /email/verify/bulk", 201, `{"message":"Bulk verification operating","id":"bulk_2","credits_used":0.5,"estimated_duration":1}`)
	api.on("GET /email/verify/bulk", 200, `{
		"general": {"id": "bulk_2", "status": "completed"},
		"results": [
			{"email": "pam.beesly@dundermifflin.com", "index": "195f87c6-376e-4eca-9861-f92935413b0c", "qualification": "valid"},
			{"email": "angela.martin@dundermifflin.com", "index": "662784ae-3722-47af-b056-7554e013f73b", "qualification": "invalid"}
		],
		"stats": {"credits_cost": 0.5, "requested": 2, "valid": 1}
	}`)

	batch, err := client.Verify.Bulk(VerifyBulkParams{
		Emails:   []string{"pam.beesly@dundermifflin.com", "angela.martin@dundermifflin.com"},
		Custom:   map[string]any{"list": "newsletter"},
		Settings: &WebhookOnlySetting{Webhook: "https://your-app.com/webhooks/enrow/verify"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, api.only(t, "POST /email/verify/bulk").body, `{
		"verifications": ["pam.beesly@dundermifflin.com", "angela.martin@dundermifflin.com"],
		"custom": {"list": "newsletter"},
		"settings": {"webhook": "https://your-app.com/webhooks/enrow/verify"}
	}`)
	if batch.ID != "bulk_2" || batch.CreditsUsed != 0.5 {
		t.Fatalf("batch: %+v", batch)
	}

	results, err := client.Verify.GetBulk(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, api.calls("GET /email/verify/bulk"), "bulk_2")
	if results.General.ID != "bulk_2" || results.General.Status != "completed" || results.Stats.Valid != 1 ||
		results.Stats.CreditsCost != 0.5 || len(results.Results) != 2 ||
		results.Results[1].Email != "angela.martin@dundermifflin.com" || results.Results[1].Qualification != "invalid" {
		t.Fatalf("results: %+v", results)
	}
}

func TestVerifyGetBulkWhileOngoing(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("GET /email/verify/bulk", 200, `{"general":{"id":"bulk_2","status":"ongoing"},"stats":{"finished":1,"requested":2}}`)

	results, err := client.Verify.GetBulk("bulk_2")
	if err != nil {
		t.Fatal(err)
	}
	if results.General.Status != "ongoing" || results.Stats.Finished != 1 || results.Stats.Requested != 2 || len(results.Results) != 0 {
		t.Fatalf("results: %+v", results)
	}
}

// ── Phone Finder ──

func TestPhoneFindByNameSendsFirstnameAndLastname(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /phone/single", 201, `{"credits_used":40,"id":"phone_1","message":"Single search operating"}`)
	api.on("GET /phone/single", 202, `{"qualification":"ongoing"}`)
	api.on("GET /phone/single", 200, `{"params":{"firstname":"Michael","lastname":"Scott","company_domain":"dundermifflin.com"},
		"custom":"#u84hde941Jdx","number":"+15705551234","country":"US","qualification":"found"}`)

	result, err := client.Phone.Find(PhoneFindParams{
		FirstName:     "Michael",
		LastName:      "Scott",
		CompanyDomain: "dundermifflin.com",
		Custom:        "#u84hde941Jdx",
	}, fastPoll)
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, api.only(t, "POST /phone/single").body, `{
		"firstname": "Michael",
		"lastname": "Scott",
		"company_domain": "dundermifflin.com",
		"custom": "#u84hde941Jdx"
	}`)
	assertIDs(t, api.calls("GET /phone/single"), "phone_1")
	if result.ID != "phone_1" || result.Qualification != "found" || result.Number != "+15705551234" ||
		result.Country != "US" || result.Custom != "#u84hde941Jdx" {
		t.Fatalf("result: %+v", result)
	}
}

func TestPhoneFindWithoutWaiting(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /phone/single", 201, `{"credits_used":40,"id":"phone_2","message":"Single search operating"}`)

	result, err := client.Phone.Find(PhoneFindParams{LinkedinURL: "https://www.linkedin.com/in/michael-scott"})
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, api.only(t, "POST /phone/single").body, `{"linkedin_url": "https://www.linkedin.com/in/michael-scott"}`)
	if result.ID != "phone_2" || result.Qualification != "ongoing" {
		t.Fatalf("result: %+v", result)
	}
}

func TestPhoneGet(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("GET /phone/single", 200, `{"params":{"linkedin_url":"https://www.linkedin.com/in/michael-scott/"},
		"number":"+33612345678","country":"FR","qualification":"found"}`)

	result, err := client.Phone.Get("phone_123")
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, api.calls("GET /phone/single"), "phone_123")
	if result.ID != "phone_123" || result.Number != "+33612345678" || result.Country != "FR" {
		t.Fatalf("result: %+v", result)
	}
}

func TestPhoneBulkUsesTheReturnedID(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on("POST /phone/bulk", 201, `{"credits_used":80,"estimated_duration":1,"id":"bulk_3","message":"Bulk search operating",
		"credits":{"amount":80,"source":"sub","split":{"fromPaygo":0,"fromSub":80}}}`)
	api.on("GET /phone/bulk", 200, `{
		"general": {"id": "bulk_3", "status": "completed"},
		"results": [
			{"country": "US", "index": 0, "number": "+15705551234", "params": {"linkedin_url": "https://www.linkedin.com/in/michael-scott"}, "qualification": "found"},
			{"index": 1, "params": {"firstname": "Dwight", "lastname": "Schrute", "company_name": "Dunder Mifflin"}, "qualification": "not_found"}
		],
		"stats": {"credits_cost": 80, "credits_to_refund": 40, "found": 1, "requested": 2}
	}`)

	batch, err := client.Phone.FindBulk(PhoneFindBulkParams{
		Searches: []PhoneFindParams{
			{LinkedinURL: "https://www.linkedin.com/in/michael-scott"},
			{FirstName: "Dwight", LastName: "Schrute", CompanyName: "Dunder Mifflin"},
		},
		Settings: &WebhookOnlySetting{Webhook: "https://your-app.com/webhooks/enrow/phone"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, api.only(t, "POST /phone/bulk").body, `{
		"searches": [
			{"linkedin_url": "https://www.linkedin.com/in/michael-scott"},
			{"firstname": "Dwight", "lastname": "Schrute", "company_name": "Dunder Mifflin"}
		],
		"settings": {"webhook": "https://your-app.com/webhooks/enrow/phone"}
	}`)
	if batch.ID != "bulk_3" || batch.CreditsUsed != 80 {
		t.Fatalf("batch: %+v", batch)
	}

	results, err := client.Phone.GetBulk(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, api.calls("GET /phone/bulk"), "bulk_3")
	if results.General.Status != "completed" || results.Stats.Found != 1 || results.Stats.CreditsToRefund != 40 ||
		len(results.Results) != 2 {
		t.Fatalf("results: %+v", results)
	}
	if r := results.Results[0]; r.Index != 0 || r.Number != "+15705551234" ||
		r.Params["linkedin_url"] != "https://www.linkedin.com/in/michael-scott" {
		t.Fatalf("first result: %+v", r)
	}
	if r := results.Results[1]; r.Index != 1 || r.Qualification != "not_found" || r.Number != "" {
		t.Fatalf("second result: %+v", r)
	}
}

// ── Errors ──

func isError[T error](err error) bool {
	_, ok := err.(T)
	return ok
}

func TestErrorStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		is     func(error) bool
		want   string
	}{
		{"401", 401, `{"message":"This apikey is not valid"}`,
			isError[*AuthenticationError], "enrow: HTTP 401: This apikey is not valid"},
		{"402 on a single endpoint", 402, `{"reason":"Insufficient credits","success":false}`,
			isError[*InsufficientBalanceError], "enrow: HTTP 402: Insufficient credits"},
		{"402 on a bulk endpoint", 402, `{"message":"Insufficient credits"}`,
			isError[*InsufficientBalanceError], "enrow: HTTP 402: Insufficient credits"},
		{"422 is not a balance error", 422, `{"message":"Unprocessable"}`,
			isError[*EnrowError], "enrow: HTTP 422: Unprocessable"},
		{"429", 429, `{"message":"Too Many Requests"}`,
			isError[*RateLimitError], "enrow: HTTP 429: Too Many Requests"},
		{"400", 400, `{"message":"missing fullname"}`,
			isError[*EnrowError], "enrow: HTTP 400: missing fullname"},
		{"not JSON", 502, `Bad Gateway`,
			isError[*EnrowError], "enrow: HTTP 502: Bad Gateway"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api, client := newFakeAPI(t)
			api.on("POST /email/find/single", tc.status, tc.body)

			result, err := client.Email.Find(EmailFindParams{FullName: "Dwight Schrute", CompanyDomain: "dundermifflin.com"})
			if result != nil || err == nil || !tc.is(err) {
				t.Fatalf("got %T %v (result %+v)", err, err, result)
			}
			if err.Error() != tc.want {
				t.Fatalf("message: got %q, want %q", err.Error(), tc.want)
			}
		})
	}
}
