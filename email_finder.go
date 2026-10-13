package enrow

type EmailFinder struct {
	client *Client
}

// Find launches a single email search. Without PollOptions it returns at once
// with the search ID and Qualification "ongoing"; with WaitForResult it polls
// Get until the search has ended (see PollOptions).
func (e *EmailFinder) Find(params EmailFindParams, opts ...*PollOptions) (*EmailFindResult, error) {
	var started EmailFindResult
	if err := e.client.doPost("/email/find/single", params, &started); err != nil {
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
		func() (*EmailFindResult, error) { return e.Get(started.ID) },
		func(r *EmailFindResult) bool { return r.Qualification != ongoing },
		o.PollInterval, o.Timeout,
	)
	if err != nil {
		return &started, err
	}
	if result.Qualification == ongoing {
		// Timed out: an ongoing GET answers no cost, keep the launch cost.
		result.CreditsUsed = started.CreditsUsed
	}
	return result, nil
}

// Get reads the state of a single email search.
func (e *EmailFinder) Get(id string) (*EmailFindResult, error) {
	var result EmailFindResult
	if err := e.client.doGet("/email/find/single", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	// The GET response does not repeat the search ID.
	result.ID = id
	return &result, nil
}

// FindBulk launches a bulk email search. Pass the returned ID to GetBulk.
func (e *EmailFinder) FindBulk(params EmailFindBulkParams) (*EmailFindBulkResponse, error) {
	var result EmailFindBulkResponse
	if err := e.client.doPost("/email/find/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBulk reads the state and, once completed, the results of a bulk email
// search.
func (e *EmailFinder) GetBulk(id string) (*EmailFindBulkResult, error) {
	var result EmailFindBulkResult
	if err := e.client.doGet("/email/find/bulk", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
