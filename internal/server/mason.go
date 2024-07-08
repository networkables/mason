// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package server provides the Mason application used for network discovery and management
package server

import (
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"runtime"
	"runtime/debug"
	"slices"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/emicklei/tre"

	"github.com/networkables/mason/internal/asn"
	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/discovery"
	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/netflows"
	"github.com/networkables/mason/internal/oui"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

type Mason struct {
	bus bus.Bus
	cfg *Config

	// datastores
	store     Storer
	flowstore NetflowStorer

	// Workers
	enrichmentWorker     *enrichment.Worker
	discoveryWorker      *discovery.Worker
	networkScannerWorker *discovery.NetworkScannerWorker
	pingerWorker         *pinger.Worker
	netflowsWorker       *netflows.Worker

	// status stuff
	currentNetworkScan *string
	busBackPressure    atomic.Int32
	enrichBackPressure atomic.Int32
}

func New(opts ...Option) *Mason {
	o := applyOptionsToDefault(opts...)
	m := &Mason{
		cfg:                o.cfg,
		currentNetworkScan: o.scanstatus,
		bus:                o.bus,
		store:              o.store,
		flowstore:          o.nfstore,
	}

	if o.cfg.Oui.Enabled {
		oui.Load(
			oui.WithUrl(o.cfg.Oui.Url),
			oui.WithDirectory(o.cfg.Oui.Directory),
			oui.WithFilename(o.cfg.Oui.Filename),
		)
	}

	if o.cfg.Asn.Enabled {
		asn.Load(
			asn.WithAsnUrl(o.cfg.Asn.AsnUrl),
			asn.WithCountryUrl(o.cfg.Asn.CountryUrl),
			asn.WithDirectory(o.cfg.Asn.Directory),
			asn.WithCacheFilename(o.cfg.Asn.CacheFilename),
			asn.WithStorer(o.nfstore),
		)
	}

	return m
}

func (m *Mason) createWorkerPools(ctx context.Context) {
	m.discoveryWorker = discovery.NewWorker(m.cfg.Discovery)
	m.networkScannerWorker = discovery.NewNetworkScannerWorker(
		m.currentNetworkScan,
		m.discoveryWorker.In,
	)
	m.enrichmentWorker = enrichment.NewWorker()
	m.pingerWorker = pinger.NewWorker(m.cfg.Pinger)
	if m.cfg.NetFlows.Enabled {
		if m.flowstore == nil {
			log.Fatal("netflows enabled, but flowstore is nil")
		}
		input := netflows.Listen(ctx, m.cfg.NetFlows)
		m.netflowsWorker = netflows.NewWorker(m.cfg.NetFlows, input)
	}
}

func (m *Mason) shutdown() {
	m.enrichmentWorker.Close()
	m.discoveryWorker.Close()
	m.networkScannerWorker.Close()
	m.pingerWorker.Close()
	if m.netflowsWorker != nil {
		m.netflowsWorker.Close()
	}
	m.store.Close()
}

func (m *Mason) publish(e bus.Event) {
	m.busBackPressure.Add(1)
	go func() {
		m.bus.Publish(e)
		m.busBackPressure.Add(-1)
	}()
}

func (m *Mason) Run(ctx context.Context) {
	m.createWorkerPools(ctx)

	// Mason Bus Listener
	busch := m.bus.AddListener()

	// Bus
	go m.bus.Run(ctx)

	// Setup timers (tickers) for regularly scheduled actions
	networkScanTrigger := time.NewTicker(m.cfg.Discovery.CheckInterval)
	pingerTrigger := time.NewTicker(m.cfg.Pinger.CheckInterval)
	snmpArpTableRescanTrigger := time.NewTicker(m.cfg.Discovery.Snmp.ArpTableRescanInterval)
	snmpInterfaceRescanTrigger := time.NewTicker(m.cfg.Discovery.Snmp.InterfaceRescanInterval)
	defer func() {
		networkScanTrigger.Stop()
		pingerTrigger.Stop()
		snmpArpTableRescanTrigger.Stop()
		snmpInterfaceRescanTrigger.Stop()
	}()

	// kick off the worker pools
	go m.discoveryWorker.Run(ctx, m.cfg.Discovery.MaxWorkers)
	go m.networkScannerWorker.Run(ctx)
	go m.enrichmentWorker.Run(ctx, m.cfg.Enrichment.MaxWorkers)
	go m.pingerWorker.Run(ctx, m.cfg.Pinger.MaxWorkers)
	if m.cfg.NetFlows.Enabled {
		go m.netflowsWorker.Run(ctx, m.cfg.NetFlows.MaxWorkers)
	}

	if m.store.CountNetworks(ctx) == 0 && m.cfg.Discovery.BootstrapOnFirstRun {
		go func() {
			log.Debug("bootstraping mason")
			m.bootstrapnetworks()
		}()
	}

	for {
		select {

		case <-ctx.Done():
			log.Info("mason shutdown begin")
			m.shutdown()
			return

			//
			//
			// Time Triggers
			//
			//
		case <-networkScanTrigger.C:
			if m.cfg.Discovery.Enabled {
				m.publish(model.ScanAllNetworksRequest{})
			}

		case <-pingerTrigger.C:
			if m.cfg.Pinger.Enabled {
				m.publish(pinger.PerfPingDevicesEvent{})
			}

		case <-snmpArpTableRescanTrigger.C:
			go func() {
				devs := m.store.GetFilteredDevices(ctx,
					discovery.SnmpArpTableRescanFilter(m.cfg.Discovery.Snmp),
				)
				for _, dev := range devs {
					m.publish(discovery.DiscoverDevicesFromSNMPDevice{Device: dev})
				}
			}()

		case <-snmpInterfaceRescanTrigger.C:
			go func() {
				devs := m.store.GetFilteredDevices(ctx,
					discovery.SnmpArpTableRescanFilter(m.cfg.Discovery.Snmp),
				)
				for _, dev := range devs {
					m.publish(discovery.DiscoverNetworksFromSNMPDevice{Device: dev})
				}
			}()

		//
		//
		// Permanent WorkerPool handling
		// - pool.C is the results channel
		// - pool.E is any errors encountered while performing worker
		//
		//
		case discoveredDevice := <-m.discoveryWorker.C:
			m.publish(discoveredDevice)

		case err := <-m.discoveryWorker.E:
			if !errors.Is(err, discovery.ErrNoDeviceDiscovered) {
				// log.Errorf("address scan %T: %s", err, err)
				m.publish(tre.New(err, "discovery worker error"))
			}

		case enrichedDevice := <-m.enrichmentWorker.C:
			_, err := m.store.UpdateDevice(ctx, enrichedDevice)
			if err != nil {
				// log.Error("enrich, update device", "error", err)
				m.publish(tre.New(err, "enriched device store update", "addr", enrichedDevice.Addr))
			}
			if enrichedDevice.SNMP.Community != "" {
				m.publish(discovery.DiscoverDevicesFromSNMPDevice{Device: enrichedDevice})
				m.publish(discovery.DiscoverNetworksFromSNMPDevice{Device: enrichedDevice})
			}

		case err := <-m.enrichmentWorker.E:
			m.publish(tre.New(err, "enrichmentworker error"))

		case <-m.networkScannerWorker.C:
		// nohting todo for networkscan output

		case err := <-m.networkScannerWorker.E:
			m.publish(tre.New(err, "networkscanner worker error"))

		case pingPerf := <-m.pingerWorker.C:
			_, err := m.store.UpdateDevice(ctx, pingPerf.Device)
			if err != nil {
				m.publish(tre.New(err, "update device to store", "addr", pingPerf.Device.Addr))
			}
			err = m.store.WritePerformancePing(
				ctx,
				pingPerf.Start,
				pingPerf.Device,
				pingPerf.Stats,
			)
			if err != nil {
				m.publish(tre.New(err, "write pinger point", "addr", pingPerf.Device.Addr))
			}
			m.publish(model.EventDeviceUpdated(pingPerf.Device))

		case err := <-m.pingerWorker.E:
			m.publish(tre.New(err, "pinger worker error"))

		case flows := <-m.netflowsWorker.C:
			go func() {
				var err error
				for idx, flow := range flows {
					srcasn := m.LookupIP(flow.SrcAddr)
					dstasn := m.LookupIP(flow.DstAddr)
					flows[idx].SrcASN = srcasn
					flows[idx].DstASN = dstasn
				}
				err = m.flowstore.AddNetflows(ctx, flows)
				if err != nil {
					m.publish(err)
				}
			}()

		case err := <-m.netflowsWorker.E:
			m.publish(tre.New(err, "netflows worker"))

			//
			//
			// Event bus dispatching
			//
			//
		case event := <-busch:
			switch event := event.(type) {
			//
			//
			// ONLINE EVENTS: These events are synchronus events and should not be backgrounded
			//  - events hear handle tasks like saving data to disk
			//  - they should run quick and anything should hold until they are done
			//
			//
			case model.EventDeviceDiscovered:
				// - try to add to ds
				d := model.Device(event)
				err := m.store.AddDevice(ctx, d)
				if err == nil {
					// - if new emit new device event
					m.publish(model.EventDeviceAdded(event))
					continue
				}
				if errors.Is(err, model.ErrDeviceExists) {
					enrich, err := m.store.UpdateDevice(ctx, d)
					if err == nil {
						if enrich {
							m.publish(
								enrichment.EnrichDeviceRequest{
									Device: model.Device(event),
									Fields: enrichment.DefaultEnrichmentFields(m.cfg.Enrichment),
								})
						}
						continue
					}
				}
				m.publish(tre.New(err, "adding discovered device"))

			case model.EventDeviceUpdated:
				enrich, err := m.store.UpdateDevice(ctx, model.Device(event))
				if err != nil {
					m.publish(tre.New(err, "storing updated device"))
				}
				if enrich {
					m.publish(
						enrichment.EnrichDeviceRequest{
							Device: model.Device(event),
							Fields: enrichment.DefaultEnrichmentFields(m.cfg.Enrichment),
						})
				}

				// NewDeviceEvent is to spawn off further tasks
				// keeping as an event (instead of just moving the code into DiscoveredDevice), so we can see history
			case model.EventDeviceAdded:
				if m.cfg.Enrichment.Enabled {
					m.publish(
						enrichment.EnrichDeviceRequest{
							Device: model.Device(event),
							Fields: enrichment.DefaultEnrichmentFields(m.cfg.Enrichment),
						})
				}

			case model.DiscoveredNetwork:
				err := m.store.AddNetwork(ctx, model.Network(event))
				if err != nil {
					if !errors.Is(err, model.ErrNetworkExists) {
						m.publish(tre.New(err, "adding discovered network"))
					}
					continue
				}
				m.publish(model.NetworkAddedEvent(event))
				if m.cfg.Discovery.AutoDiscoverNewNetworks {
					m.publish(model.ScanNetworkRequest(event))
				}

				//
				//
				// BATCH EVENTS: These events make take a long time and will use a gorountine like a queue
				// - these events could take a long time to complete
				// - If using like a queue (EnrichDevice for example), then the goroutine will exit
				//    soon as the item is placed on the queue
				//
				//

			case model.ScanNetworkRequest:
				network := model.Network(event)
				network.LastScan = time.Now()
				m.store.UpdateNetwork(ctx, network)
				go func() {
					select {
					case <-ctx.Done():
						return
					case m.networkScannerWorker.In <- network:
					}
				}()

			case model.ScanAllNetworksRequest:
				go func() {
					// do a filter on the network list based on last scan time
					networks := m.store.GetFilteredNetworks(ctx, discovery.NetworkRescanFilter(m.cfg.Discovery))
					for _, n := range networks {
						m.publish(model.ScanNetworkRequest(n))
					}
				}()

			// Ping all devices who need to be pinged again
			case pinger.PerfPingDevicesEvent:
				go func() {
					devices := m.store.GetFilteredDevices(ctx, pinger.PerformancePingerFilter(m.cfg.Pinger))
					for _, device := range devices {
						m.pingerWorker.In <- device
					}
				}()

			case enrichment.EnrichDeviceRequest:
				m.enrichBackPressure.Add(1)
				go func() {
					select {
					case <-ctx.Done():
						return
					case m.enrichmentWorker.In <- event:
						m.enrichBackPressure.Add(-1)
					}
				}()

			case enrichment.EnrichAllDevicesEvent:
				if m.cfg.Enrichment.Enabled {
					go func() {
						devices := m.store.GetFilteredDevices(ctx, enrichment.PortScannerFilter(m.cfg.Enrichment.PortScan))
						for _, d := range devices {
							m.publish(enrichment.EnrichDeviceRequest{Device: d, Fields: enrichment.EnrichmentFields(event)})
						}
					}()
				}

			case discovery.DiscoverNetworksFromSNMPDevice:
				go discoverNetworksFromSnmp(ctx, event, m.cfg.Enrichment.Snmp.Timeout, m.publish, m.AddNetworkByName)

			case discovery.DiscoverDevicesFromSNMPDevice:
				go discoverDevicesFromSnmp(ctx, event, m.cfg.Enrichment.Snmp.Timeout, m.publish)
			}
		}
	}
}

func discoverNetworksFromSnmp(
	ctx context.Context,
	event discovery.DiscoverNetworksFromSNMPDevice,
	timeout time.Duration,
	publish func(bus.Event),
	addNetworkByName func(context.Context, string, string, bool) error,
) {
	prefixes, err := nettools.SnmpGetInterfaces(ctx, event.Addr.Addr(),
		nettools.WithSnmpCommunity(event.SNMP.Community),
		nettools.WithSnmpPort(event.SNMP.Port),
		nettools.WithSnmpReplyTimeout(timeout),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrConnectionRefused) ||
			errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return
		}
		publish(tre.New(err, "snmp get interfaces", "addr", event.Addr))
		return
	}
	for _, prefix := range prefixes {
		err = addNetworkByName(ctx, prefix.String(), prefix.String(), true)
		if err != nil {
			publish(
				tre.New(
					err,
					"adding snmp discovered network",
					"addr",
					event.Addr,
					"network",
					prefix.String(),
				),
			)
		}
	}
	event.Device.SNMP.LastInterfacesScan = time.Now()
	if len(prefixes) > 0 {
		event.Device.SNMP.HasInterfaces = true
		event.Device.SetUpdated()
	}
	publish(model.EventDeviceUpdated(event.Device))
}

func discoverDevicesFromSnmp(
	ctx context.Context,
	event discovery.DiscoverDevicesFromSNMPDevice,
	timeout time.Duration,
	publish func(bus.Event),
) {
	arps, err := nettools.SnmpGetArpTable(ctx, event.Addr.Addr(),
		nettools.WithSnmpCommunity(event.SNMP.Community),
		nettools.WithSnmpPort(event.SNMP.Port),
		nettools.WithSnmpReplyTimeout(timeout),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrConnectionRefused) ||
			errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return
		}
		publish(tre.New(err, "snmp get arp table", "addr", event.Addr))
		return
	}
	for _, arp := range arps {
		publish(model.EventDeviceDiscovered{
			Addr:         model.AddrToModelAddr(arp.Addr),
			MAC:          model.HardwareAddrToMAC(arp.MAC),
			DiscoveredBy: discovery.SNMPArpDiscoverySource,
			DiscoveredAt: time.Now(),
		})
	}
	event.Device.SNMP.LastArpTableScan = time.Now()
	if len(arps) > 0 {
		event.Device.SNMP.HasArpTable = true
		event.Device.SetUpdated()
	}
	publish(model.EventDeviceUpdated(event.Device))
}

// AddNetwork is a helper function to introduce a new network into the system
func (m *Mason) AddNetworkByName(
	ctx context.Context,
	name string,
	prefix string,
	scannow bool,
) error {
	newnet, err := model.New(name, prefix)
	if err != nil {
		return err
	}
	m.publish(model.DiscoveredNetwork(newnet))
	return nil
}

func (m *Mason) bootstrapnetworks() {
	ctx := context.Background()
	ifaces, err := net.Interfaces()
	if err != nil {
		// log.Error("bootstrap interfaces fetch", "error", err)
		m.publish(err)
		return
	}
	log.Info("mason bootstrap from network interfaces", "ifcount", len(ifaces))
	for _, iface := range ifaces {
		if !model.IsUsefulInterface(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			m.publish(err)
			continue
		}

		for _, addr := range addrs {
			ns := addr.String()
			err = m.AddNetworkByName(ctx, ns, ns, m.cfg.Discovery.Enabled)
			if err != nil {
				m.publish(err)
			}
		}
	}
	log.Debug("bootstrap complete")
}

func (m *Mason) recordIfError(err error) {
	if err != nil && m.bus != nil {
		m.publish(err)
	}
}

func (m *Mason) ListNetworks(ctx context.Context) []model.Network {
	return m.store.ListNetworks(ctx)
}

func (m *Mason) CountNetworks(ctx context.Context) int {
	return m.store.CountNetworks(ctx)
}

func (m *Mason) AddNetwork(ctx context.Context, network model.Network) error {
	err := m.store.AddNetwork(ctx, network)
	m.recordIfError(err)
	return err
}

func (m *Mason) ListDevices(ctx context.Context) []model.Device {
	return m.store.ListDevices(ctx)
}

func (m *Mason) CountDevices(ctx context.Context) int {
	return m.store.CountDevices(ctx)
}

func (m *Mason) GetDeviceByAddr(ctx context.Context, addr model.Addr) (model.Device, error) {
	d, err := m.store.GetDeviceByAddr(ctx, addr)
	m.recordIfError(err)
	return d, err
}

func (m *Mason) ReadPerformancePings(
	ctx context.Context,
	device model.Device,
	duration time.Duration,
) ([]pinger.Point, error) {
	points, err := m.store.ReadPerformancePings(ctx, device, duration)
	m.recordIfError(err)
	return points, err
}

func (m *Mason) GetConfig() *Config {
	return m.cfg
}

func (m *Mason) StringToAddr(str string) (model.Addr, error) {
	addr, err := model.ParseAddr(str)
	if err != nil {
		m.recordIfError(err)
	}
	return addr, err
}

func (m *Mason) IcmpPingAddr(
	ctx context.Context,
	addr model.Addr,
	count int,
	timeout time.Duration,
	priviledged bool,
) (nettools.Icmp4EchoResponseStatistics, error) {
	responses, err := nettools.Icmp4Echo(
		ctx,
		addr.Addr(),
		nettools.I4EWithCount(count),
		nettools.I4EWithReadTimeout(timeout),
		nettools.I4EWithPrivileged(priviledged),
		nettools.I4EWithTTL(3),
	)
	m.recordIfError(err)
	stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
	return stats, err
}

func (m *Mason) IcmpPing(
	ctx context.Context,
	target string,
	count int,
	timeout time.Duration,
	priviledged bool,
) (stats nettools.Icmp4EchoResponseStatistics, err error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return stats, err
	}
	stats, err = m.IcmpPingAddr(
		ctx,
		addr,
		count,
		timeout,
		priviledged,
	)
	return stats, err
}

func (m *Mason) ArpPing(
	ctx context.Context,
	target string,
	timeout time.Duration,
) (model.MAC, error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return model.MAC{}, err
	}
	entry, err := nettools.FindHardwareAddrOf(
		ctx,
		addr.Addr(),
		nettools.WithArpReplyTimeout(timeout),
	)
	m.recordIfError(err)
	return model.HardwareAddrToMAC(entry.MAC), err
}

func (m *Mason) Portscan(
	ctx context.Context,
	target string,
	cfg *enrichment.PortScanConfig,
) ([]int, error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return nil, err
	}
	ports, err := nettools.ScanTcpPorts(ctx, addr.Addr(),
		nettools.WithPortscanReplyTimeout(cfg.Timeout),
		nettools.WithPortscanPortlistName(cfg.PortList),
		nettools.WithPortscanMaxworkers(cfg.MaxWorkers),
	)
	m.recordIfError(err)
	return ports, err
}

func (m *Mason) GetExternalAddr(ctx context.Context) (model.Addr, error) {
	addr, err := nettools.GetExternalAddr(ctx)
	m.recordIfError(err)
	return model.AddrToModelAddr(addr), err
}

func (m *Mason) Traceroute(
	ctx context.Context,
	target string,
) (stat []nettools.Icmp4EchoResponseStatistics, err error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return stat, err
	}
	return m.TracerouteAddr(ctx, addr)
}

func (m *Mason) TracerouteAddr(
	ctx context.Context,
	target model.Addr,
) (stats []nettools.Icmp4EchoResponseStatistics, err error) {
	if !m.cfg.Discovery.Icmp.Privileged {
		return nil, errors.New("cannot execute traceroute in unpriviledged mode")
	}
	respOfResp, err := nettools.Traceroute4(
		ctx,
		target.Addr(),
		nettools.I4EWithPrivileged(m.cfg.Discovery.Icmp.Privileged),
	)
	if err != nil {
		m.recordIfError(err)
		return stats, err
	}
	stats = make([]nettools.Icmp4EchoResponseStatistics, 0, len(respOfResp))
	for _, resp := range respOfResp {
		stats = append(stats,
			nettools.CalculateIcmp4EchoResponseStatistics(resp))
	}

	if m.cfg.Asn.Enabled {
		for idx, stat := range stats {
			if !stat.Peer.IsValid() {
				continue
			}
			asn := m.LookupIP(model.AddrToModelAddr(stat.Peer))
			if asn == "" {
				continue
			}

			asninfo, err := m.GetAsn(ctx, asn)
			if err != nil {
				m.recordIfError(err)
				return stats, err
			}
			stats[idx].Asn = asninfo.Asn
			stats[idx].OrgName = asninfo.Name
		}
	}
	return stats, err
}

func (m *Mason) FetchTLSInfo(ctx context.Context, target string) (nettools.TLS, error) {
	return nettools.FetchTLS(target)
}

func (m *Mason) FetchSNMPInfo(
	ctx context.Context,
	target string,
) (info nettools.SnmpInfo, err error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		m.recordIfError(err)
		return info, err
	}
	return m.FetchSNMPInfoAddr(ctx, addr)
}

func (m *Mason) FetchSNMPInfoAddr(
	ctx context.Context,
	target model.Addr,
) (info nettools.SnmpInfo, err error) {
	return nettools.FetchSNMPInfo(ctx, target.Addr())
}

func (m *Mason) CheckDNS(
	ctx context.Context,
	target string,
) (map[string]map[string][]netip.Addr, error) {
	return nettools.DNSCheckAllServers(ctx, target)
}

func (m *Mason) GetUserAgent() string {
	return nettools.GetUserAgent()
}

func (m *Mason) OuiLookup(mac net.HardwareAddr) string {
	return oui.Lookup(mac)
}

type MasonInternalsView struct {
	NumberOfGoProcs int

	NetworkStoreCount int
	DeviceStoreCount  int

	DiscoveryMaxWorkers    int
	EnrichmentMaxWorkers   int
	EnrichmentBackPressure int
	PortScanMaxWorkers     int
	PingerMaxWorkers       int

	AddressScanActive  int
	DeviceEnrichActive int
	PerfPingActive     int
	NetworkScanActive  int

	BusBackPressure int

	CurrentNetworkScan string
	Events             []bus.HistoricalEvent
	Errors             []bus.HistoricalError

	Memstats  runtime.MemStats
	Buildinfo debug.BuildInfo
}

func (m *Mason) GetInternalsSnapshot(ctx context.Context) MasonInternalsView {
	iv := MasonInternalsView{}

	iv.NumberOfGoProcs = runtime.NumGoroutine()

	iv.NetworkStoreCount = m.store.CountNetworks(ctx)
	iv.DeviceStoreCount = m.store.CountDevices(ctx)

	iv.DiscoveryMaxWorkers = m.cfg.Discovery.MaxWorkers
	iv.PingerMaxWorkers = m.cfg.Pinger.MaxWorkers
	iv.EnrichmentMaxWorkers = m.cfg.Enrichment.MaxWorkers
	iv.EnrichmentBackPressure = int(m.enrichBackPressure.Load())
	iv.PortScanMaxWorkers = m.cfg.Enrichment.PortScan.MaxWorkers
	iv.CurrentNetworkScan = *m.currentNetworkScan

	iv.AddressScanActive = m.discoveryWorker.Active()
	iv.DeviceEnrichActive = m.enrichmentWorker.Active()
	iv.PerfPingActive = m.pingerWorker.Active()
	iv.NetworkScanActive = m.networkScannerWorker.Active()

	iv.BusBackPressure = int(m.busBackPressure.Load())

	iv.Events = m.bus.History()
	slices.Reverse(iv.Events)
	iv.Errors = m.bus.Errors()
	slices.Reverse(iv.Errors)

	runtime.ReadMemStats(&iv.Memstats)
	bi, ok := debug.ReadBuildInfo()
	if ok {
		iv.Buildinfo = *bi
	}

	return iv
}

func (m *Mason) GetNetworkStats(ctx context.Context) []model.NetworkStats {
	return buildNetworkStats(m.store.ListNetworks(ctx), m.store.ListDevices(ctx))
}

func (m *Mason) PingFailures(ctx context.Context) []model.Device {
	pf := make([]model.Device, 0)
	for _, d := range m.ListDevices(ctx) {
		if d.PerformancePing.LastFailed {
			pf = append(pf, d)
		}
	}
	return pf
}

func (m *Mason) ServerDevices(ctx context.Context) []model.Device {
	sd := make([]model.Device, 0)
	for _, d := range m.ListDevices(ctx) {
		if d.IsServer() {
			sd = append(sd, d)
		}
	}
	return sd
}

func (m *Mason) FlowSummaryByIP(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByIP, error) {
	return m.flowstore.FlowSummaryByIP(ctx, addr)
}

func (m *Mason) FlowSummaryByName(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByName, error) {
	return m.flowstore.FlowSummaryByName(ctx, addr)
}

func (m *Mason) FlowSummaryByCountry(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByCountry, error) {
	return m.flowstore.FlowSummaryByCountry(ctx, addr)
}

func buildNetworkStats(
	networks []model.Network,
	devices []model.Device,
) (nss []model.NetworkStats) {
	nss = make([]model.NetworkStats, 0, len(networks))
	for _, nw := range networks {
		ns := model.NetworkStats{}
		ns.Network = nw
		ns.IPTotal = math.Pow(float64(2), float64(32-nw.Prefix.Bits()))
		if nw.Prefix.Is6() {
			ns.IPTotal = math.Pow(float64(2), float64(128-nw.Prefix.Bits()))
		}
		var totalavg, totalmax time.Duration
		for _, dv := range devices {
			if !nw.Contains(dv) {
				continue
			}
			if dv.PerformancePing.LastFailed {
				continue
			}
			ns.IPUsed++
			totalavg += dv.PerformancePing.Mean
			totalmax += dv.PerformancePing.Maximum
		}
		if ns.IPUsed > 0 {
			ns.AvgPing = totalavg / time.Duration(ns.IPUsed)
			ns.MaxPing = totalmax / time.Duration(ns.IPUsed)
		}
		nss = append(nss, ns)
	}
	return nss
}

func (m *Mason) LookupIP(addr model.Addr) string {
	// TODO: check if asn is enabled in config
	return asn.FindAsn(addr.Addr())
}

func (m *Mason) GetAsn(ctx context.Context, asn string) (model.Asn, error) {
	return m.flowstore.GetAsn(ctx, asn)
}
