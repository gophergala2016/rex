package room

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/bmatsuo/mdns"
)

// ServerDisco is a running instance of Room accesable at Addr.
type ServerDisco struct {
	Room    *Room
	TCPAddr *net.TCPAddr
	Entry   *mdns.ServiceEntry
}

// LookupRoom finds server applications with rooms that look like r.
// LookupRoom ignores the instance name of advertised services and relies only
// on the service identifier.
//
// BUG?  Not sure how mdns lookup handled channels when an error is
// encountered.
func LookupRoom(r *Room, servers chan<- *ServerDisco) error {
	c := make(chan *mdns.ServiceEntry)
	go func() {
		for entry := range c {
			var ip net.IP
			if entry.AddrV4 != nil {
				ip = entry.AddrV4
			} else if entry.AddrV6 != nil {
				ip = entry.AddrV6
			}

			tcpaddr := &net.TCPAddr{
				IP:   ip,
				Port: entry.Port,
			}
			addr := &ServerDisco{
				Room:    r,
				TCPAddr: tcpaddr,
				Entry:   entry,
			}

			servers <- addr
		}
	}()
	return mdns.Lookup(r.Service, c)
}

// ZoneConfig configures mDNS for a Room.
type ZoneConfig struct {
	Room *Room
	Port int
	IPs  []net.IP
	TXT  []string
}

// NewZoneConfig returns a default mDNS zone configuration derived from s.
func NewZoneConfig(s *Server) (*ZoneConfig, error) {
	zc := &ZoneConfig{
		Room: s.config.Room,
	}

	addr := s.Addr()
	if addr == "" {
		return nil, fmt.Errorf("server not bound to a port")
	}
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if host != "" {
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("invalid host ip: %v", err)
		}
		zc.IPs = []net.IP{ip}
	}
	zc.Port, err = strconv.Atoi(_port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return zc, nil
}

// Instance returns the mdns instance identifier corresponding to zc.Room.Name.
func (zc *ZoneConfig) Instance() string {
	return fmt.Sprintf("%s_%d_%s", time.Now(), os.Getpid(), zc.Room.Name)
}

func (zc *ZoneConfig) mdnsService() *mdns.MDNSService {
	return &mdns.MDNSService{
		Instance: zc.Instance(),
		Service:  zc.Room.Service,
		Port:     zc.Port,
		IPs:      zc.IPs,
		TXT:      zc.TXT,
	}
}

func (zc *ZoneConfig) mdnsConfig(iface *net.Interface) *mdns.Config {
	return &mdns.Config{
		Zone:  zc.mdnsService(),
		Iface: iface,
	}
}

// Discovery is an opaque type that contains an mDNS discovery server.
type Discovery interface {
	Close() error
	discoveryServer()
}

// DiscoveryServer returns a new Discovery server that is advertizing the Room
// in zc.
func DiscoveryServer(zc *ZoneConfig) (Discovery, error) {
	config := zc.mdnsConfig(nil)
	srv, err := mdns.NewServer(config)
	if err != nil {
		return nil, err
	}
	d := &mdnsDiscovery{srv: srv}
	return d, nil
}

type mdnsDiscovery struct {
	srv *mdns.Server
}

var _ Discovery = &mdnsDiscovery{}

func (d *mdnsDiscovery) Close() error {
	defer func() { d.srv = nil }()
	return d.srv.Shutdown()
}

func (d *mdnsDiscovery) discoveryServer() {
}
