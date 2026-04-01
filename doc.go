// Package centreon provides a Go client for the Centreon Web REST API.
//
// # Getting Started
//
// Create a client with session-based authentication:
//
//	client, err := centreon.NewClient("https://centreon.example.com",
//	    centreon.WithCredentials("admin", "password"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := client.Login(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Logout(ctx)
//
// Or use a pre-existing API token:
//
//	client, err := centreon.NewClient("https://centreon.example.com",
//	    centreon.WithAPIToken("your-token"),
//	)
//
// # API Version
//
// The client defaults to "latest". Pin a specific version to avoid breaking changes:
//
//	client, err := centreon.NewClient(url, centreon.WithVersion("v24.04"))
//
// # Listing Resources
//
// All list endpoints support pagination, search filters, and sorting:
//
//	resp, err := client.Hosts.List(ctx,
//	    centreon.WithSearch(centreon.Eq("host.name", "prod-01")),
//	    centreon.WithSort(map[string]string{"host.name": "ASC"}),
//	    centreon.WithLimit(50),
//	)
//
// Use the All method to iterate over all pages automatically:
//
//	for host, err := range client.Hosts.All(ctx) {
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(host.Name)
//	}
//
// # Search Filters
//
// Build search queries using the fluent filter API with 12 operators:
//
//	filter := centreon.And(
//	    centreon.Lk("host.name", "prod-%"),
//	    centreon.Gt("host.id", 100),
//	)
//
// Available operators: [Eq], [Neq], [Lt], [Le], [Gt], [Ge], [Lk], [Nk], [In], [Ni], [Rg].
// Combine with [And] and [Or].
//
// # Update Patterns
//
// Resources use either PATCH (partial update) or PUT (full replacement):
//
//	// PATCH — only specified fields are changed
//	client.Hosts.Update(ctx, id, &centreon.UpdateHostRequest{
//	    Alias: new("updated"),
//	})
//
//	// PUT — all fields are sent
//	client.HostGroups.Update(ctx, id, centreon.UpdateHostGroupRequest{
//	    Name:  "linux-servers",
//	    Alias: "Linux Servers",
//	})
//
// # Error Handling
//
// API errors are returned as [*APIError]. Not-found errors from [GetByID]
// methods are returned as [*NotFoundError]. Use [errors.AsType] for type-safe checks:
//
//	if apiErr, ok := errors.AsType[*centreon.APIError](err); ok {
//	    fmt.Printf("HTTP %d: %s\n", apiErr.HTTPStatus, apiErr.Message)
//	}
package centreon
