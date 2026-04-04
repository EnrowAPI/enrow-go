package enrow

type EmailFinder struct {
	client *Client
}

func (e *EmailFinder) Find(params EmailFindParams, opts ...*PollOptions) (*EmailFindResult, error) {
	var result EmailFindResult
	if err := e.client.doPost("/email/find/single", params, &result); err != nil {
		return nil, err
	}

	o := defaultPoll(nil)
	if len(opts) > 0 && opts[0] != nil {
		o = opts[0]
	}

	if o.WaitForResult && result.Status == "pending" {
		return poll(
			func() (*EmailFindResult, error) { return e.Get(result.ID) },
			func(r *EmailFindResult) bool { return r.Status != "pending" && r.Status != "processing" },
			o.PollInterval, o.Timeout,
		)
	}

	return &result, nil
}

func (e *EmailFinder) Get(id string) (*EmailFindResult, error) {
	var result EmailFindResult
	if err := e.client.doGet(formatPath("/email/find/single/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *EmailFinder) FindBulk(params EmailFindBulkParams) (*EmailFindBulkResponse, error) {
	var result EmailFindBulkResponse
	if err := e.client.doPost("/email/find/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *EmailFinder) GetBulk(id string) (*EmailFindBulkResult, error) {
	var result EmailFindBulkResult
	if err := e.client.doGet(formatPath("/email/find/bulk/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
