# Mason

![Mason mascot](/internal/static/static/image/mason_gopher_small.jpeg)

The network toolbox

---

Mason is a network discovery and monitoring tool.  It includes common network tools such as ping, traceroute, and snmp data fetching providing visibility of networks.  Discovered devices are reguarly monitored by ping with historical data graphed for analysis.

---

## Quickstart

1. Download pre-built binary (coming soon)
2. Grant admin network permissions to the binary (instead of running mason as root): `sudo ./mason sys setcap`
3. Start Mason as a server: `./mason server`
4. Open a browser at http://localhost:4380/

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
- Device monitoring
    - Ping requests on regular intervals with recording of response time statistics
    - Different monitoring intervals for servers vs. client devices
- Charting of ping response times over time

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
