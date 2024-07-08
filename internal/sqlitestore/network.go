// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"slices"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

// AddNetwork adds a network to the store, will return error if the network already exists in the store
func (cs *Store) AddNetwork(ctx context.Context, newnetwork model.Network) error {
	for _, n := range cs.networks {
		if model.CompareNetwork(n, newnetwork) == 0 {
			return model.ErrNetworkExists
		}
	}
	cs.networks = append(cs.networks, newnetwork)
	return cs.saveNetworks(ctx)
}

// RemoveNetworkByName remove the named network from the store
func (cs *Store) RemoveNetworkByName(ctx context.Context, name string) error {
	for idx, n := range cs.networks {
		if n.Name == name {
			cs.networks = slices.Delete(cs.networks, idx, idx+1)
			return cs.saveNetworks(ctx)
		}
	}
	return model.ErrNetworkDoesNotExist
}

// UpdateNetwork will freshen up the network using the given network
func (cs *Store) UpdateNetwork(ctx context.Context, newnetwork model.Network) error {
	idx := 0
	found := false
	for i, network := range cs.networks {
		if model.CompareNetwork(network, newnetwork) == 0 {
			idx = i
			found = true
		}
	}
	if !found {
		return model.ErrNetworkDoesNotExist
	}
	cs.networks[idx] = newnetwork
	return cs.saveNetworks(ctx)
}

// UpsertNetwork will either add the given network and if it already exists then it will run an update
func (cs *Store) UpsertNetwork(ctx context.Context, n model.Network) error {
	_, err := cs.DB.NamedExecContext(
		ctx,
		`insert into networks (prefix, name, lastscan, tags)
    values (:prefix, :name, :lastscan, :tags)
    on conflict (prefix) do update set name=:name, lastscan=:lastscan, tags=:tags`,
		n,
	)
	return err
}

// GetNetworkByName returns a network with the given name
func (cs *Store) GetNetworkByName(ctx context.Context, name string) (model.Network, error) {
	for _, network := range cs.networks {
		if network.Name == name {
			return network, nil
		}
	}
	return model.Network{}, model.ErrNetworkDoesNotExist
}

// GetFilteredNetworks returns the networks which match the given GetFilteredNetworks
func (cs *Store) GetFilteredNetworks(
	ctx context.Context,
	filter model.NetworkFilter,
) []model.Network {
	networks := make([]model.Network, 0)
	for _, network := range cs.networks {
		if filter(network) {
			networks = append(networks, network)
		}
	}
	return networks
}

// ListNetworks returns all stored networks
func (cs *Store) ListNetworks(ctx context.Context) []model.Network {
	return slices.Clone(cs.networks)
}

// CountNetworks returns the number of networks in the store
func (cs *Store) CountNetworks(ctx context.Context) int {
	return len(cs.networks)
}

func (cs *Store) saveNetworks(ctx context.Context) (err error) {
	for _, network := range cs.networks {
		err = cs.UpsertNetwork(ctx, network)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *Store) readNetworksInitial(ctx context.Context) (err error) {
	err = cs.readNetworks(ctx)
	if err != nil && strings.EqualFold(err.Error(), "no such table: networks") {
		return nil
	}
	return err
}

func (cs *Store) readNetworks(ctx context.Context) (err error) {
	cs.networks, err = cs.selectNetworks(ctx)
	return err
}

func (cs *Store) selectNetworks(ctx context.Context) ([]model.Network, error) {
	var fs []model.Network
	err := cs.DB.SelectContext(
		ctx,
		&fs,
		`select name, prefix, lastscan, tags from networks`,
	)
	return fs, err
}
