package builder

import (
	"context"
	"fmt"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/client"

	"github.com/superfly/flyctl/flaps"
	"github.com/superfly/flyctl/internal/logger"
)

type Builder struct {
	Machine *api.Machine
	App     *api.AppCompact
	Client  *flaps.Client
}

func NewBuilder(ctx context.Context, orgSlug string) (builder *Builder, err error) {

	api := client.FromContext(ctx).API()

	org, err := api.GetOrganizationBySlug(ctx, orgSlug)

	if err != nil {
		return nil, err
	}

	builderApp := org.RemoteBuilderApp

	if builderApp == nil {
		_, builderApp, err = api.EnsureRemoteBuilder(ctx, org.ID)

		if err != nil {
			return nil, err
		}
	}

	flapsClient, err := flaps.New(ctx, builderApp)

	if err != nil {
		return
	}

	machines, err := flapsClient.List(ctx, "")

	if err != nil {
		return
	}

	if len(machines) < 1 {
		return nil, fmt.Errorf("builder app %s has no machines", org.RemoteBuilderApp.Name)
	}

	machine, err := flapsClient.Get(ctx, machines[0].ID)

	if err != nil {
		return
	}

	builder = &Builder{
		App:     builderApp,
		Machine: machine,
		Client:  flapsClient,
	}

	return
}

// Start will attempt to start and wait for a machine to be ready, regardless of its current state
func (b *Builder) Start(ctx context.Context) (err error) {
	logger := logger.FromContext(ctx)

	logger.Debugf("Starting builder instance %s for builder app %s", b.Machine.ID, b.App.Name)

	_, err = b.Client.Start(ctx, b.Machine.ID)

	if err != nil {
		return
	}

	logger.Debugf("Starting builder instance %s for builder app %s", b.Machine.ID, b.App.Name)

	err = b.Client.Wait(ctx, b.Machine)

	if err != nil {
		return
	}

	logger.Debugf("Builder instance %s for builder app %s  was started", b.Machine.ID, b.App.Name)

	return
}

// Wake sends a USR1 signal to a machine, which for remote builders means "stay alive"
func (b *Builder) Wake(ctx context.Context) (err error) {
	logger := logger.FromContext(ctx)

	logger.Debugf("Waking builder instance %s for builder app %s", b.Machine.ID, b.App.Name)

	err = b.Client.Wake(ctx, b.Machine.ID)

	return
}
