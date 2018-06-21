package server

import (
	"net"
	"sync"

	"github.com/songgao/water"
	"golang.org/x/net/ipv4"
	"../logging"
	"../config"
)

const (
	UDP_PACKET_SIZE		= 1500
	HEADER_LEN		= 20
)

type Server struct {
	iface		*water.Interface
	cfg		config.Server
	peers		peersMap
	serverConn	*net.UDPConn
	stopped		bool

	stopChan	chan bool

}

type peersMap struct {
	sync.RWMutex
	m map[string]*net.UDPAddr
}

func New(iface *water.Interface, cfg config.Server) (*Server, error) {
	server := &Server{
		iface:		iface,
		cfg:		cfg,
		stopChan:	make(chan bool),
		peers:		peersMap{m: make(map[string]*net.UDPAddr)},
	}

	return server, nil
}

func (this *Server) startTun() error {
	log := logging.For("server")
	go func(){
		packet := make([]byte, UDP_PACKET_SIZE)
		for {
			n, err := this.iface.Read(packet)
			if err != nil {
				log.Error(err)
				return
			}
			header, _ := ipv4.ParseHeader(packet[:n])
			if header.Version != 4 {
				continue
			}
			log.Debug("read from tun", header)
			this.peers.RLock()
			peerAddr, ok := this.peers.m[header.Src.String()]
			this.peers.RUnlock()
			if ok {
				log.Debug("write packet to ", peerAddr)
				_, err := this.serverConn.WriteToUDP(packet[:n], peerAddr)
				if err != nil {
					log.Error(err)
					if this.stopped {
						return
					}

				}
			} else {
				log.Debug("peer not found, packet dropped.")
			}

		}
	}()

	return nil
}

func (this *Server) Start() error {
	log := logging.For("server")
	log.Info("Starting server.")

	this.stopped = false
	this.listen()
	this.startTun()

	for {
		select {
		case <-this.stopChan:
			this.stopped = true
			this.serverConn.Close()
			return nil
		}
	}

	return nil
}

func (this *Server) listen() error {
	log := logging.For("server")

	listenAddr, err := net.ResolveUDPAddr("udp", this.cfg.Bind)
	if err != nil {
		log.Error("Error ResolveUDPAddr ", err)
		return err
	}

	serverConn, err := net.ListenUDP("udp", listenAddr)

	if err != nil {
		log.Error("Error listen ", err)
		return err
	}

	this.serverConn = serverConn

	go func() {
		buf := make([]byte, UDP_PACKET_SIZE)
		for {
			n, remoteAddr, err := this.serverConn.ReadFromUDP(buf)
			if err != nil {
				if this.stopped {
					return
				}
				continue
			}
			if n < HEADER_LEN {
				continue
			}
			header, _ := ipv4.ParseHeader(buf[:n])
			if header.Version == 0  && n == HEADER_LEN { // healthcheck
				this.serverConn.WriteToUDP(buf[:n], remoteAddr)
			}
			if header.Version != 4 {
				continue
			}
			this.peers.Lock()
			this.peers.m[header.Dst.String()] = remoteAddr
			this.peers.Unlock()
			this.iface.Write(buf[:n])
		}
	}()

	return nil
}

func (this *Server) Stop(){
	select {
	case this.stopChan <- true:
	default:
	}
}
