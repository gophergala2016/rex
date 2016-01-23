package room

import (
	"fmt"
	"log"
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

	params := mdns.DefaultParams(r.Service)
	params.Entries = c
	params.Timeout = time.Minute
	err := mdns.Query(params)
	if err != nil {
		return err
	}
	return nil
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
	if host != "" && host != "::" {
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
	now := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%d_%s", now, os.Getpid(), zc.Room.Name)
}

func (zc *ZoneConfig) mdnsService() (*mdns.MDNSService, error) {
	return mdns.NewMDNSService(
		zc.Instance(),
		zc.Room.Service,
		"",
		"",
		zc.Port,
		zc.IPs,
		zc.TXT,
	)
}

func (zc *ZoneConfig) mdnsConfig(iface *net.Interface) (*mdns.Config, error) {
	zone, err := zc.mdnsService()
	if err != nil {
		return nil, err
	}
	config := &mdns.Config{
		Zone:  zone,
		Iface: iface,
	}
	return config, nil
}

// Discovery is an opaque type that contains an mDNS discovery server.
type Discovery interface {
	Close() error
	discoveryServer()
}

// DiscoveryServer returns a new Discovery server that is advertizing the Room
// in zc.
func DiscoveryServer(zc *ZoneConfig) (Discovery, error) {
	config, err := zc.mdnsConfig(nil)
	if err != nil {
		return nil, fmt.Errorf("invalid discovery configuration: %v", err)
	}
	log.Printf("[INFO] discovery configuration: %v", config.Zone)
	srv, err := mdns.NewServer(config)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] discover server started")
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
