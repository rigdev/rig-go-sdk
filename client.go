package rig

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/activity/activityconnect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/authentication/authenticationconnect"
	"github.com/rigdev/rig-go-api/api/v1/capsule/capsuleconnect"
	"github.com/rigdev/rig-go-api/api/v1/cluster/clusterconnect"
	"github.com/rigdev/rig-go-api/api/v1/environment/environmentconnect"
	"github.com/rigdev/rig-go-api/api/v1/group/groupconnect"
	"github.com/rigdev/rig-go-api/api/v1/image/imageconnect"
	"github.com/rigdev/rig-go-api/api/v1/metrics/metricsconnect"
	"github.com/rigdev/rig-go-api/api/v1/project/projectconnect"
	"github.com/rigdev/rig-go-api/api/v1/role/roleconnect"
	"github.com/rigdev/rig-go-api/api/v1/service_account/service_accountconnect"
	"github.com/rigdev/rig-go-api/api/v1/settings/settingsconnect"
	"github.com/rigdev/rig-go-api/api/v1/user/userconnect"
	"golang.org/x/net/http2"
)

// Client for interacting with the Rig APIs. Each of the services are available as a `connect.build` Client,
// allowing for a variety of communication options, such as gRPC, connect.build, HTTP/JSON.
type Client interface {
	// Authentication service for logging in and registering new users.
	// If you are using OAuth Client Credentials for, see `WithClientCredential`.
	Authentication() authenticationconnect.ServiceClient
	// User service for managing users.
	User() userconnect.ServiceClient
	// ServiceAccount service for creating and maintaining OAuth2 Service Accounts.
	ServiceAccount() service_accountconnect.ServiceClient
	// Group service for managing groups and associating users to them.
	Group() groupconnect.ServiceClient
	// Capsule API for managing the lifecycle of Capsules.
	Capsule() capsuleconnect.ServiceClient
	// Project API for configuring the overall settings of the project.
	Project() projectconnect.ServiceClient
	// Cluster service for managing the Rig cluster
	Cluster() clusterconnect.ServiceClient

	Image() imageconnect.ServiceClient

	Environment() environmentconnect.ServiceClient

	Role() roleconnect.ServiceClient

	Settings() settingsconnect.ServiceClient

	Metrics() metricsconnect.ServiceClient

	Activity() activityconnect.ServiceClient

	// Set the access- and refresh token pair. This will use the underlying SessionManager.
	// The client will refresh the tokens in the background as needed.
	SetAccessToken(accessToken, refreshToken string)
}

// ClientCredential to use for authenticating with the backend using the OAuth2 Client Credentials flow.
// Use `WithClientCredentials` when creating a new client, to set them.
// The client will automatically register a ClientCredential if the `RIG_CLIENT_ID` and
// `RIG_CLIENT_SECRET` environment variables are set.
type ClientCredential struct {
	ClientID     string
	ClientSecret string
}

type config struct {
	host      string
	login     *authentication.LoginRequest
	sm        SessionManager
	hc        *http.Client
	ics       []Interceptor
	basicAuth string
}

type client struct {
	cfg             *config
	authentication  authenticationconnect.ServiceClient
	user            userconnect.ServiceClient
	service_account service_accountconnect.ServiceClient
	group           groupconnect.ServiceClient
	capsule         capsuleconnect.ServiceClient
	project         projectconnect.ServiceClient
	cluster         clusterconnect.ServiceClient
	image           imageconnect.ServiceClient
	environment     environmentconnect.ServiceClient
	role            roleconnect.ServiceClient
	settings        settingsconnect.ServiceClient
	metrics         metricsconnect.ServiceClient
	activity        activityconnect.ServiceClient
}

var _h2cClient = &http.Client{
	Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		ReadIdleTimeout:  30 * time.Second,
		WriteByteTimeout: 30 * time.Second,
	},
}

func NewClient(opts ...Option) Client {
	cfg := &config{
		host: getEnv("RIG_HOST", "http://localhost:4747"),
		sm:   &simpleSessionManager{},
	}

	if clientID, ok := os.LookupEnv("RIG_CLIENT_ID"); ok {
		cfg.login = &authentication.LoginRequest{
			Method: &authentication.LoginRequest_ClientCredentials{
				ClientCredentials: &authentication.ClientCredentials{
					ClientId:     clientID,
					ClientSecret: os.Getenv("RIG_CLIENT_SECRET"),
				},
			},
		}
	}

	for _, o := range opts {
		o.apply(cfg)
	}

	if cfg.hc == nil {
		// Support h2c (http2 plaintext) for http servers, to support BIDI streams.
		if strings.HasPrefix(cfg.host, "http:") {
			cfg.hc = _h2cClient
		} else {
			cfg.hc = http.DefaultClient
		}
	}

	i := &authInterceptor{
		cfg: cfg,
	}

	ics := []connect.Interceptor{i}
	ics = append(ics, cfg.ics...)

	return &client{
		cfg:             cfg,
		authentication:  authenticationconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		service_account: service_accountconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		user:            userconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		group:           groupconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		capsule:         capsuleconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		project:         projectconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		cluster:         clusterconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		image:           imageconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		environment:     environmentconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		role:            roleconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		settings:        settingsconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		metrics:         metricsconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		activity:        activityconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
	}
}

func (c *client) SetAccessToken(accessToken, refreshToken string) {
	c.cfg.sm.SetAccessToken(accessToken, refreshToken)
}

func (c *client) Authentication() authenticationconnect.ServiceClient {
	return c.authentication
}

func (c *client) User() userconnect.ServiceClient {
	return c.user
}

func (c *client) ServiceAccount() service_accountconnect.ServiceClient {
	return c.service_account
}

func (c *client) Group() groupconnect.ServiceClient {
	return c.group
}

func (c *client) Capsule() capsuleconnect.ServiceClient {
	return c.capsule
}

func (c *client) Project() projectconnect.ServiceClient {
	return c.project
}

func (c *client) Settings() settingsconnect.ServiceClient {
	return c.settings
}

func (c *client) Cluster() clusterconnect.ServiceClient {
	return c.cluster
}

func (c *client) Image() imageconnect.ServiceClient {
	return c.image
}

func (c *client) Environment() environmentconnect.ServiceClient {
	return c.environment
}

func (c *client) Role() roleconnect.ServiceClient {
	return c.role
}

func (c *client) Metrics() metricsconnect.ServiceClient {
	return c.metrics
}

func (c *client) Activity() activityconnect.ServiceClient {
	return c.activity
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}
