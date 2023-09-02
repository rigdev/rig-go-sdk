# Rig Golang SDK

## Overview

Rig provides the tools, modules and infrastructure you need to develop and manage applications on Kubernetes. The Rig Golang SDK enables access to Rig services from privileged environments (such as servers or cloud) in Golang.

For more information, visit the [Rig Golang SDK setup guide](https://docs.rig.dev/sdks/golang).

## Installation

The Rig Golang SDK can be installed using the go install utility:

```
# Install the latest version:
go get github.com/rigdev/rig-go-sdk@latest

# Or install a specific version:
go get github.com/rigdev/rig-go-sdk@x.x.x
```

## Setup the Client

To setup the client use the `rig.NewClient` method:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bufbuild/connect-go"
	rig "github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig-go-api/api/v1/user"
)

func main() {
	client := rig.NewClient()

	// you can now make requests to Rig
	if _, err := client.User().Create(context.Background(), connect.NewRequest(&user.CreateRequest{
		Initializers: []*user.Update{},
	})); err != nil {
		log.Fatal(err)
	}

	fmt.Println("success")
}
```

### Host

By default, the SDK will connect to `http://localhost:4747`. To change this, use the `rig.WithHost(...)` option:

```go
	client := rig.NewClient(rig.WithHost("my-rig:4747"))
```

### Credentials

By default, the SDK will use the environment variables `RIG_CLIENT_ID` and `RIG_CLIENT_SECRET` to read the credentials. To explicitly set the credentials, use `rig.WithClientCredentials(...)` option:

```go
	client := rig.NewClient(rig.WithClientCredentials(rig.ClientCredential{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
	}))
```

## Documentation

- [Install Rig](https://docs.rig.dev/getting-started)
- [Setup Users](https://docs.rig.dev/users)
- [Deploy Capsules](https://docs.rig.dev/capsules)
