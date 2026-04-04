# Enrow Go SDK

Find and verify professional emails, phone numbers, and contact information with the [Enrow API](https://enrow.io).

## Install

```bash
go get github.com/enrow/enrow-go
```

## Quick start

```go
package main

import (
    "fmt"
    enrow "github.com/enrow/enrow-go"
)

func main() {
    client := enrow.New("your_api_key")

    result, err := client.Email.Find(enrow.EmailFindParams{
        CompanyDomain: "apple.com",
        FullName:      "Tim Cook",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Email) // tcook@apple.com
}
```

## Email Finder

```go
// Find a single email
result, _ := client.Email.Find(enrow.EmailFindParams{
    CompanyDomain: "apple.com",
    FullName:      "Tim Cook",
})

// Wait for result (auto-polling)
result, _ := client.Email.Find(enrow.EmailFindParams{
    CompanyDomain: "apple.com",
    FullName:      "Tim Cook",
}, &enrow.PollOptions{WaitForResult: true})

// Get by ID
result, _ := client.Email.Get("search_abc123")

// Bulk
batch, _ := client.Email.FindBulk(enrow.EmailFindBulkParams{
    Searches: []enrow.EmailFindParams{
        {CompanyDomain: "apple.com", FullName: "Tim Cook"},
        {CompanyDomain: "microsoft.com", FullName: "Satya Nadella"},
    },
})
results, _ := client.Email.GetBulk(batch.BatchID)
```

## Email Verifier

```go
result, _ := client.Verify.Single(enrow.VerifySingleParams{Email: "tcook@apple.com"})
fmt.Println(result.Qualification) // "valid"

batch, _ := client.Verify.Bulk(enrow.VerifyBulkParams{Emails: []string{"a@b.com", "c@d.com"}})
results, _ := client.Verify.GetBulk(batch.BatchID)
```

## Phone Finder

```go
// By LinkedIn URL
result, _ := client.Phone.Find(enrow.PhoneFindParams{
    LinkedinURL: "https://linkedin.com/in/timcook",
}, &enrow.PollOptions{WaitForResult: true})

// By name + company
result, _ := client.Phone.Find(enrow.PhoneFindParams{
    FirstName:     "Tim",
    LastName:      "Cook",
    CompanyDomain: "apple.com",
}, &enrow.PollOptions{WaitForResult: true})
```

## Reverse Email

```go
result, _ := client.ReverseEmail.Find(enrow.ReverseEmailParams{Email: "tcook@apple.com"})
fmt.Println(result.FirstName, result.LinkedinURL)
```

## Account

```go
info, _ := client.Account.Info()
fmt.Println(info.Credits)
```

## Error handling

```go
result, err := client.Email.Find(enrow.EmailFindParams{CompanyDomain: "apple.com"})
if err != nil {
    switch err.(type) {
    case *enrow.RateLimitError:
        // 429
    case *enrow.InsufficientBalanceError:
        // 422
    case *enrow.AuthenticationError:
        // 401
    case *enrow.EnrowError:
        // other API error
    }
}
```

## Credits

| Endpoint | Cost |
|----------|------|
| Email Finder | 1 credit/email |
| Email Verifier | 0.25 credit/email |
| Phone Finder | 50 credits/phone |
| Reverse Email | 5 credits/lookup |

## Links

- [API Documentation](https://docs.enrow.io)
- [Enrow](https://enrow.io)

## License

MIT
