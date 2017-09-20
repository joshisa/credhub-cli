// CredHub credential value types
package values

type Value string

type JSON map[string]interface{}

type Password string

type User struct {
	Username *string `json:"username,omitempty"`
	Password string `json:"password"`
}

type Certificate struct {
	Ca          string `json:"ca,omitempty"`
	CaName      string `json:"ca_name,omitempty" yaml:"ca_name,omitempty"`
	Certificate string `json:"certificate,omitempty"`
	PrivateKey  string `json:"private_key,omitempty" yaml:"private_key,omitempty"`
}

type RSA struct {
	PublicKey  string `json:"public_key" yaml:"public_key"`
	PrivateKey string `json:"private_key" yaml:"private_key"`
}

type SSH struct {
	PublicKey  string `json:"public_key" yaml:"public_key"`
	PrivateKey string `json:"private_key" yaml:"private_key"`
}
