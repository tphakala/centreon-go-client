package centreon_test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	centreon "github.com/tphakala/centreon-go-client"
)

func ExampleNewClient_sessionAuth() {
	ctx := context.Background()

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
	defer client.Logout(ctx) //nolint:errcheck // example cleanup

	_ = client // use client
}

func ExampleNewClient_apiToken() {
	client, err := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("your-long-lived-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client
}

func ExampleNewClient_withOptions() {
	client, err := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
		centreon.WithTimeout(60*time.Second),
		centreon.WithLogger(slog.Default()),
		centreon.WithVersion("v24.04"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client
}

func ExampleHostService_List() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	// List hosts with search filter and sorting
	resp, err := client.Hosts.List(ctx,
		centreon.WithSearch(centreon.Lk("host.name", "prod-%")),
		centreon.WithSort(map[string]string{"host.name": "ASC"}),
		centreon.WithLimit(50),
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, h := range resp.Result {
		fmt.Printf("%d: %s (%s)\n", h.ID, h.Name, h.Address)
	}
}

func ExampleHostService_All() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	// Iterate all hosts across all pages
	for host, err := range client.Hosts.All(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(host.Name)
	}
}

func ExampleHostService_Create() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	id, err := client.Hosts.Create(ctx, centreon.CreateHostRequest{
		MonitoringServerID: 1,
		Name:               "new-host",
		Address:            "192.168.1.100",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created host:", id)
}

func ExampleHostService_Update() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	// PATCH — only specified fields are changed
	err := client.Hosts.Update(ctx, 42, centreon.UpdateHostRequest{
		Alias: new("Production Server"),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleAnd() {
	// Complex search: prod hosts in a specific address range
	filter := centreon.And(
		centreon.Lk("host.name", "prod-%"),
		centreon.Or(
			centreon.Lk("host.address", "10.0.%"),
			centreon.Lk("host.address", "192.168.%"),
		),
	)
	_ = filter // use with centreon.WithSearch(filter)
}

func ExampleOperationsService_Acknowledge() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	err := client.Operations.Acknowledge(ctx, &centreon.AcknowledgeRequest{
		Resources: []centreon.ResourceRef{
			{Type: "host", ID: 42},
		},
		Comment:  "Acknowledged by automation",
		IsSticky: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleOperationsService_Downtime() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	err := client.Operations.Downtime(ctx, &centreon.DowntimeRequest{
		Resources: []centreon.ResourceRef{
			{Type: "host", ID: 42},
		},
		Comment:   "Scheduled maintenance",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Hour),
		Fixed:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleMonitoringHostService_StatusCounts() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	counts, err := client.MonitoringHosts.StatusCounts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Hosts: %d UP, %d DOWN, %d Unreachable\n",
		counts.Up, counts.Down, counts.Unreachable)
}

func ExampleTimePeriodService_Create() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	id, err := client.TimePeriods.Create(ctx, centreon.CreateTimePeriodRequest{
		Name:  "business-hours",
		Alias: "Business Hours",
		Days: []centreon.TimePeriodDay{
			{Day: "monday", TimeRanges: []centreon.TimeRange{{Start: "08:00", End: "17:00"}}},
			{Day: "tuesday", TimeRanges: []centreon.TimeRange{{Start: "08:00", End: "17:00"}}},
			{Day: "wednesday", TimeRanges: []centreon.TimeRange{{Start: "08:00", End: "17:00"}}},
			{Day: "thursday", TimeRanges: []centreon.TimeRange{{Start: "08:00", End: "17:00"}}},
			{Day: "friday", TimeRanges: []centreon.TimeRange{{Start: "08:00", End: "17:00"}}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created time period:", id)
}

func ExampleMonitoringServerService_GenerateAndReloadAll() {
	ctx := context.Background()
	client, _ := centreon.NewClient("https://centreon.example.com",
		centreon.WithAPIToken("token"),
	)

	// Apply configuration changes to all pollers
	if err := client.MonitoringServers.GenerateAndReloadAll(ctx); err != nil {
		log.Fatal(err)
	}
}
