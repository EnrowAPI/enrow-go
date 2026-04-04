package enrow

type EmailVerifier struct {
	client *Client
}

func (v *EmailVerifier) Single(params VerifySingleParams) (*VerifySingleResult, error) {
	var result VerifySingleResult
	if err := v.client.doPost("/email/verify/single", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) Get(id string) (*VerifySingleResult, error) {
	var result VerifySingleResult
	if err := v.client.doGet("/email/verify/single", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) Bulk(params VerifyBulkParams) (*VerifyBulkResponse, error) {
	var result VerifyBulkResponse
	if err := v.client.doPost("/email/verify/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *EmailVerifier) GetBulk(id string) (*VerifyBulkResult, error) {
	var result VerifyBulkResult
	if err := v.client.doGet("/email/verify/bulk", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
