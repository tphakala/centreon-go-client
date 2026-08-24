# centreon-go-client

Go client library for the [Centreon Web REST API](https://docs-api.centreon.com/api/centreon-web/).

Zero external dependencies. Requires Go 1.26+.

## Install

```bash
go get github.com/tphakala/centreon-go-client/v2
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	centreon "github.com/tphakala/centreon-go-client/v2"
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
| Hosts | yes | yes** / by ID* | yes | PATCH | yes |
| Host Groups | yes | yes | yes | PUT | yes |
| Host Categories | yes | yes | yes | PUT | yes |
| Host Severities | yes | yes | yes | PUT | yes |
| Host Templates | yes | yes** / by ID* | yes | PATCH | yes |
| Services | yes | yes** | yes | PATCH | yes |
| Service Groups | yes | yes | yes | PUT | yes |
| Service Categories | yes | yes | yes | PUT | yes |
| Service Severities | yes | - | yes | PUT | yes |
| Service Templates | yes | yes** / by ID* | yes | PATCH | yes |
| Time Periods******* | yes | yes | yes | PUT | yes |
| Agent Configurations***** | yes | yes | yes | PUT | yes |
| Additional Connector Configurations***** | yes | yes | yes | PUT | yes |
| Commands | yes | - | yes*** | - | - |
| Monitoring Servers | yes | - | - | - | - |
| Media Library‡ | yes | yes | yes | - | yes |
| Dashboards‡‡ | yes | yes | yes | POST (by id) | yes |
| Access Groups**** | yes | - | - | - | - |
| Connectors**** | yes | - | - | - | - |
| Icons**** | yes | - | - | - | - |
| Graph Templates**** | yes | - | - | - | - |
| Tokens (administration)****** | yes | - | yes | - | yes |

*\* by ID = filtered list lookup (API has no direct GET endpoint)*

*\*\* `Get(ctx, id)` requires Centreon 25.10+ and returns the full configuration including custom macros (hosts: macros defined directly on the host; services: also includes macros inherited from service templates and commands). Host and service templates likewise expose a 25.10+ `Get(ctx, id)` returning the full template with its macros plus category and parent-template (host) or category and group (service) relationships that the list endpoint omits. Earlier versions have no single-resource GET route and return an API error with HTTP status 404: on those versions use `GetByID` for hosts and templates, and `ListByHost` for services (the services endpoint has no by-id lookup).*

*\*\*\* `Commands.Create` returns the full `*Command`, not just an ID, because Centreon Web 25.10 exposes no per-id route for commands (GET/PUT/PATCH/DELETE on `/configuration/commands/{id}` return 404). There is therefore no `Get`, `Update`, or `Delete`, and a created command cannot be removed via the client.*

*\*\*\*\* Read-only lookups exposed by Centreon Web 25.10: `List` and `All` only, with no per-id, create, update, or delete route (POST returns HTTP 405). The Icons list is empty on a stock install because the media library is unpopulated.*

*\*\*\*\*\* `Get(ctx, id)` returns the full configuration including a type-dependent nested object, exposed as a `json.RawMessage` for the caller to decode per `type`: `configuration` for agent configurations (`telegraf` vs `centreon-agent`) and `parameters` for additional connector configurations (`vmware_v6`). The list representation omits that object (and, for agent configurations, `connection_mode`; for connector configurations, `pollers`). `Create` and `Update` (PUT) take the nested object as a `json.RawMessage` too; on update, the `vmware_v6` `parameters.vcenters[]` entries must carry their server-assigned `id`. Requires Centreon 25.10+.*

*\*\*\*\*\*\* API tokens live under `/administration/tokens`, not `/configuration`. `Create` returns the full `*Token` including the one-time secret (`Value`), which the API returns only on create: store it securely and never log it (`Token` redacts `Value` from its `String`/`slog` output). There is no usable per-id lookup (the `GET /administration/tokens/{name}` route is registered but always returns 404 regardless of the identifier) and no `Update` route; `Delete` takes the token name because tokens have no numeric id.*

*‡‡ Dashboards update via `POST /configuration/dashboards/{id}` (`Dashboards.Update`), not PUT or PATCH (`PUT` returns HTTP 405). Sharing is a separate call, `Dashboards.UpdateShares` (`PUT /configuration/dashboards/{id}/shares`), which replaces the contact and contact-group share lists. The `List` representation omits each dashboard's `panels` and `refresh` (use `Get` for those); the create response omits `shares`, `thumbnail`, and `is_favorite`. A panel's `widget_settings` and a dashboard's `thumbnail` are kept as raw JSON (`json.RawMessage`) because their shape is widget- or install-specific; on update the endpoint requires `widget_settings` as a JSON-encoded string, which `Update` re-encodes for you (pass an object, as on create). The `shares.contact_groups` element shape is inferred (it is empty on installs without dashboard-ACL-enabled contact groups).*

*‡ Media images live under `/configuration/medias`. `Create` uploads via multipart/form-data (the raw image bytes as a file part plus a required `directory`), not a JSON body; the client builds the multipart request for you. There is no JSON update modeled, and there is no content-download endpoint (`GET /configuration/medias/{id}/content` returns HTTP 405 on 25.10.16): the list `Media.URL` is the static path to fetch the raw image. Folders are listed via `ListFolders` (`/configuration/media/folders`) and are created implicitly when a media is uploaded into a new directory.*

*\*\*\*\*\*\*\* Time period exceptions are typed, a breaking change from the `[]any` used through v2.0.0. Reads (`TimePeriod.Exceptions`) return `[]TimePeriodException`, each carrying a server-assigned, read-only `id` plus `day_range` and `time_range`. To set exceptions on an update, build `[]TimePeriodExceptionRequest` (`UpdateTimePeriodRequest.Exceptions`); it has no `id` field, so the read-only id is never sent by mistake. This read/write split mirrors a time period's templates, which are read as `[]NamedRef` but written as `[]int`. `Update` still sends `exceptions: []` when none are set, because Centreon 25.10.x requires the field.*

### Platform & Version Gating

`Platform.Versions(ctx)` returns the Centreon web core version plus module and widget versions (`GET /platform/versions`), and `Platform.InstallationStatus(ctx)` returns install and upgrade flags (`GET /platform/installation/status`). These are read-only platform metadata under `/platform`, not `/configuration`.

The version's numeric parts are strings (for example `major` `"25"`, `minor` `"10"`), so comparing them lexically is wrong: `"9"` sorts after `"10"`, and a leading zero such as `"03"` compounds it. Use `PlatformVersions.AtLeast(major, minor)` to gate endpoints that exist only on a newer Centreon (for example the 25.10+ per-id `Get` routes), instead of inferring support from a 404:

```go
v, err := client.Platform.Versions(ctx)
if err == nil && v.AtLeast(25, 10) {
    // safe to call a 25.10-only endpoint such as Hosts.Get(ctx, id)
}
```

`AtLeast` compares major then minor (the fix level is ignored) and returns false if the version does not parse, so an unknown version is treated as too old. Pairing this with `IsRouteNotFound` (see [Error Handling](#error-handling)) lets a consumer both gate ahead of time and recover from an absent endpoint at call time.

`Platform.Features(ctx)` returns the platform feature flags (`GET /platform/features`): `IsCloudPlatform` plus a version-dependent `FeatureFlags map[string]bool`. Some v2 REST route families register only when their feature flag is enabled (for example `/configuration/notifications` under `notification`, resource-access rules under `resource_access_management`, and the vault endpoints under `vault`), so a consumer can call `Features` once and gate those calls with `PlatformFeatures.IsEnabled(flag)`, which returns false for an unknown or disabled flag:

```go
f, err := client.Platform.Features(ctx)
if err == nil && f.IsEnabled("notification") {
    // the /configuration/notifications routes are registered on this instance
}
```

### Administration & Proxy

`AdministrationParameters.Get(ctx)` returns the global monitoring default parameters (`GET /administration/parameters`): the default downtime and acknowledgement behaviour, refresh intervals, and the resource-status full-search flag. It is read-only on Centreon Web 25.10.x (a GET-only route; writes return HTTP 405).

`Proxy.Get(ctx)` returns the central outbound-proxy configuration (`GET /configuration/proxy`), and `Proxy.Update(ctx, cfg)` replaces it (`PUT`, full replace; a nil field is sent as JSON null, which the server treats as clearing that setting). The `password` is a credential returned in cleartext, so `ProxyConfiguration` redacts it from its `String`/`GoString`/`slog` output while `json.Marshal` still writes it (so `Update` sends the real value). `Update` deliberately omits the `protocol` field from the request body: on Centreon Web 25.10.16 any body carrying the `protocol` key returns HTTP 500 for every value type, while the GET response always includes `protocol`. Omitting it lets a `Get`-modify-`Update` round-trip succeed; `protocol` is not settable via v2 REST and keeps its server default (`"http://"`).

`Authentication` reads and updates the four authentication providers exposed as singletons under `/administration/authentication/providers`: `GetLocal`/`UpdateLocal` (the local password policy), `GetOpenID`/`UpdateOpenID` (OpenID Connect), `GetSAML`/`UpdateSAML` (SAML), and `GetWebSSO`/`UpdateWebSSO` (Web-SSO). Each update is a full-object replace (PUT), so the pattern is read, mutate, write back; the read struct doubles as the write body. `openid.client_secret` and `saml.certificate` are credentials returned in cleartext, so `OpenIDProvider` and `SAMLProvider` redact them from `String`/`GoString`/`slog` output while `json.Marshal` still writes them (the PUT needs them). OpenID also accepts PATCH for a partial update, but that subset is not modeled; `UpdateOpenID` uses the full-replace PUT. The OpenID and SAML `roles_mapping`/`groups_mapping` `relations` are kept as raw JSON pending a configured instance to pin their element shape.

### User & Contact Management

| Resource | List | Update | Notes |
|----------|------|--------|-------|
| Users | yes | PATCH* | No create/delete via API |
| Contact Groups | yes | - | Read-only |
| Contact Templates | yes | - | Read-only |
| User Filters | yes | PUT + PATCH** | PUT replaces; PATCH reorders |

*\* On Centreon Web 25.10.x, `UserService.Update` (`PATCH /configuration/users/{id}`) returns an `*APIError` with HTTP 404 (`No route found`): the route is unregistered, and PUT, POST, and DELETE for users and contacts are likewise absent, so users and contacts are effectively read-only through the v2 REST API on that version. The method is retained for other Centreon versions that register the route.*

*\*\* User filters send the plural `criterias` array on Create and Update (Centreon 25.10.x rejects the singular `criteria`), and `UserFilter` now also decodes the `order` field. `Update` (PUT) replaces a filter, including its name; `Patch` (PATCH) is a reorder-only route that sends `order` and cannot rename. The element shape of `FilterCriteria` is not yet verified against a populated filter.*

`CurrentUser` provides access to the authenticated user's own context: `GetParameters(ctx)` returns the profile and UI preferences (`GET /configuration/users/current/parameters`), `GetACLActions(ctx)` the allowed real-time actions per object type (`GET /users/acl/actions`), and `GetACLPermissions(ctx)` the effective feature permissions as a sparse `map[string]bool` (`GET /users/acl/permissions`). `UpdateParameters(ctx, req)` partially updates the current user's own preferences (`PATCH /configuration/users/current/parameters`). The endpoint's schema is closed, so only `theme` and `user_interface_density` are writable (every other property is rejected on 25.10.16); `UpdateCurrentUserParametersRequest` models exactly those two as optional pointer fields, and only the ones set are sent.

### Monitoring (real-time)

| Resource | Methods |
|----------|---------|
| Unified Resources | List, GetHost, GetService† |
| Monitoring Hosts | List, Get, StatusCounts, Services, Timeline |
| Monitoring Services | List, StatusCounts, Timeline, Metrics‡ |
| Monitoring Server Status | List§ |
| Downtimes | List, Get, Cancel, ListForHost, ListForService, CreateForHost, CreateForService, CancelForHost, CancelForService |
| Acknowledgements | List, Get, ListForHost, ListForService, CreateForHost, CreateForService, CancelForHost, CancelForService |
| Notification Policies | GetForHost, GetForService |

*Notification policy reads may return an `*APIError` with HTTP 500 or 404 on some hosts on Centreon 25.10.x; this is a server-side defect on that version, not a client error.*

*† `GetHost` populates `HostID`, and `GetService` populates `HostID` and `ServiceID`, from the call arguments, because the Centreon 25.10.x per-id detail endpoint omits them at the top level; that endpoint also does not report `is_notification_enabled`, so `NotificationEnabled` is populated only by `List` (and `All`).*

*‡ `Metrics` returns an empty result rather than an error when a service has no performance data: Centreon 25.10.x answers that case with HTTP 404 `metrics not found`, which the client maps to an empty slice. Any other 404 (for example a nonexistent host or service) is surfaced as an `*APIError`.*

*§ `Monitoring Server Status` is the real-time `/monitoring/servers` endpoint (`MonitoringServerStatus`), distinct from the configuration-side `Monitoring Servers` (`/configuration/monitoring-servers`) in the Configuration CRUD table above. It is list-only: `GET /monitoring/servers/{id}` returns HTTP 404 on Centreon 25.10.16, so there is no `Get`. `MonitoringServerStatus.LastAlive` is a Unix epoch `int64` (the poller's last heartbeat), in contrast to the config `MonitoringServer.LastRestart`, an RFC3339 `*time.Time`. The endpoint coerces a null heartbeat and running flag to `0` / `false` rather than emitting JSON null (verified live), so both are plain value fields. The sibling real-time `/monitoring/hostgroups` and `/monitoring/servicegroups` endpoints (also from issue #86) are deferred: their populated element shape could not be captured live and is tracked separately.*

*The `MonitoringHost`, `MonitoringService`, and unified `MonitoringResource` structs decode additional real-time fields across these structs, with the exact set differing per endpoint (for example `duration` on services and resources, `last_update` and the per-check timers on hosts, `criticality` on hosts and services, and flapping and check-toggle flags on resources). `check_attempt` is asymmetric across endpoints: it is a JSON number on `/monitoring/hosts` (typed `int` on `MonitoringHost`) and a JSON string on `/monitoring/services` (typed `string` on `MonitoringService`). On unified resources, `severity` and `icon` are exposed as `*json.RawMessage` (nil for a null or absent value) because they are null on Centreon 25.10.16 and their populated shape is not yet verified against a live instance; `GetHost` and `GetService` do not populate these enrichment fields (they were verified only on the list endpoint).*

### Downtime Management

```go
// List all active downtimes
downtimes, _ := client.Downtimes.List(ctx)

// List downtimes for a specific host
downtimes, _ := client.Downtimes.ListForHost(ctx, hostID)

// Schedule a downtime on a host
client.Downtimes.CreateForHost(ctx, hostID, &centreon.CreateHostDowntimeRequest{
    Comment:   "Scheduled maintenance",
    StartTime: time.Now(),
    EndTime:   time.Now().Add(2 * time.Hour),
    IsFixed:   true,
})

// Schedule a downtime on a service
client.Downtimes.CreateForService(ctx, hostID, serviceID, &centreon.CreateServiceDowntimeRequest{
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
client.Acknowledgements.CreateForHost(ctx, hostID, &centreon.CreateHostAcknowledgementRequest{
    Comment:  "Investigating",
    IsSticky: true,
})

// Acknowledge a service
client.Acknowledgements.CreateForService(ctx, hostID, serviceID, &centreon.CreateServiceAcknowledgementRequest{
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

## Resource Status Filters

The unified `/monitoring/resources` listing (`client.Monitoring.List` / `client.Monitoring.All`) accepts dedicated name-based array filters that the numeric search DSL cannot express. Combine them freely; each is sent as a scalar query parameter whose value is a JSON-encoded array (for example `statuses=["OK","WARNING"]`).

```go
resp, err := client.Monitoring.List(ctx,
    centreon.WithResourceTypes("host", "service"),
    centreon.WithStatuses("DOWN", "CRITICAL", "UNKNOWN"),
    centreon.WithHostGroupNames("ESX-Paris"),
)
```

Prefer these name-based filters over the numeric status codes exposed by the search DSL: the codes collide across resource kinds, whereas the names do not. A host `DOWN` and a service `WARNING` share status code `1`, and a host `UNREACHABLE` and a service `CRITICAL` share code `2`, so a numeric `status_code` filter cannot tell them apart. `WithStatuses("DOWN")` selects exactly the down hosts.

| Helper | Query parameter | Allowed values |
|--------|-----------------|----------------|
| `WithResourceTypes` | `types` | `host`, `service`, `metaservice` |
| `WithStatuses` | `statuses` | `OK`, `UP`, `WARNING`, `DOWN`, `CRITICAL`, `UNREACHABLE`, `UNKNOWN`, `PENDING` |
| `WithStatusTypes` | `status_types` | `hard`, `soft` |
| `WithStates` | `states` | `unhandled_problems`, `resources_problems`, `in_downtime`, `acknowledged`, `in_flapping`, `all` |
| `WithHostGroupNames` | `hostgroup_names` | host group names |
| `WithServiceGroupNames` | `servicegroup_names` | service group names |
| `WithHostCategoryNames` | `host_category_names` | host category names |
| `WithServiceCategoryNames` | `service_category_names` | service category names |
| `WithMonitoringServerNames` | `monitoring_server_names` | monitoring server (poller) names |

Repeated calls to the same helper accumulate values. For any filter these helpers do not cover, use the general escape hatch `WithArrayFilter(key string, values ...string)`, which emits the same JSON-scalar wire format under an arbitrary key.

The parameter names and wire format were verified against the centreon-web `FindResourcesRequestValidator.php` source, not yet against a live 25.10 instance.

## Update Patterns

The API uses two update methods depending on the resource:

```go
// PATCH (partial update): hosts, services, templates
// Only specified fields are changed. Use pointer fields.
err := client.Hosts.Update(ctx, hostID, centreon.UpdateHostRequest{
    Alias: new("updated alias"),  // Go 1.26 new(expr)
})

// PUT (full replacement): groups, categories, severities, time periods
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

    // Distinguish an absent endpoint (routing 404) from a missing resource
    // (resource 404, for example "Host not found"). Useful to degrade
    // gracefully when an endpoint exists only on a newer Centreon. The client
    // classifies the 404 internally, so a gateway consumer never has to parse
    // the (remote-controlled) error message itself.
    if centreon.IsRouteNotFound(err) {
        fmt.Println("endpoint not available on this Centreon version")
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

Apache License 2.0. See [LICENSE](LICENSE) for details.
