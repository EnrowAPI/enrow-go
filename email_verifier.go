package enrow

type EmailVerifier struct {
	client *Client
}

func (v *EmailVerifier) Single(params VerifySingleParams) (*VerifySingleResult, error) {
	var result VerifySingleResult
	if err := v.client.doPost("/verify/single", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) Get(id string) (*VerifySingleResult, error) {
	var result VerifySingleResult
	if err := v.client.doGet(formatPath("/verify/single/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) Bulk(params VerifyBulkParams) (*VerifyBulkResponse, error) {
	var result VerifyBulkResponse
	if err := v.client.doPost("/verify/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) GetBulk(id string) (*VerifyBulkResult, error) {
	var result VerifyBulkResult
	if err := v.client.doGet(formatPath("/verify/bulk/%s", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
