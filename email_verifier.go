package enrow

type EmailVerifier struct {
	client *Client
}

// Single launches a single email verification. Without PollOptions it returns
// at once with the verification ID and Qualification "ongoing"; with
// WaitForResult it polls Get until the verification has ended (see
// PollOptions).
func (v *EmailVerifier) Single(params VerifySingleParams, opts ...*PollOptions) (*VerifySingleResult, error) {
	var started VerifySingleResult
	if err := v.client.doPost("/email/verify/single", params, &started); err != nil {
		return nil, err
	}
	if started.Qualification == "" {
		started.Qualification = ongoing
	}

	o := pollOptions(opts)
	if !o.WaitForResult {
		return &started, nil
	}

	result, err := poll(
		func() (*VerifySingleResult, error) { return v.Get(started.ID) },
		func(r *VerifySingleResult) bool { return r.Qualification != ongoing },
		o.PollInterval, o.Timeout,
	)
	if err != nil {
		return &started, err
	}
	result.CreditsUsed = started.CreditsUsed
	return result, nil
}

// Get reads the state of a single email verification.
func (v *EmailVerifier) Get(id string) (*VerifySingleResult, error) {
	var result VerifySingleResult
	if err := v.client.doGet("/email/verify/single", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	// The GET response does not repeat the verification ID.
	result.ID = id
	return &result, nil
}

// Bulk launches a bulk email verification. Pass the returned ID to GetBulk.
func (v *EmailVerifier) Bulk(params VerifyBulkParams) (*VerifyBulkResponse, error) {
	var result VerifyBulkResponse
	if err := v.client.doPost("/email/verify/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBulk reads the state and, once completed, the results of a bulk email
// verification.
func (v *EmailVerifier) GetBulk(id string) (*VerifyBulkResult, error) {
	var result VerifyBulkResult
	if err := v.client.doGet("/email/verify/bulk", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
