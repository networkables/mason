# Mason

![Mason mascot](/internal/static/static/images/mason_gopher_2_small.jpeg)

The network toolbox

---

Mason is a network discovery and monitoring tool.  It includes common network tools such as ping, traceroute, and snmp data fetching providing visibility of networks.  Discovered devices are reguarly monitored by ping with historical data graphed for analysis.

[![Go Report Card](https://goreportcard.com/badge/github.com/networkables/mason)](https://goreportcard.com/report/github.com/networkables/mason)

---

## Quickstart

### OS binary

1. Download pre-built binary [here](https://github.com/networkables/mason/releases)
2. Grant admin network permissions to the binary (instead of running mason as root): `sudo ./mason sys setcap`
3. Start Mason as a server: `./mason server`
4. Open your browser at http://localhost:4380/

### Docker

#### Trial without Persisting any data

```
docker run \
  --name mason \
  --rm \
  -e TZ="America/New_York" \
  -p "4380:4380" \
  -p "4322:4322" \
  -p "2055:2055/udp" \
  mason server
```

#### Persist data between runs

```
mkdir -p mason/data mason/config
chmod 0775 mason/data mason/config
sudo chgrp 65532 mason/data mason/config
docker run \
  --name mason \
  --rm \
  -u root \
  -e TZ="America/New_York" \
  -p "4380:4380" \
  -p "4322:4322" \
  -p "2055:2055/udp" \
  -v ./mason/data:/home/nonroot/data \
  -v ./mason/config:/home/nonroot/config \
  mason server
```

#### Persist and run privileged to allow traceroute

```
mkdir -p mason/data mason/config
docker run \
	--name mason \
  --rm \
  -u root \
  -e TZ="America/New_York" \
  -p "4380:4380" \
  -p "4322:4322" \
  -p "2055:2055/udp" \
  -v ./mason/data:/home/nonroot/data \
  -v ./mason/config:/home/nonroot/config \
  mason server --asn.enabled=true --oui.enabled=true --discovery.icmp.privileged=true
```

## Features

- Single binary with no external runtime dependencies
- Can be used as a server or a cli tool
- Multiple core networking tools 
    * Ping
    * Traceroute
    * SNMP
    * DNS Checks
    * TCP Port Scaning
    * TLS certificate information
- Default configuration designed to be productive on the initial run
- Core tools are additional exposed via command line and as network services
- Built in Web and Terminal UIs
- Low memory requirements ( 25-50 MB )
- Discovery Techniques
    * ARP Requests over address space for local LANs
    * Ping (ICMPv4) requests over address space for known/discovered networks
    * SNMP probes for ARP tables and network interfaces on discovered devices
    * Scans a /24 network in less than 60 seconds and a /16 clocks in around 15 minutes
- Device monitoring
    - Ping requests on regular intervals with recording of response time statistics
    - Different monitoring intervals for servers vs. client devices
- Charting of ping response times over time
- Use OUI data from ieee.org to find manufacturer of a device
    * Enable usage with __--oui.enabled=true__
- Use IP/ASN data from [https://github.com/sapics](https://github.com/sapics/ip-location-db/) to find Network/Country data
    * Enable usage with __--asn.enabled=true__
- IPFIX/Netflow listener to record in/out traffic flows of devices
    * See flows grouped by Network Organaization, Country, and IP

## Screenshots

### Dashboard / Start Page
![Mason Dashboard](https://github.com/networkables/images/raw/main/mason/screenshots/mason_dashboard.png)

### Device

Devices List
![Mason Devices Table](https://github.com/networkables/images/raw/main/mason/screenshots/mason_devices.png)

Device Detail
![Mason Device Detail](https://github.com/networkables/images/raw/main/mason/screenshots/mason_device_detail.png)

Device Detail (Ping Response Time)
![Mason Device Detail Ping](https://github.com/networkables/images/raw/main/mason/screenshots/mason_device_ping_focus.png)

Device Detail (Netflows by Org Name)
![Mason Device Detail Netflows OrgName](https://github.com/networkables/images/raw/main/mason/screenshots/mason_netflows_org.png)

### Tools

Ping
![Mason Tools Ping](https://github.com/networkables/images/raw/main/mason/screenshots/mason_tool_ping.png)

Traceroute
![Mason Tools Traceroute](https://github.com/networkables/images/raw/main/mason/screenshots/mason_tool_traceroute.png)

TLS Information
![Mason Tools TLS](https://github.com/networkables/images/raw/main/mason/screenshots/mason_tool_tls.png)

### Config

Mason ships with sane defaults you might want to customize to your needs.  You can use command line switches, environment variables, or a yaml file.  The default location of the config file is __config/config.yaml__ and you can change the directory using the __--config.directory__ command line switch or __MASON_CONFIG_DIRECTORY__ environment variable.

This is a full config file showing all the default values.  Customizations via config file only need to include what values you wish to modify (you do not have to duplicate every configuration value)
```
asn:
    asnurl: https://github.com/sapics/ip-location-db/raw/main/asn/asn-ipv4.csv
    cachefilename: cache.mpz1
    countryurl: https://github.com/sapics/ip-location-db/raw/main/geo-whois-asn-country/geo-whois-asn-country-ipv4.csv
    directory: data/asn
    enabled: true
bus:
    enabledebuglog: true
    enableerrorlog: true
    inboundsize: 0
    maxerrors: 100
    maxevents: 100
    minimumprioritylevel: 20
config:
    directory: config
discovery:
    arp:
        enabled: false
        timeout: 50ms
    autodiscovernewnetworks: true
    bootstraponfirstrun: true
    checkinterval: 1h0m0s
    enabled: true
    icmp:
        enabled: true
        pingcount: 2
        privileged: false
        timeout: 100ms
    maxworkers: 2
    networkscaninterval: 24h0m0s
    snmp:
        arptablerescaninterval: 1h0m0s
        community:
            - public
        enabled: true
        interfacerescaninterval: 24h0m0s
        ports:
            - 161
        timeout: 100ms
enrichment:
    dns:
        enabled: true
    enabled: true
    maxworkers: 2
    oui:
        enabled: true
    portscan:
        defaultscaninterval: 168h0m0s
        enabled: true
        maxworkers: 2
        portlist: general
        serverscaninterval: 24h0m0s
        timeout: 20ms
    snmp:
        community:
            - public
        enabled: true
        ports:
            - 161
        timeout: 50ms
netflows:
    enabled: true
    listenaddress: :2055
    maxworkers: 1
    packetsize: 16384
oui:
    directory: data/oui
    enabled: true
    filename: oui.mpz1
    url: https://standards-oui.ieee.org/oui/oui.txt
pinger:
    checkinterval: 5m0s
    defaultinterval: 1h0m0s
    enabled: true
    maxworkers: 2
    pingcount: 3
    privileged: false
    serverinterval: 5m0s
    timeout: 100ms
store:
    combo:
        directory: data
        enabled: false
        wspretention: 10m:3d,1h:3w
    sqlite:
        connectionmaxidle: 1h0m0s
        connectionmaxlifetime: 1h0m0s
        directory: data
        enabled: true
        filename: mason.db
        maxidleconnections: 5
        maxopenconnections: 5
        url: ""
tui:
    enabled: true
    listenaddress: :4322
wui:
    enabled: true
    listenaddress: :4380
```

## Support

Mason is still in initial development and has plenty of rough edges.  If you find a bug or have an issue, please open an Github issue.

## nettools library

The nettools package contains the basic network tools so you can build your own network tooling

- Send and receive ARP requests
- DNS resolution
- Send and receive ICMP4 Echo requests
- TCP Port scanning for a target
- SNMP information retrival
- TLS certiicate fetching and details parsing
- Traceroute using ICMP4 to a target
