package vault

type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

type InitResponse struct {
	KeysB64   []string `json:"keys_base64"`
	RootToken string   `json:"root_token"`
}

type UnsealOpts struct {
	Key string `json:"key"`
}

type MountInput struct {
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}

type Policy struct {
	Capabilities []string `json:"capabilities"`
}

type TokenCreateRequest struct {
	ID          string   `json:"id,omitempty"`
	Policies    []string `json:"policies,omitempty"`
	DisplayName string   `json:"display_name"`
}
