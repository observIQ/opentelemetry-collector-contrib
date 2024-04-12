// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver

// func TestGetComputes(t *testing.T) {
// 	simulator.Test(func(ctx context.Context, c *vim25.Client) {
// 		finder := find.NewFinder(c)
// 		client := vcenterClient{
// 			vimDriver: c,
// 			finder:    finder,
// 		}
// 		dc, err := finder.DefaultDatacenter(ctx)
// 		require.NoError(t, err)
// 		computes, err := client.ComputeResources(ctx, nil)
// 		require.NoError(t, err)
// 		require.NotEmpty(t, computes, 0)
// 	})
// }

// func TestGetResourcePools(t *testing.T) {
// 	simulator.Test(func(ctx context.Context, c *vim25.Client) {
// 		finder := find.NewFinder(c)
// 		client := vcenterClient{
// 			vimDriver: c,
// 			finder:    finder,
// 		}
// 		resourcePools, err := client.ResourcePools(ctx, nil)
// 		require.NoError(t, err)
// 		require.NotEmpty(t, resourcePools)
// 	})
// }

// func TestGetVMs(t *testing.T) {
// 	simulator.Test(func(ctx context.Context, c *vim25.Client) {
// 		viewManager := view.NewManager(c)
// 		finder := find.NewFinder(c)
// 		client := vcenterClient{
// 			vimDriver: c,
// 			finder:    finder,
// 			vm:        viewManager,
// 		}
// 		vms, err := client.VMs(ctx, nil)
// 		require.NoError(t, err)
// 		require.NotEmpty(t, vms)
// 	})
// }

// func TestSessionReestablish(t *testing.T) {
// 	simulator.Test(func(ctx context.Context, c *vim25.Client) {
// 		sm := session.NewManager(c)
// 		moClient := &govmomi.Client{
// 			Client:         c,
// 			SessionManager: sm,
// 		}
// 		pw, _ := simulator.DefaultLogin.Password()
// 		client := vcenterClient{
// 			vimDriver: c,
// 			cfg: &Config{
// 				Username: simulator.DefaultLogin.Username(),
// 				Password: configopaque.String(pw),
// 				Endpoint: fmt.Sprintf("%s://%s", c.URL().Scheme, c.URL().Host),
// 				ClientConfig: configtls.ClientConfig{
// 					Insecure: true,
// 				},
// 			},
// 			moClient: moClient,
// 		}
// 		err := sm.Logout(ctx)
// 		require.NoError(t, err)

// 		connected, err := client.moClient.SessionManager.SessionIsActive(ctx)
// 		require.NoError(t, err)
// 		require.False(t, connected)

// 		err = client.EnsureConnection(ctx)
// 		require.NoError(t, err)

// 		connected, err = client.moClient.SessionManager.SessionIsActive(ctx)
// 		require.NoError(t, err)
// 		require.True(t, connected)
// 	})
// }
