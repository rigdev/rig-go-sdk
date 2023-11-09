package rig

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/authentication/authenticationconnect"
	"github.com/rigdev/rig-go-api/api/v1/build/buildconnect"
	"github.com/rigdev/rig-go-api/api/v1/capsule/capsuleconnect"
	"github.com/rigdev/rig-go-api/api/v1/cluster/clusterconnect"
	"github.com/rigdev/rig-go-api/api/v1/database/databaseconnect"
	"github.com/rigdev/rig-go-api/api/v1/group/groupconnect"
	"github.com/rigdev/rig-go-api/api/v1/project/projectconnect"
	projectsettingsconnect "github.com/rigdev/rig-go-api/api/v1/project/settings/settingsconnect"
	"github.com/rigdev/rig-go-api/api/v1/service_account/service_accountconnect"
	storagesettingsconnect "github.com/rigdev/rig-go-api/api/v1/storage/settings/settingsconnect"
	"github.com/rigdev/rig-go-api/api/v1/storage/storageconnect"
	usersettingsconnect "github.com/rigdev/rig-go-api/api/v1/user/settings/settingsconnect"
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
	// UserSettings service for managing settings for the entire User module.
	UserSettings() usersettingsconnect.ServiceClient
	// ServiceAccount service for creating and maintaining OAuth2 Service Accounts.
	ServiceAccount() service_accountconnect.ServiceClient
	// Group service for managing groups and associating users to them.
	Group() groupconnect.ServiceClient
	// Storage service for interacting with the Storage backends, such as creating buckets and uploading files.
	Storage() storageconnect.ServiceClient
	// StorageSettings service for configuring the Storage backends.
	StorageSettings() storagesettingsconnect.ServiceClient
	// Database service for managing databases related to the project.
	Database() databaseconnect.ServiceClient
	// Capsule API for managing the lifecycle of Capsules.
	Capsule() capsuleconnect.ServiceClient
	// Project API for configuring the overall settings of the project.
	Project() projectconnect.ServiceClient
	// ProjectSettings service for managing settings of projects
	ProjectSettings() projectsettingsconnect.ServiceClient
	// Cluster service for managing the Rig cluster
	Cluster() clusterconnect.ServiceClient
	Build() buildconnect.ServiceClient

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
	host  string
	login *authentication.LoginRequest
	sm    SessionManager
	hc    *http.Client
	ics   []Interceptor
}

type client struct {
	cfg             *config
	authentication  authenticationconnect.ServiceClient
	user            userconnect.ServiceClient
	userSettings    usersettingsconnect.ServiceClient
	service_account service_accountconnect.ServiceClient
	group           groupconnect.ServiceClient
	storage         storageconnect.ServiceClient
	storageSettings storagesettingsconnect.ServiceClient
	database        databaseconnect.ServiceClient
	capsule         capsuleconnect.ServiceClient
	project         projectconnect.ServiceClient
	projectSettings projectsettingsconnect.ServiceClient
	cluster         clusterconnect.ServiceClient
	build           buildconnect.ServiceClient
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
		userSettings:    usersettingsconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		group:           groupconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		storage:         storageconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		storageSettings: storagesettingsconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		database:        databaseconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		capsule:         capsuleconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		project:         projectconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		projectSettings: projectsettingsconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		cluster:         clusterconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
		build:           buildconnect.NewServiceClient(cfg.hc, cfg.host, connect.WithInterceptors(ics...)),
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

func (c *client) UserSettings() usersettingsconnect.ServiceClient {
	return c.userSettings
}

func (c *client) ServiceAccount() service_accountconnect.ServiceClient {
	return c.service_account
}

func (c *client) Group() groupconnect.ServiceClient {
	return c.group
}

func (c *client) Storage() storageconnect.ServiceClient {
	return c.storage
}

func (c *client) StorageSettings() storagesettingsconnect.ServiceClient {
	return c.storageSettings
}

func (c *client) Database() databaseconnect.ServiceClient {
	return c.database
}

func (c *client) Capsule() capsuleconnect.ServiceClient {
	return c.capsule
}

func (c *client) Project() projectconnect.ServiceClient {
	return c.project
}

func (c *client) ProjectSettings() projectsettingsconnect.ServiceClient {
	return c.projectSettings
}

func (c *client) Cluster() clusterconnect.ServiceClient {
	return c.cluster
}

func (c *client) Build() buildconnect.ServiceClient {
	return c.build
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}
