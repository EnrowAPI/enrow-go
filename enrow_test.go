package enrow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := New("test_key", WithBaseURL(server.URL))
	return server, client
}

func TestEmailFind(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test_key" {
			t.Fatal("missing api key")
		}
		if r.Method != "POST" || r.URL.Path != "/email/find/single" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["company_domain"] != "apple.com" {
			t.Fatalf("expected apple.com, got %v", body["company_domain"])
		}
		if body["fullname"] != "Tim Cook" {
			t.Fatalf("expected fullname=Tim Cook, got %v", body["fullname"])
		}

		json.NewEncoder(w).Encode(EmailFindResult{
			ID:            "search_123",
			Status:        "completed",
			Email:         "tcook@apple.com",
			Qualification: "valid",
		})
	})
	defer server.Close()

	result, err := client.Email.Find(EmailFindParams{
		CompanyDomain: "apple.com",
		FullName:      "Tim Cook",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Email != "tcook@apple.com" {
		t.Fatalf("expected tcook@apple.com, got %s", result.Email)
	}
}

func TestEmailGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/email/find/single" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(EmailFindResult{
			ID:    "search_123",
			Email: "tcook@apple.com",
		})
	})
	defer server.Close()

	result, err := client.Email.Get("search_123")
	if err != nil {
		t.Fatal(err)
	}
	if result.Email != "tcook@apple.com" {
		t.Fatalf("expected tcook@apple.com, got %s", result.Email)
	}
}

func TestAuthenticationError(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Unauthorized", "message": "Invalid API key",
		})
	})
	defer server.Close()

	_, err := client.Email.Find(EmailFindParams{CompanyDomain: "test.com"})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}
}

func TestRateLimitError(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "RateLimitExceeded", "message": "Rate limit exceeded",
		})
	})
	defer server.Close()

	_, err := client.Email.Find(EmailFindParams{CompanyDomain: "test.com"})
	if _, ok := err.(*RateLimitError); !ok {
		t.Fatalf("expected RateLimitError, got %T", err)
	}
}

func TestInsufficientBalanceError(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "InsufficientBalance", "message": "Not enough credits",
		})
	})
	defer server.Close()

	_, err := client.Email.Find(EmailFindParams{CompanyDomain: "test.com"})
	if _, ok := err.(*InsufficientBalanceError); !ok {
		t.Fatalf("expected InsufficientBalanceError, got %T", err)
	}
}

func TestPhoneGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "phone_123" {
			t.Fatalf("expected query param id=phone_123, got %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(PhoneFindResult{
			ID:            "phone_123",
			Qualification: "found",
			Number:        "+33612345678",
			Country:       "FR",
		})
	})
	defer server.Close()

	result, err := client.Phone.Get("phone_123")
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != "+33612345678" {
		t.Fatalf("expected +33612345678, got %s", result.Number)
	}
}
