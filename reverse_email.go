package enrow

type ReverseEmailResource struct {
	client *Client
}

func (r *ReverseEmailResource) Find(params ReverseEmailParams, opts ...*PollOptions) (*ReverseEmailResult, error) {
	var result ReverseEmailResult
	if err := r.client.doPost("/reverse-email/single", params, &result); err != nil {
		return nil, err
	}

	o := defaultPoll(nil)
	if len(opts) > 0 && opts[0] != nil {
		o = opts[0]
	}

	if o.WaitForResult && (result.Status == "pending" || result.Status == "processing") {
		return poll(
			func() (*ReverseEmailResult, error) { return r.Get(result.ID) },
			func(r *ReverseEmailResult) bool { return r.Status != "pending" && r.Status != "processing" },
			o.PollInterval, o.Timeout,
		)
	}

	return &result, nil
}

func (r *ReverseEmailResource) Get(id string) (*ReverseEmailResult, error) {
	var result ReverseEmailResult
	if err := r.client.doGet(formatPath("/reverse-email/single/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ReverseEmailResource) FindBulk(params ReverseEmailBulkParams) (*ReverseEmailBulkResponse, error) {
	var result ReverseEmailBulkResponse
	if err := r.client.doPost("/reverse-email/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ReverseEmailResource) GetBulk(id string) (*ReverseEmailBulkResult, error) {
	var result ReverseEmailBulkResult
	if err := r.client.doGet(formatPath("/reverse-email/bulk/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
