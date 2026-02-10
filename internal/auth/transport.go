package auth

import (
	"fmt"
	"net/http"
)

// Transport is an http.RoundTripper that injects Authorization and X-AP-Context headers.
type Transport struct {
	Base     http.RoundTripper
	Token    *TokenProvider
	OrgID    string
	Verbose  bool
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.Token.GetToken()
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+token)
	if t.OrgID != "" {
		req2.Header.Set("X-AP-Context", "orgId="+t.OrgID)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	if t.Verbose {
		fmt.Printf("> %s %s\n", req2.Method, req2.URL)
		for k, v := range req2.Header {
			switch k {
			case "Authorization":
				fmt.Printf("> %s: Bearer ***\n", k)
			case "X-Ap-Context":
				fmt.Printf("> %s: orgId=***\n", k)
			default:
				fmt.Printf("> %s: %s\n", k, v)
			}
		}
	}

	resp, err := base.RoundTrip(req2)
	if err != nil {
		return nil, err
	}

	if t.Verbose {
		fmt.Printf("< %s %s\n", resp.Status, resp.Proto)
	}

	return resp, nil
}
