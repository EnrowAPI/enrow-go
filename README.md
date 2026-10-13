# Enrow Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/EnrowAPI/enrow-go)](https://github.com/EnrowAPI/enrow-go)
[![Last commit](https://img.shields.io/github/last-commit/EnrowAPI/enrow-go)](https://github.com/EnrowAPI/enrow-go/commits)

Find and verify professional emails and find phone numbers with the [Enrow API](https://enrow.io).

## Install

The SDK is in early access and has no tagged release yet: install it from `main`.

```bash
go get github.com/EnrowAPI/enrow-go@main
```

Requires Go 1.21 or later.

## Quick start

```go
package main

import (
	"fmt"

	enrow "github.com/EnrowAPI/enrow-go"
)

func main() {
	client := enrow.New("YOUR_API_KEY")

	result, err := client.Email.Find(enrow.EmailFindParams{
		FullName:      "Dwight Schrute",
		CompanyDomain: "dundermifflin.com",
	}, &enrow.PollOptions{WaitForResult: true})
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Qualification, result.Email) // valid dwight.schrute@dundermifflin.com
}
```

## Searches are asynchronous

Every search and verification starts with a POST that answers an `id`, and its result is read later with a GET (or received on a [webhook](https://docs.enrow.io/how-webhooks-work)).

- Without `PollOptions`, `Find` and `Verify.Single` return right away with the `ID` and `Qualification: "ongoing"`. Read the result later with `Get(ID)`.
- With `&enrow.PollOptions{WaitForResult: true}`, they call `Get` every `PollInterval` (default 2s) until the search has ended or `Timeout` (default 30s) has elapsed. On timeout, the result comes back as it is, with `Qualification: "ongoing"` and its `ID`: call `Get` again later.
- Bulk operations answer an `ID` too: pass it to `GetBulk`, or set a webhook. `General.Status` is `ongoing`, `completed` or `failed`, and `Results` is filled once the batch is `completed`.

## Email Finder

```go
// Launch a search by name and company, and read it later
search, err := client.Email.Find(enrow.EmailFindParams{
	FullName:    "Dwight Schrute",
	CompanyName: "Dunder Mifflin",
	Custom:      map[string]any{"crm_id": "lead_001"},
	Settings: &enrow.SearchSettings{
		CountryCode:    "US",
		RetrieveGender: true,
		Webhook:        "https://your-app.com/webhooks/enrow",
	},
})
result, err := client.Email.Get(search.ID)
fmt.Println(result.Qualification) // valid, invalid or ongoing

// Search from a LinkedIn profile URL (single searches only)
fromProfile, err := client.Email.Find(enrow.EmailFindParams{
	LinkedinURL: "https://www.linkedin.com/in/dwight-schrute",
}, &enrow.PollOptions{WaitForResult: true})
fmt.Println(fromProfile.Email, fromProfile.LinkedinStatus) // found, profile_not_found, no_current_company, email_not_found or lookup_failed

// Company information (only effective when CountryCode is "FR")
withCompany, err := client.Email.Find(enrow.EmailFindParams{
	FullName:    "Jean Dupont",
	CompanyName: "Acme France",
	Settings:    &enrow.SearchSettings{CountryCode: "FR", RetrieveCompanyInfo: true},
}, &enrow.PollOptions{WaitForResult: true})
if withCompany.Info != nil {
	fmt.Println(withCompany.Info.CompanyInfo)
}

// Bulk: up to 5,000 searches per batch
batch, err := client.Email.FindBulk(enrow.EmailFindBulkParams{
	Searches: []enrow.EmailFindParams{
		{FullName: "Dwight Schrute", CompanyDomain: "dundermifflin.com"},
		{FullName: "Jim Halpert", CompanyDomain: "dundermifflin.com"},
	},
	Settings: &enrow.SearchSettings{Webhook: "https://your-app.com/webhooks/enrow"},
})
results, err := client.Email.GetBulk(batch.ID)
if results.General.Status == "completed" {
	for _, r := range results.Results {
		fmt.Println(r.Index, r.Qualification, r.Email)
	}
}
```

## Email Verifier

```go
// Verify one address and wait for the verdict
verification, err := client.Verify.Single(enrow.VerifySingleParams{
	Email: "pam.beesly@dundermifflin.com",
}, &enrow.PollOptions{WaitForResult: true})
fmt.Println(verification.Qualification) // valid or invalid

// Bulk: up to 5,000 addresses per batch
batch, err := client.Verify.Bulk(enrow.VerifyBulkParams{
	Emails: []string{"pam.beesly@dundermifflin.com", "angela.martin@dundermifflin.com"},
})
results, err := client.Verify.GetBulk(batch.ID)
fmt.Println(results.General.Status, results.Stats.Valid, len(results.Results))
```

## Phone Finder

```go
// From a LinkedIn profile URL (recommended: it prevails over the name and company)
phone, err := client.Phone.Find(enrow.PhoneFindParams{
	LinkedinURL: "https://www.linkedin.com/in/michael-scott",
}, &enrow.PollOptions{WaitForResult: true})
if phone.Qualification == "found" {
	fmt.Println(phone.Number, phone.Country)
}

// From a first name, a last name and a company
byName, err := client.Phone.Find(enrow.PhoneFindParams{
	FirstName:     "Michael",
	LastName:      "Scott",
	CompanyDomain: "dundermifflin.com",
})
fmt.Println(byName.ID) // read it later with client.Phone.Get(byName.ID)

// Bulk: up to 3,000 searches per batch
batch, err := client.Phone.FindBulk(enrow.PhoneFindBulkParams{
	Searches: []enrow.PhoneFindParams{
		{LinkedinURL: "https://www.linkedin.com/in/michael-scott"},
		{LinkedinURL: "https://www.linkedin.com/in/dwight-schrute"},
	},
})
results, err := client.Phone.GetBulk(batch.ID)
fmt.Println(results.General.Status, results.Stats.Found)
```

## Account

```go
info, err := client.Account.Info()
fmt.Println(info.Credits)
```

## Error handling

```go
_, err := client.Email.Find(enrow.EmailFindParams{FullName: "Dwight Schrute", CompanyDomain: "dundermifflin.com"})
if err != nil {
	switch e := err.(type) {
	case *enrow.AuthenticationError:
		// 401: missing or invalid API key, or a feature not enabled on the account
	case *enrow.InsufficientBalanceError:
		// 402: not enough credits
	case *enrow.RateLimitError:
		// 429: over 10 requests per second on POST endpoints, retry with backoff
	case *enrow.EnrowError:
		// any other API error, such as a 400 on invalid input
		fmt.Println(e.Status, e.Message)
	}
}
```

## Credits

| Endpoint | Cost |
|----------|------|
| Email Finder | 1 credit per email found |
| Email Verifier | 0.25 credit per verification |
| Phone Finder | 40 credits per phone found (50 on legacy plans) |

A search from a LinkedIn profile URL costs 0.25 credit when no email is found. See [Credits & billing](https://docs.enrow.io/credits-billing).

## Links

- [API Documentation](https://docs.enrow.io)
- [Enrow](https://enrow.io)

## License

MIT
