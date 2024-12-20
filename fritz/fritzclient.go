package fritz

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bpicode/fritzctl/config"
	"github.com/bpicode/fritzctl/httpread"
	"github.com/bpicode/fritzctl/internal/errors"
	"github.com/bpicode/fritzctl/logger"
	"golang.org/x/crypto/pbkdf2"
)

// Client encapsulates the FRITZ!Box interaction API.
type Client struct {
	Config      *config.Config // The client configuration.
	HTTPClient  *http.Client   // The HTTP client.
	SessionInfo *SessionInfo   // The current session data of the client.
}

// SessionInfo models the xml upon accessing the login endpoint.
// See also https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AVM_Technical_Note_-_Session_ID.pdf.
type SessionInfo struct {
	Challenge string `xml:"Challenge"` // A challenge provided by the FRITZ!Box.
	SID       string `xml:"SID"`       // The session id issued by the FRITZ!Box, "0000000000000000" is considered invalid/"no session".
	BlockTime int    `xml:"BlockTime"` // The time that needs to expire before the next login attempt can be made.
	Rights    Rights `xml:"Rights"`    // The Rights associated withe the session.
	IsPBKDF2  bool
}

// Rights wrap set of pairs (name, access-level).
type Rights struct {
	Names        []string `xml:"Name"`
	AccessLevels []string `xml:"Access"`
}

type LoginResponse struct {
	Challenge string `xml:"Challenge"`
	BlockTime int    `xml:"BlockTime"`
	SID       string `xml:"SID"`
}

// NewClient creates a new Client with values read from a config file, given by the parameter configfile.
// Deprecated: use NewClientFromConfig.
func NewClient(configfile string) (*Client, error) {
	cfg, err := config.New(configfile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read configuration")
	}
	return NewClientFromConfig(cfg), nil
}

// NewClientFromConfig creates a new Client with the passed configuration.
func NewClientFromConfig(cfg *config.Config) *Client {
	tlsConfig := tlsConfigFrom(cfg)
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	httpClient := &http.Client{Transport: transport}
	return &Client{Config: cfg, HTTPClient: httpClient}
}

// Login tries to login into the box and obtain the session id.
// https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AVM_Technical_Note_-_Session_ID_english_2021-05-03.pdf
func (client *Client) Login() error {
	sessionInfo, err := client.obtainChallenge()
	if err != nil {
		return errors.Wrapf(err, "unable to obtain login challenge")
	}
	client.SessionInfo = sessionInfo
	logger.Debug("FRITZ!Box challenge is", client.SessionInfo.Challenge)
	err = client.solveChallenge()
	if err != nil {
		return errors.Wrapf(err, "unable to solve login challenge")
	}
	//client.SessionInfo = newSession
	logger.Info("Login successful")
	return nil
}

func (client *Client) obtainChallenge() (*SessionInfo, error) {
	url := client.Config.GetLoginURL()
	getRemote := func() (*http.Response, error) {
		return client.HTTPClient.Get(url)
	}
	var sessionInfo SessionInfo

	err := httpread.XML(getRemote, &sessionInfo)
	sessionInfo.IsPBKDF2 = strings.HasPrefix(sessionInfo.Challenge, "2$")
	return &sessionInfo, err
}

func (client *Client) solveChallenge() error {
	var challengeResponse string
	var err error
	if client.SessionInfo.IsPBKDF2 {
		logger.Debug("PBKDF2 supported")
		challengeResponse, err = calculatePBKDF2Response(client.SessionInfo.Challenge, client.Config.Login.Password)
		if err != nil {
			return fmt.Errorf("failed to calculate PBKDF2 response: %w", err)
		}
	} else {
		logger.Debug("Falling back to MD5")
		challengeResponse = calculateMD5Response(client.SessionInfo.Challenge, client.Config.Login.Password)
	}

	if client.SessionInfo.BlockTime > 0 {
		logger.Info(fmt.Sprintf("Waiting for %d seconds...\n", client.SessionInfo.BlockTime))
		time.Sleep(time.Duration(client.SessionInfo.BlockTime) * time.Second)
	}

	sid, err := client.sendResponse(challengeResponse)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	client.SessionInfo.SID = sid
	if client.SessionInfo.SID == "0000000000000000" || client.SessionInfo.SID == "" {
		return fmt.Errorf("challenge not solved, got '%s' as session id, check login data", client.SessionInfo.SID)
	}

	return nil
}

func (client *Client) sendResponse(challengeResponse string) (string, error) {
	formData := url.Values{}
	formData.Set("username", client.Config.Login.Username)
	formData.Set("response", challengeResponse)

	resp, err := client.HTTPClient.PostForm(client.Config.GetLoginURL(), formData)
	if err != nil {
		return "", fmt.Errorf("failed to send response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var loginResp LoginResponse
	err = xml.Unmarshal(body, &loginResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	return loginResp.SID, nil
}

func calculatePBKDF2Response(challenge, password string) (string, error) {
	parts := strings.Split(challenge, "$")
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid PBKDF2 challenge format")
	}

	iter1, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid iteration count: %w", err)
	}

	salt1, err := hex.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid salt1: %w", err)
	}

	iter2, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", fmt.Errorf("invalid iteration count: %w", err)
	}

	salt2, err := hex.DecodeString(parts[4])
	if err != nil {
		return "", fmt.Errorf("invalid salt2: %w", err)
	}

	hash1 := pbkdf2.Key([]byte(password), salt1, iter1, sha256.Size, sha256.New)
	hash2 := pbkdf2.Key(hash1, salt2, iter2, sha256.Size, sha256.New)

	return fmt.Sprintf("%s$%x", parts[4], hash2), nil
}

func calculateMD5Response(challenge, password string) string {
	response := challenge + "-" + password
	encoded := []byte(response)

	hasher := md5.New()
	hasher.Write(encoded)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return challenge + "-" + hash
}

func tlsConfigFrom(cfg *config.Config) *tls.Config {
	caCertPool := buildCertPool(cfg)
	return &tls.Config{InsecureSkipVerify: cfg.Pki.SkipTLSVerify, RootCAs: caCertPool}
}

func buildCertPool(cfg *config.Config) *x509.CertPool {
	if cfg.Pki.SkipTLSVerify {
		return nil
	}
	caCertPool := x509.NewCertPool()
	logger.Debug("Reading certificate file", cfg.Pki.CertificateFile)
	caCert, err := ioutil.ReadFile(cfg.Pki.CertificateFile)
	if err != nil {
		logger.Debug("Using host certificates as fallback. Reason: could not read certificate file: ", err)
		return nil
	}
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		logger.Warn("Using host certificates as fallback. Reason: certificate file ", cfg.Pki.CertificateFile, " is not a valid PEM file.")
		return nil
	}
	return caCertPool
}

func (client *Client) query() fritzURLBuilder {
	return newURLBuilder(client.Config).query("sid", client.SessionInfo.SID)
}

func (client *Client) getf(url string) func() (*http.Response, error) {
	return func() (*http.Response, error) {
		logger.Debug("GET", url)
		return client.HTTPClient.Get(url)
	}
}
