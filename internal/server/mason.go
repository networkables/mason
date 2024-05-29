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
	"time"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/discovery"
	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/network"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/stackerr"
	"github.com/networkables/mason/internal/workerpool"
	"github.com/networkables/mason/nettools"
)

type (
	MasonReaderWriter interface {
		MasonReader
		MasonWriter
		MasonNetworker
	}
)

type MasonReader interface {
	ListNetworks() []network.Network
	CountNetworks() int
	ListDevices() []device.Device
	CountDevices() int
	GetDeviceByAddr(netip.Addr) (device.Device, error)
	ReadTimeseriesPoints(
		context.Context,
		device.Device,
		time.Duration,
		pinger.TimeseriesPoint,
	) ([]pinger.TimeseriesPoint, error)
	GetConfig() config.Config
	GetInternalsSnapshot() MasonInternalsView
	GetUserAgent() string
	OuiLookup(mac net.HardwareAddr) (string, error)
	GetNetworkStats() []network.NetworkStats
	PingFailures() []device.Device
	ServerDevices() []device.Device
}

type MasonWriter interface {
	AddNetwork(network.Network) error
	AddNetworkByName(string, string, bool) error
}

type MasonNetworker interface {
	StringToAddr(string) (netip.Addr, error)
	IcmpPingAddr(
		context.Context,
		netip.Addr,
		int,
		time.Duration,
	) (nettools.Icmp4EchoResponseStatistics, error)
	IcmpPing(
		context.Context,
		string,
		int,
		time.Duration,
	) (nettools.Icmp4EchoResponseStatistics, error)
	ArpPing(context.Context, string, time.Duration) (net.HardwareAddr, error)
	Portscan(context.Context, string, *config.PortScannerConfig) ([]int, error)
	GetExternalAddr(ctx context.Context) (netip.Addr, error)
	Traceroute(context.Context, string) ([]nettools.Icmp4EchoResponseStatistics, error)
	TracerouteAddr(context.Context, netip.Addr) ([]nettools.Icmp4EchoResponseStatistics, error)
	FetchTLSInfo(context.Context, string) (nettools.TLS, error)
	FetchSNMPInfo(context.Context, string) (nettools.SnmpInfo, error)
	FetchSNMPInfoAddr(context.Context, netip.Addr) (nettools.SnmpInfo, error)
	// ScanNetworkPrefix
	// HostInformation
	// InstanceInfo (external ip, traceroute to 1.1.1.1 & 8.8.8.8, hostinfo on external ip)
	// SNMPWalk
}

var _ MasonReaderWriter = (*Mason)(nil)

type Mason struct {
	bus bus.Bus
	cfg *config.Config

	// datastores
	store StoreAll

	// "Queues"
	addressScanQ  chan netip.Addr
	deviceEnrichQ chan enrichment.EnrichDeviceRequest
	perfPingQ     chan device.Device
	networkScanQ  chan network.Network

	// worker pools
	addressScanWP  *workerpool.Pool[netip.Addr, device.EventDeviceDiscovered]
	deviceEnrichWP *workerpool.Pool[enrichment.EnrichDeviceRequest, device.Device]
	perfPingWP     *workerpool.Pool[device.Device, pinger.PerformancePingResponseEvent]
	networkScanWP  *workerpool.Pool[network.Network, string]

	// status stuff
	currentNetworkScan *string
}

func New(
	cfg *config.Config,
	b bus.Bus,
	store StoreAll,
) *Mason {
	scanstatus := ""
	m := &Mason{
		bus:                b,
		store:              store,
		cfg:                cfg,
		currentNetworkScan: &scanstatus,
	}
	return m
}

// NewTool returns a Mason instance meant for network tool usage (ping, traceroute, etc..) and not persisting data or operating as a server
func NewTool() *Mason { return New(config.GetConfig(), nil, nil) }

func (m *Mason) createWorkerPools(ctx context.Context) {
	// Address scan setup
	m.addressScanQ = make(chan netip.Addr)
	m.addressScanWP = workerpool.New(
		"discovery",
		m.addressScanQ,
		discovery.BuildAddrScannerFunc(discovery.BuildAddrScanners(m.cfg.Discovery)),
	)

	// Device enrichment setup
	m.deviceEnrichQ = make(chan enrichment.EnrichDeviceRequest)
	m.deviceEnrichWP = workerpool.New("enrichment", m.deviceEnrichQ, enrichment.EnrichDevice)

	// Performance ping setup
	m.perfPingQ = make(chan device.Device)
	m.perfPingWP = workerpool.New(
		"perfping",
		m.perfPingQ,
		pinger.BuildPingDevice(ctx, m.cfg.PerformancePinger),
	)

	// Network scan setup
	m.networkScanQ = make(chan network.Network)
	m.networkScanWP = workerpool.New(
		"networkscan",
		m.networkScanQ,
		discovery.BuildNetworkScanFunc(m.addressScanQ, m.currentNetworkScan),
	)
}

func (m *Mason) shutdown() {
	close(m.addressScanQ)
	close(m.deviceEnrichQ)
	close(m.perfPingQ)
	close(m.networkScanQ)
}

func (m *Mason) Run(ctx context.Context) {
	m.createWorkerPools(ctx)

	// Mason Bus Listener
	busch := m.bus.AddListener()

	// Bus
	go m.bus.Run(ctx)

	// Setup timers (tickers) for regularly scheduled actions
	networkScanTrigger := time.NewTicker(*m.cfg.Discovery.CheckInterval)
	performancePingTrigger := time.NewTicker(*m.cfg.PerformancePinger.CheckInterval)
	snmpArpTableRescanTrigger := time.NewTicker(*m.cfg.Discovery.Snmp.ArpTableRescanInterval)
	snmpInterfaceRescanTrigger := time.NewTicker(*m.cfg.Discovery.Snmp.InterfaceRescanInterval)
	defer func() {
		networkScanTrigger.Stop()
		performancePingTrigger.Stop()
		snmpArpTableRescanTrigger.Stop()
		snmpInterfaceRescanTrigger.Stop()
	}()

	// kick off the worker pools
	go m.addressScanWP.Run(ctx, *m.cfg.Discovery.MaxWorkers)
	go m.deviceEnrichWP.Run(ctx, *m.cfg.Enrichment.MaxWorkers)
	go m.perfPingWP.Run(ctx, *m.cfg.PerformancePinger.MaxWorkers)
	go m.networkScanWP.Run(ctx, *m.cfg.Discovery.MaxNetworkScanners)

	if m.store.CountNetworks() == 0 && *m.cfg.Discovery.BootstrapOnFirstRun {
		go func() {
			log.Debug("bootstraping mason")
			m.bootstrapnetworks()
		}()
	}

	for {
		select {

		case <-ctx.Done():
			log.Debug("mason shutdown begin")
			m.shutdown()
			return

			//
			//
			// Time Triggers
			//
			//
		case <-networkScanTrigger.C:
			if *m.cfg.Discovery.Enabled {
				go m.bus.Publish(network.ScanAllNetworksRequest{})
			}

		case <-performancePingTrigger.C:
			if *m.cfg.PerformancePinger.Enabled {
				go m.bus.Publish(pinger.PerfPingDevicesEvent{})
			}

		case <-snmpArpTableRescanTrigger.C:
			go func() {
				devs := m.store.GetFilteredDevices(
					discovery.SnmpArpTableRescanFilter(m.cfg.Discovery.Snmp),
				)
				for _, dev := range devs {
					m.bus.Publish(discovery.DiscoverDevicesFromSNMPDevice{Device: dev})
				}
			}()

		case <-snmpInterfaceRescanTrigger.C:
			go func() {
				devs := m.store.GetFilteredDevices(
					discovery.SnmpArpTableRescanFilter(m.cfg.Discovery.Snmp),
				)
				for _, dev := range devs {
					m.bus.Publish(discovery.DiscoverNetworksFromSNMPDevice{Device: dev})
				}
			}()

		//
		//
		// Permanent WorkerPool handling
		// - pool.C is the results channel
		// - pool.E is any errors encountered while performing worker
		//
		//
		case discoveredDevice := <-m.addressScanWP.C:
			go m.bus.Publish(discoveredDevice)

		case err := <-m.addressScanWP.E:
			if !errors.Is(err, discovery.ErrNoDeviceDiscovered) {
				// log.Errorf("address scan %T: %s", err, err)
				go m.bus.Publish(err)
			}

		case enrichedDevice := <-m.deviceEnrichWP.C:
			err := m.store.UpdateDevice(enrichedDevice)
			if err != nil {
				// log.Error("enrich, update device", "error", err)
				go m.bus.Publish(err)
			}
			if enrichedDevice.SNMP.Community != "" {
				go func() {
					m.bus.Publish(discovery.DiscoverDevicesFromSNMPDevice{Device: enrichedDevice})
					m.bus.Publish(discovery.DiscoverNetworksFromSNMPDevice{Device: enrichedDevice})
				}()
			}

		case err := <-m.deviceEnrichWP.E:
			go func() {
				// log.Error("enrich device", "error", err)
				m.bus.Publish(err)
			}()

		case <-m.networkScanWP.C:
		// nohting todo for networkscan output

		case err := <-m.networkScanWP.E:
			go func() {
				// log.Error("network scan", "error", err)
				m.bus.Publish(err)
			}()

		case pingPerf := <-m.perfPingWP.C:
			err := m.store.UpdateDevice(pingPerf.Device)
			if err != nil {
				go m.bus.Publish(stackerr.New(err))
			}
			err = m.store.WriteTimeseriesPoint(
				ctx,
				pingPerf.Start,
				pingPerf.Device,
				pingPerf.Points()...,
			)
			if err != nil {
				go m.bus.Publish(stackerr.New(err))
			}
			go m.bus.Publish(device.EventDeviceUpdated(pingPerf.Device))

		case err := <-m.perfPingWP.E:
			go func() {
				// log.Error("perf ping", "error", err)
				m.bus.Publish(err)
			}()

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
			case device.EventDeviceDiscovered:
				// - try to add to ds
				d := device.Device(event)
				err := m.store.AddDevice(d)
				if err == nil {
					// - if new emit new device event
					go m.bus.Publish(device.EventDeviceAdded(event))
					continue
				}
				if errors.Is(err, device.ErrDeviceExists) {
					err = m.store.UpdateDevice(d)
					if err == nil {
						continue
					}
				}
				go m.bus.Publish(stackerr.New(err))

			case device.EventDeviceUpdated:
				err := m.store.UpdateDevice(device.Device(event))
				if err != nil {
					go m.bus.Publish(stackerr.New(err))
				}

				// NewDeviceEvent is to spawn off further tasks
				// keeping as an event (instead of just moving the code into DiscoveredDevice), so we can see history
			case device.EventDeviceAdded:
				if *m.cfg.Enrichment.Enabled {
					go m.bus.Publish(
						enrichment.EnrichDeviceRequest{
							Device: device.Device(event),
							Fields: enrichment.DefaultEnrichmentFields(m.cfg.Enrichment),
						})
				}

			case network.DiscoveredNetwork:
				err := m.store.AddNetwork(network.Network(event))
				if err != nil && !errors.Is(err, network.ErrNetworkExists) {
					go m.bus.Publish(stackerr.New(err))
				}
				go m.bus.Publish(network.NetworkAddedEvent(event))
				if *m.cfg.Discovery.AutoDiscoverNewNetworks {
					go m.bus.Publish(network.ScanNetworkRequest(event))
				}

				//
				//
				// BATCH EVENTS: These events make take a long time and will use a gorountine like a queue
				// - these events could take a long time to complete
				// - If using like a queue (EnrichDevice for example), then the goroutine will exit
				//    soon as the item is placed on the queue
				//
				//

			case network.ScanNetworkRequest:
				network := network.Network(event)
				network.LastScan = time.Now()
				m.store.UpdateNetwork(network)
				go func() {
					m.networkScanQ <- network
				}()

			case network.ScanAllNetworksRequest:
				go func() {
					// do a filter on the network list based on last scan time
					networks := m.store.GetFilteredNetworks(discovery.NetworkRescanFilter(m.cfg.Discovery))
					for _, n := range networks {
						go m.bus.Publish(network.ScanNetworkRequest(n))
					}
				}()

			// Ping all devices who need to be pinged again
			case pinger.PerfPingDevicesEvent:
				go func() {
					devices := m.store.GetFilteredDevices(pinger.PerformancePingerFilter(m.cfg.PerformancePinger))
					for _, device := range devices {
						m.perfPingQ <- device
					}
				}()

			case enrichment.EnrichDeviceRequest:
				go func() {
					m.deviceEnrichQ <- event
				}()

			case enrichment.EnrichAllDevicesEvent:
				if *m.cfg.Enrichment.Enabled {
					go func() {
						devices := m.store.GetFilteredDevices(enrichment.PortScannerFilter(m.cfg.Enrichment.PortScanner))
						for _, d := range devices {
							m.bus.Publish(enrichment.EnrichDeviceRequest{Device: d, Fields: enrichment.EnrichmentFields(event)})
							time.Sleep(time.Second)
						}
					}()
				}

			case discovery.DiscoverNetworksFromSNMPDevice:
				go func() {
					prefixes, err := nettools.SnmpGetInterfaces(ctx, event.Addr,
						nettools.WithSnmpCommunity(event.SNMP.Community),
						nettools.WithSnmpPort(event.SNMP.Port),
						nettools.WithSnmpReplyTimeout(*m.cfg.Enrichment.SNMP.Timeout),
					)
					if err != nil {
						if errors.Is(err, nettools.ErrConnectionRefused) || errors.Is(err, nettools.ErrNoResponseFromRemote) {
							return
						}
						m.bus.Publish(stackerr.New(err))
						return
					}
					for _, prefix := range prefixes {
						err = m.AddNetworkByName(prefix.String(), prefix.String(), true)
						if err != nil {
							m.bus.Publish(stackerr.New(err))
						}
					}
					event.SNMP.LastInterfacesScan = time.Now()
					if len(prefixes) > 0 {
						event.SNMP.HasInterfaces = true
						event.SetUpdated()
					}
					m.bus.Publish(device.EventDeviceUpdated(event.Device))
				}()
			case discovery.DiscoverDevicesFromSNMPDevice:
				go func() {
					arps, err := nettools.SnmpGetArpTable(ctx, event.Addr,
						nettools.WithSnmpCommunity(event.SNMP.Community),
						nettools.WithSnmpPort(event.SNMP.Port),
						nettools.WithSnmpReplyTimeout(*m.cfg.Enrichment.SNMP.Timeout),
					)
					if err != nil {
						if errors.Is(err, nettools.ErrConnectionRefused) || errors.Is(err, nettools.ErrNoResponseFromRemote) {
							return
						}
						m.bus.Publish(stackerr.New(err))
						return
					}
					for _, arp := range arps {
						m.bus.Publish(device.EventDeviceDiscovered{
							Addr:         arp.Addr,
							MAC:          arp.MAC,
							DiscoveredBy: discovery.SNMPArpDiscoverySource,
							DiscoveredAt: time.Now(),
						})
					}
					event.SNMP.LastArpTableScan = time.Now()
					if len(arps) > 0 {
						event.SNMP.HasArpTable = true
						event.SetUpdated()
					}
					m.bus.Publish(device.EventDeviceUpdated(event.Device))
				}()
			}
		}
	}
}

// AddNetwork is a helper function to introduce a new network into the system
func (m *Mason) AddNetworkByName(name string, prefix string, scannow bool) error {
	newnet, err := network.New(name, prefix)
	if err != nil {
		return err
	}
	m.bus.Publish(network.DiscoveredNetwork(newnet))
	return nil
}

func (m *Mason) bootstrapnetworks() {
	ifaces, err := net.Interfaces()
	if err != nil {
		// log.Error("bootstrap interfaces fetch", "error", err)
		m.bus.Publish(err)
		return
	}
	log.Info("mason bootstrap from network interfaces", "ifcount", len(ifaces))
	for _, iface := range ifaces {
		if !network.IsUsefulInterface(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			// log.Error("bootstrap interface addresses fetch", "error", err)
			m.bus.Publish(err)
			continue
		}

		for _, addr := range addrs {
			ns := addr.String()
			err = m.AddNetworkByName(ns, ns, *m.cfg.Discovery.Enabled)
			if err != nil {
				m.bus.Publish(err)
			}
		}
	}
	log.Debug("bootstrap complete")
}

func (m *Mason) recordIfError(err error) {
	if err != nil && m.bus != nil {
		m.bus.Publish(err)
	}
}

func (m *Mason) ListNetworks() []network.Network {
	return m.store.ListNetworks()
}

func (m *Mason) CountNetworks() int {
	return m.store.CountNetworks()
}

func (m *Mason) AddNetwork(network network.Network) error {
	err := m.store.AddNetwork(network)
	m.recordIfError(err)
	return err
}

func (m *Mason) ListDevices() []device.Device {
	return m.store.ListDevices()
}

func (m *Mason) CountDevices() int {
	return m.store.CountDevices()
}

func (m *Mason) GetDeviceByAddr(addr netip.Addr) (device.Device, error) {
	d, err := m.store.GetDeviceByAddr(addr)
	m.recordIfError(err)
	return d, err
}

func (m *Mason) ReadTimeseriesPoints(
	ctx context.Context,
	device device.Device,
	duration time.Duration,
	hydrator pinger.TimeseriesPoint,
) ([]pinger.TimeseriesPoint, error) {
	points, err := m.store.ReadTimeseriesPoints(ctx, device, duration, hydrator)
	m.recordIfError(err)
	return points, err
}

func (m Mason) GetConfig() config.Config {
	return *m.cfg
}

func (m Mason) StringToAddr(str string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(str)
	if err != nil {
		m.recordIfError(err)
	}
	return addr, err
}

func (m Mason) IcmpPingAddr(
	ctx context.Context,
	addr netip.Addr,
	count int,
	timeout time.Duration,
) (nettools.Icmp4EchoResponseStatistics, error) {
	responses, err := nettools.Icmp4Echo(
		ctx,
		addr,
		nettools.I4EWithCount(count),
		nettools.I4EWithReadTimeout(timeout),
	)
	m.recordIfError(err)
	stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
	return stats, err
}

func (m Mason) IcmpPing(
	ctx context.Context,
	target string,
	count int,
	timeout time.Duration,
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
	)
	return stats, err
}

func (m Mason) ArpPing(
	ctx context.Context,
	target string,
	timeout time.Duration,
) (net.HardwareAddr, error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return nil, err
	}
	entry, err := nettools.FindHardwareAddrOf(
		ctx,
		addr,
		nettools.WithArpReplyTimeout(timeout),
	)
	m.recordIfError(err)
	return entry.MAC, err
}

func (m Mason) Portscan(
	ctx context.Context,
	target string,
	cfg *config.PortScannerConfig,
) ([]int, error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return nil, err
	}
	ports, err := nettools.ScanTcpPorts(ctx, addr,
		nettools.WithPortscanReplyTimeout(*cfg.PortTimeout),
		nettools.WithPortscanPortlistName(*cfg.PortList),
		nettools.WithPortscanMaxworkers(*cfg.MaxWorkers),
	)
	m.recordIfError(err)
	return ports, err
}

func (m Mason) GetExternalAddr(ctx context.Context) (netip.Addr, error) {
	addr, err := nettools.GetExternalAddr(ctx)
	m.recordIfError(err)
	return addr, err
}

func (m Mason) Traceroute(
	ctx context.Context,
	target string,
) (stat []nettools.Icmp4EchoResponseStatistics, err error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return stat, err
	}
	return m.TracerouteAddr(ctx, addr)
}

func (m Mason) TracerouteAddr(
	ctx context.Context,
	target netip.Addr,
) (stats []nettools.Icmp4EchoResponseStatistics, err error) {
	respOfResp, err := nettools.Traceroute4(ctx, target)
	if err != nil {
		m.recordIfError(err)
		return stats, err
	}
	stats = make([]nettools.Icmp4EchoResponseStatistics, 0, len(respOfResp))
	for _, resp := range respOfResp {
		stats = append(stats,
			nettools.CalculateIcmp4EchoResponseStatistics(resp))
	}
	return stats, err
}

func (m Mason) FetchTLSInfo(ctx context.Context, target string) (nettools.TLS, error) {
	return nettools.FetchTLS(target)
}

func (m Mason) FetchSNMPInfo(
	ctx context.Context,
	target string,
) (info nettools.SnmpInfo, err error) {
	addr, err := m.StringToAddr(target)
	if err != nil {
		return info, err
	}
	return m.FetchSNMPInfoAddr(ctx, addr)
}

func (m Mason) FetchSNMPInfoAddr(
	ctx context.Context,
	target netip.Addr,
) (info nettools.SnmpInfo, err error) {
	return nettools.FetchSNMPInfo(ctx, target)
}

func (m Mason) CheckDNS(
	ctx context.Context,
	target string,
) (map[string]map[string][]netip.Addr, error) {
	return nettools.DNSCheckAllServers(ctx, target)
}

func (m Mason) GetUserAgent() string {
	return nettools.GetUserAgent()
}

func (m Mason) OuiLookup(mac net.HardwareAddr) (string, error) {
	x, err := nettools.OuiLookup(mac)
	m.recordIfError(err)
	return x, err
}

type MasonInternalsView struct {
	NumberOfGoProcs int

	NetworkStoreCount int
	DeviceStoreCount  int

	NetworkScanMaxWorkers int
	DiscoveryMaxWorkers   int
	EnrichmentMaxWorkers  int
	PortScanMaxWorkers    int
	PingerMaxWorkers      int

	AddressScanActive  int
	DeviceEnrichActive int
	PerfPingActive     int
	NetworkScanActive  int

	CurrentNetworkScan string
	Events             []bus.HistoricalEvent
	Errors             []bus.HistoricalError

	Memstats  runtime.MemStats
	Buildinfo debug.BuildInfo
}

func (m *Mason) GetInternalsSnapshot() MasonInternalsView {
	iv := MasonInternalsView{}

	iv.NumberOfGoProcs = runtime.NumGoroutine()

	iv.NetworkStoreCount = m.store.CountNetworks()
	iv.DeviceStoreCount = m.store.CountDevices()

	iv.NetworkScanMaxWorkers = *m.cfg.Discovery.MaxNetworkScanners
	iv.DiscoveryMaxWorkers = *m.cfg.Discovery.MaxWorkers
	iv.PingerMaxWorkers = *m.cfg.PerformancePinger.MaxWorkers
	iv.EnrichmentMaxWorkers = *m.cfg.Enrichment.MaxWorkers
	iv.PortScanMaxWorkers = *m.cfg.Enrichment.PortScanner.MaxWorkers
	iv.CurrentNetworkScan = *m.currentNetworkScan

	iv.AddressScanActive = m.addressScanWP.Active()
	iv.DeviceEnrichActive = m.deviceEnrichWP.Active()
	iv.PerfPingActive = m.perfPingWP.Active()
	iv.NetworkScanActive = m.networkScanWP.Active()

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

func (m Mason) GetNetworkStats() []network.NetworkStats {
	return buildNetworkStats(m.store.ListNetworks(), m.store.ListDevices())
}

func (m Mason) PingFailures() []device.Device {
	pf := make([]device.Device, 0)
	for _, d := range m.ListDevices() {
		if d.PerformancePing.LastFailed {
			pf = append(pf, d)
		}
	}
	return pf
}

func (m Mason) ServerDevices() []device.Device {
	sd := make([]device.Device, 0)
	for _, d := range m.ListDevices() {
		if d.IsServer() {
			sd = append(sd, d)
		}
	}
	return sd
}

func buildNetworkStats(
	networks []network.Network,
	devices []device.Device,
) (nss []network.NetworkStats) {
	nss = make([]network.NetworkStats, 0, len(networks))
	for _, nw := range networks {
		ns := network.NetworkStats{}
		ns.Network = nw
		ns.IPTotal = math.Pow(float64(2), float64(32-nw.Prefix.Bits()))
		if nw.Prefix.Addr().Is6() {
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
