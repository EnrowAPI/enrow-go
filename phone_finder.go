package enrow

type PhoneFinder struct {
	client *Client
}

func (p *PhoneFinder) Find(params PhoneFindParams, opts ...*PollOptions) (*PhoneFindResult, error) {
	var resp PhoneFindResponse
	if err := p.client.doPost("/phone/single", params, &resp); err != nil {
		return nil, err
	}

	o := defaultPoll(nil)
	if len(opts) > 0 && opts[0] != nil {
		o = opts[0]
	}

	if o.WaitForResult {
		return poll(
			func() (*PhoneFindResult, error) { return p.Get(resp.ID) },
			func(r *PhoneFindResult) bool { return r.Qualification != "ongoing" },
			o.PollInterval, o.Timeout,
		)
	}

	// Return minimal result from POST response
	return &PhoneFindResult{ID: resp.ID, Qualification: "ongoing"}, nil
}

func (p *PhoneFinder) Get(id string) (*PhoneFindResult, error) {
	var result PhoneFindResult
	if err := p.client.doGet("/phone/single", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *PhoneFinder) FindBulk(params PhoneFindBulkParams) (*PhoneBulkResponse, error) {
	var result PhoneBulkResponse
	if err := p.client.doPost("/phone/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *PhoneFinder) GetBulk(id string) (*PhoneBulkResult, error) {
	var result PhoneBulkResult
	if err := p.client.doGet("/phone/bulk", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
