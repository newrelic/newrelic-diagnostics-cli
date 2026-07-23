package agentcontrol

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"runtime"

	"gopkg.in/yaml.v3"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const (
	defaultOpAMPEndpoint = "https://opamp.service.newrelic.com/v1/opamp"
	defaultJWKSEndpoint  = "https://publickeys.newrelic.com/r/blob-management/global/agentconfiguration/jwks.json"
	defaultOAuthEndpoint = "https://system-identity-oauth.service.newrelic.com/oauth2/token"
	defaultStatusHost    = "127.0.0.1"
	defaultStatusPort    = 51200
)

var acLocalConfigPath = func() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files\New Relic\newrelic-agent-control\local-data\agent-control\local_config.yaml`
	}
	return "/etc/newrelic-agent-control/local-data/agent-control/local_config.yaml"
}()

// acConfig mirrors only the fields of local_config.yaml that we need.
type acConfig struct {
	FleetControl struct {
		Endpoint            string `yaml:"endpoint"`
		SignatureValidation struct {
			PublicKeyServerURL string `yaml:"public_key_server_url"`
		} `yaml:"signature_validation"`
		AuthConfig struct {
			TokenURL       string `yaml:"token_url"`
			ClientID       string `yaml:"client_id"`
			PrivateKeyPath string `yaml:"private_key_path"`
		} `yaml:"auth_config"`
	} `yaml:"fleet_control"`
	Server struct {
		Enabled bool   `yaml:"enabled"`
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
	} `yaml:"server"`
}

// readACConfig parses the agent-control local config file.
// Returns a zero-value acConfig and a non-nil error if the file is missing or unparseable.
func readACConfig(reader func(string) ([]byte, error)) (acConfig, error) {
	data, err := reader(acLocalConfigPath)
	if err != nil {
		return acConfig{}, err
	}
	var cfg acConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return acConfig{}, err
	}
	return cfg, nil
}

// ConnectResult holds the outcome of a single endpoint connectivity check.
type ConnectResult struct {
	URL        string
	Reachable  bool
	StatusCode int
	Note       string
}

// KeyValidationResult holds the outcome of the local private key file check.
type KeyValidationResult struct {
	Path  string
	Valid bool
	Note  string
}

// AgentControlConnectPayload is the result payload for AgentControl/Agent/Connect.
type AgentControlConnectPayload struct {
	Endpoints  []ConnectResult
	PrivateKey KeyValidationResult
}

// AgentControlAgentConnect checks network connectivity to agent-control service endpoints
// and validates the local identity private key.
type AgentControlAgentConnect struct {
	httpGetter   requestFunc
	configReader func(string) ([]byte, error)
}

func (p AgentControlAgentConnect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("AgentControl/Agent/Connect")
}

func (p AgentControlAgentConnect) Explain() string {
	return "Check network connectivity to New Relic agent-control service endpoints and validate the local identity key"
}

func (p AgentControlAgentConnect) Dependencies() []string {
	return []string{"AgentControl/Config/Agent"}
}

func (p AgentControlAgentConnect) Execute(_ tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["AgentControl/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Agent Control not detected on system.",
		}
	}

	cfg, cfgErr := readACConfig(p.configReader)
	if cfgErr != nil {
		log.Debug("Could not read agent-control config, using default endpoints: " + cfgErr.Error())
	}

	opampURL, jwksURL, oauthURL := endpointsFromConfig(cfg)

	opampResult := p.checkConnectivity("OpAMP", opampURL)
	jwksResult := p.checkJWKS(jwksURL)
	// OAuth2 token endpoint is POST-only and requires signed credentials that may be
	// expired. We only verify TCP/TLS reachability via a GET (any HTTP response = ok).
	oauthResult := p.checkConnectivity("OAuth2 token", oauthURL)
	keyResult := p.checkPrivateKey(cfg, cfgErr)

	payload := AgentControlConnectPayload{
		Endpoints:  []ConnectResult{opampResult, jwksResult, oauthResult},
		PrivateKey: keyResult,
	}

	var failures []string
	if !opampResult.Reachable {
		failures = append(failures, fmt.Sprintf("OpAMP (%s): %s", opampURL, opampResult.Note))
	}
	if !jwksResult.Reachable {
		failures = append(failures, fmt.Sprintf("JWKS (%s): %s", jwksURL, jwksResult.Note))
	}
	if !oauthResult.Reachable {
		failures = append(failures, fmt.Sprintf("OAuth2 token (%s): %s", oauthURL, oauthResult.Note))
	}
	if !keyResult.Valid {
		failures = append(failures, fmt.Sprintf("identity key (%s): %s", keyResult.Path, keyResult.Note))
	}

	if len(failures) > 0 {
		summary := "Agent Control connectivity/configuration check failed:\n"
		for _, f := range failures {
			summary += "  - " + f + "\n"
		}
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: summary,
			URL:     "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks/",
			Payload: payload,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("All agent-control endpoints reachable and identity key valid. JWKS: %s", jwksResult.Note),
		Payload: payload,
	}
}

func endpointsFromConfig(cfg acConfig) (opampURL, jwksURL, oauthURL string) {
	opampURL = defaultOpAMPEndpoint
	jwksURL = defaultJWKSEndpoint
	oauthURL = defaultOAuthEndpoint

	if cfg.FleetControl.Endpoint != "" {
		opampURL = cfg.FleetControl.Endpoint
	}
	if cfg.FleetControl.SignatureValidation.PublicKeyServerURL != "" {
		jwksURL = cfg.FleetControl.SignatureValidation.PublicKeyServerURL
	}
	if cfg.FleetControl.AuthConfig.TokenURL != "" {
		oauthURL = cfg.FleetControl.AuthConfig.TokenURL
	}
	return
}

// checkPrivateKey validates the local identity private key referenced in the config:
// - file exists and is readable
// - contains a valid PKCS8 PEM block
// - decodes as an Ed25519 private key
//
// It does NOT attempt any network call or token retrieval, because the client_id
// validity is managed backend-side and may be expired independently of the key.
func (p AgentControlAgentConnect) checkPrivateKey(cfg acConfig, cfgErr error) KeyValidationResult {
	if cfgErr != nil {
		return KeyValidationResult{
			Valid: false,
			Note:  "config unavailable, cannot validate identity key: " + cfgErr.Error(),
		}
	}

	path := cfg.FleetControl.AuthConfig.PrivateKeyPath
	if path == "" {
		return KeyValidationResult{
			Valid: false,
			Note:  "private_key_path not set in config",
		}
	}

	data, err := p.configReader(path)
	if err != nil {
		return KeyValidationResult{
			Path:  path,
			Valid: false,
			Note:  "cannot read key file: " + err.Error(),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return KeyValidationResult{
			Path:  path,
			Valid: false,
			Note:  "file does not contain a valid PEM block",
		}
	}

	// agent-control expects PKCS8 ("PRIVATE KEY"), not OpenSSH or raw formats.
	if block.Type != "PRIVATE KEY" {
		return KeyValidationResult{
			Path:  path,
			Valid: false,
			Note:  fmt.Sprintf("expected PKCS8 PEM type \"PRIVATE KEY\", got %q", block.Type),
		}
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return KeyValidationResult{
			Path:  path,
			Valid: false,
			Note:  "cannot parse PKCS8 key: " + err.Error(),
		}
	}

	var keyType string
	switch key.(type) {
	case *rsa.PrivateKey:
		return KeyValidationResult{
			Path:  path,
			Valid: true,
			Note:  "valid RSA PKCS8 private key (RS256)",
		}
	case ed25519.PrivateKey:
		keyType = "Ed25519"
	case *ecdsa.PrivateKey:
		keyType = "ECDSA"
	default:
		keyType = fmt.Sprintf("%T", key)
	}
	return KeyValidationResult{
		Path:  path,
		Valid: false,
		Note:  fmt.Sprintf("%s key is not supported; agent-control requires RSA (RS256)", keyType),
	}
}

// checkConnectivity makes a GET request and considers any HTTP response (including 4xx/5xx)
// as "reachable" — only a connection-level error means the endpoint is unreachable.
func (p AgentControlAgentConnect) checkConnectivity(name, url string) ConnectResult {
	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            url,
		TimeoutSeconds: 30,
	}

	response, err := p.httpGetter(wrapper)
	if err != nil {
		log.Debug(fmt.Sprintf("%s endpoint unreachable: %s", name, err.Error()))
		return ConnectResult{URL: url, Reachable: false, Note: err.Error()}
	}
	_ = response.Body.Close()

	log.Debug(fmt.Sprintf("%s endpoint reachable, HTTP %d", name, response.StatusCode))
	return ConnectResult{URL: url, Reachable: true, StatusCode: response.StatusCode}
}

type jwksPayload struct {
	Keys []json.RawMessage `json:"keys"`
}

// checkJWKS makes a GET request to the JWKS endpoint and validates the response contains
// at least one key, confirming both connectivity and that the key server is functioning.
func (p AgentControlAgentConnect) checkJWKS(url string) ConnectResult {
	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            url,
		TimeoutSeconds: 30,
	}

	response, err := p.httpGetter(wrapper)
	if err != nil {
		return ConnectResult{URL: url, Reachable: false, Note: err.Error()}
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != 200 {
		return ConnectResult{
			URL:        url,
			Reachable:  false,
			StatusCode: response.StatusCode,
			Note:       fmt.Sprintf("unexpected HTTP %d", response.StatusCode),
		}
	}

	var jwks jwksPayload
	if err := json.NewDecoder(response.Body).Decode(&jwks); err != nil {
		return ConnectResult{
			URL:        url,
			Reachable:  false,
			StatusCode: response.StatusCode,
			Note:       "invalid JWKS JSON: " + err.Error(),
		}
	}

	if len(jwks.Keys) == 0 {
		return ConnectResult{
			URL:        url,
			Reachable:  false,
			StatusCode: response.StatusCode,
			Note:       "JWKS response contains no keys",
		}
	}

	return ConnectResult{
		URL:        url,
		Reachable:  true,
		StatusCode: response.StatusCode,
		Note:       fmt.Sprintf("%d public key(s) found", len(jwks.Keys)),
	}
}
