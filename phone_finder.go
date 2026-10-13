package enrow

type PhoneFinder struct {
	client *Client
}

// Find launches a single phone search. Without PollOptions it returns at once
// with the search ID and Qualification "ongoing"; with WaitForResult it polls
// Get until the search has ended (see PollOptions).
func (p *PhoneFinder) Find(params PhoneFindParams, opts ...*PollOptions) (*PhoneFindResult, error) {
	var resp PhoneFindResponse
	if err := p.client.doPost("/phone/single", params, &resp); err != nil {
		return nil, err
	}
	started := &PhoneFindResult{ID: resp.ID, Qualification: ongoing}

	o := pollOptions(opts)
	if !o.WaitForResult {
		return started, nil
	}

	result, err := poll(
		func() (*PhoneFindResult, error) { return p.Get(resp.ID) },
		func(r *PhoneFindResult) bool { return r.Qualification != ongoing },
		o.PollInterval, o.Timeout,
	)
	if err != nil {
		return started, err
	}
	return result, nil
}

// Get reads the state of a single phone search.
func (p *PhoneFinder) Get(id string) (*PhoneFindResult, error) {
	var result PhoneFindResult
	if err := p.client.doGet("/phone/single", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	// The GET response does not repeat the search ID.
	result.ID = id
	return &result, nil
}

// FindBulk launches a bulk phone search. Pass the returned ID to GetBulk.
func (p *PhoneFinder) FindBulk(params PhoneFindBulkParams) (*PhoneBulkResponse, error) {
	var result PhoneBulkResponse
	if err := p.client.doPost("/phone/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBulk reads the state and, once completed, the results of a bulk phone
// search.
func (p *PhoneFinder) GetBulk(id string) (*PhoneBulkResult, error) {
	var result PhoneBulkResult
	if err := p.client.doGet("/phone/bulk", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
