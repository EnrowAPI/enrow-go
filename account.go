package enrow

type AccountResource struct {
	client *Client
}

func (a *AccountResource) Info() (*AccountInfo, error) {
	var result AccountInfo
	if err := a.client.doGet("/account/info", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
