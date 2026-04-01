# centreon-go-client

Go client library for the [Centreon Web REST API](https://docs-api.centreon.com/api/centreon-web/).

Zero external dependencies. Requires Go 1.26+.

## Install

```bash
go get github.com/tphakala/centreon-go-client
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	centreon "github.com/tphakala/centreon-go-client"
)

func main() {
	ctx := context.Background()

	// Session-based auth (auto-renews on 401)
	client, err := centreon.NewClient("https://centreon.example.com",
		centreon.WithCredentials("admin", "password"),
		centreon.WithVersion("v24.04"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Login(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(ctx)

	// Or use a pre-existing API token (no login needed)
	// client, _ := centreon.NewClient("https://centreon.example.com",
	//     centreon.WithAPIToken("your-token"),
	// )

	// List hosts
	hosts, err := client.Hosts.List(ctx,
		centreon.WithSearch(centreon.Lk("host.name", "prod-%")),
		centreon.WithSort(map[string]string{"host.name": "ASC"}),
		centreon.WithLimit(50),
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hosts.Result {
		fmt.Printf("%d: %s (%s)\n", h.ID, h.Name, h.Address)
	}
}
```

## Features

### Configuration CRUD

| Resource | List | Get | Create | Update | Delete |
|----------|------|-----|--------|--------|--------|
| Hosts | yes | by ID* | yes | PATCH | yes |
| Host Groups | yes | yes | yes | PUT | yes |
| Host Categories | yes | yes | yes | PUT | yes |
| Host Severities | yes | yes | yes | PUT | yes |
| Host Templates | yes | by ID* | yes | PATCH | yes |
| Services | yes | by ID* | yes | PATCH | yes |
| Service Groups | yes | - | yes | - | yes |
| Service Categories | yes | - | yes | - | yes |
| Service Severities | yes | - | yes | PUT | yes |
| Service Templates | yes | by ID* | yes | PATCH | yes |
| Time Periods | yes | yes | yes | PUT | yes |
| Commands | yes | - | - | - | - |
| Monitoring Servers | yes | - | - | - | - |

*\* by ID = filtered list lookup (API has no direct GET endpoint)*

### User & Contact Management

| Resource | List | Update | Notes |
|----------|------|--------|-------|
| Users | yes | PATCH | No create/delete via API |
| Contact Groups | yes | - | Read-only |
| Contact Templates | yes | - | Read-only |
| User Filters | yes | PUT + PATCH | Full CRUD |

### Monitoring (real-time)

| Resource | Methods |
|----------|---------|
| Unified Resources | List, GetHost, GetService |
| Monitoring Hosts | List, Get, StatusCounts, Services, Timeline |
| Monitoring Services | List, StatusCounts |
| Downtimes | List, Get, Cancel, ListForHost, ListForService, CreateForHost, CreateForService, CancelForHost, CancelForService |
| Acknowledgements | List, Get, ListForHost, ListForService, CreateForHost, CreateForService, CancelForHost, CancelForService |
| Notification Policies | GetForHost, GetForService |

### Downtime Management

```go
// List all active downtimes
downtimes, _ := client.Downtimes.List(ctx)

// List downtimes for a specific host
downtimes, _ := client.Downtimes.ListForHost(ctx, hostID)

// Schedule a downtime on a host
client.Downtimes.CreateForHost(ctx, hostID, &centreon.CreateDowntimeRequest{
    Comment:   "Scheduled maintenance",
    StartTime: time.Now(),
    EndTime:   time.Now().Add(2 * time.Hour),
    IsFixed:   true,
})

// Schedule a downtime on a service
client.Downtimes.CreateForService(ctx, hostID, serviceID, &centreon.CreateDowntimeRequest{
    Comment:   "Service patch",
    StartTime: time.Now(),
    EndTime:   time.Now().Add(30 * time.Minute),
    IsFixed:   true,
})

// Cancel a downtime
client.Downtimes.Cancel(ctx, downtimeID)

// Cancel all downtimes for a host
client.Downtimes.CancelForHost(ctx, hostID)
```

### Acknowledgement Management

```go
// List all acknowledgements
acks, _ := client.Acknowledgements.List(ctx)

// Acknowledge a host
client.Acknowledgements.CreateForHost(ctx, hostID, &centreon.CreateAcknowledgementRequest{
    Comment:  "Investigating",
    IsSticky: true,
})

// Acknowledge a service
client.Acknowledgements.CreateForService(ctx, hostID, serviceID, &centreon.CreateAcknowledgementRequest{
    Comment:             "Known issue, fix in progress",
    IsSticky:            true,
    IsPersistentComment: true,
})

// Cancel acknowledgement for a host
client.Acknowledgements.CancelForHost(ctx, hostID)
```

### Bulk Operational Actions

- **Acknowledge** multiple resources at once
- **Schedule downtime** on multiple resources
- **Force check** on multiple resources
- **Submit** passive check results
- **Add comments**

### Apply Configuration

```go
// Reload a specific poller
client.MonitoringServers.GenerateAndReload(ctx, serverID)

// Reload all pollers
client.MonitoringServers.GenerateAndReloadAll(ctx)
```

## Pagination

All list endpoints return paginated results:

```go
// Manual pagination
resp, err := client.Hosts.List(ctx, centreon.WithPage(1), centreon.WithLimit(10))
fmt.Println(resp.Meta.Total) // total count across all pages

// Automatic iteration (fetches pages on demand)
for host, err := range client.Hosts.All(ctx) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(host.Name)
}
```

## Search Filters

Build complex search queries with the fluent filter API:

```go
// Simple equality
filter := centreon.Eq("host.name", "prod-01")

// Combine with And/Or
filter := centreon.And(
    centreon.Lk("host.name", "prod-%"),
    centreon.Or(
        centreon.Eq("host.address", "10.0.0.1"),
        centreon.Eq("host.address", "10.0.0.2"),
    ),
)

resp, err := client.Hosts.List(ctx, centreon.WithSearch(filter))
```

**Available operators:** `Eq`, `Neq`, `Lt`, `Le`, `Gt`, `Ge`, `Lk` (like), `Nk` (not like), `In`, `Ni` (not in), `Rg` (regex)

## Update Patterns

The API uses two update methods depending on the resource:

```go
// PATCH (partial update) — hosts, services, templates
// Only specified fields are changed. Use pointer fields.
err := client.Hosts.Update(ctx, hostID, centreon.UpdateHostRequest{
    Alias: new("updated alias"),  // Go 1.26 new(expr)
})

// PUT (full replacement) — groups, categories, severities, time periods
// All fields are sent. Omitted fields reset to defaults.
err := client.HostGroups.Update(ctx, groupID, centreon.UpdateHostGroupRequest{
    Name:  "linux-servers",
    Alias: "All Linux Servers",
})
```

## Authentication

Two modes supported:

**Session-based** (username/password): Call `Login()` to get a token. The client auto-renews on 401 by re-authenticating with stored credentials. Tokens expire after 1 hour of inactivity.

**API token**: Pass a pre-existing long-lived token with `WithAPIToken()`. No login call needed.

```go
// Session-based
client, _ := centreon.NewClient(url, centreon.WithCredentials("user", "pass"))
client.Login(ctx)
defer client.Logout(ctx)

// API token
client, _ := centreon.NewClient(url, centreon.WithAPIToken("token"))
```

## Error Handling

```go
import "errors"

resp, err := client.Hosts.List(ctx)
if err != nil {
    // Check for API errors
    if apiErr, ok := errors.AsType[*centreon.APIError](err); ok {
        fmt.Printf("HTTP %d: %s\n", apiErr.HTTPStatus, apiErr.Message)
    }

    // Check for not-found (from GetByID)
    if nfErr, ok := errors.AsType[*centreon.NotFoundError](err); ok {
        fmt.Printf("%s %d not found\n", nfErr.Resource, nfErr.ID)
    }
}
```

## Timeout & Logging

```go
// Custom timeout (default 30s)
client, _ := centreon.NewClient(url,
    centreon.WithTimeout(60 * time.Second),
)

// Enable structured logging
client, _ := centreon.NewClient(url,
    centreon.WithLogger(slog.Default()),
)
// Debug: logs every request (method, URL, status)
// Info:  logs token re-authentication
// Error: logs API errors and request failures
```

## API Version

The client defaults to `latest`. Pin to a specific version to avoid breaking changes:

```go
client, _ := centreon.NewClient(url, centreon.WithVersion("v24.04"))
```

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
