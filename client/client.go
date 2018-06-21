package client

import (
	"net"

	"github.com/songgao/water"
	"golang.org/x/net/ipv4"
	"../logging"
	"../config"
)

const (
	UDP_PACKET_SIZE		= 1500
)

type Client struct {
	iface		*water.Interface
	cfg		config.Client
	remoteConn	*net.UDPConn
	stopped		bool
	stopChan	chan bool
}

func New(iface *water.Interface, cfg config.Client) (*Client, error) {
	client := &Client{
		iface:		iface,
		cfg:		cfg,
		stopChan:	make(chan bool),
	}

	return client, nil
}

func (this *Client) startTun() error {
	log := logging.For("client")
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
			log.Debug("tun read packet  ", header)
			err = this.send(packet[:n])
			if err != nil {
				if !this.stopped {
					log.Error(err)
				}
				this.Stop()
				return

			}
		}
	}()

	return nil
}

func (this *Client) Start() error {
	log := logging.For("client")
	remoteAddr, err := net.ResolveUDPAddr("udp", this.cfg.Remote)
	if err != nil {
		log.Error("Error ResolveUDPAddr ", err)
		return err
	}

	remoteConn, err := net.DialUDP("udp", nil, remoteAddr)

	if err != nil {
		log.Error("Error coonnecting remote ", err)
		return err
	}

	this.remoteConn = remoteConn

	this.stopped = false

	this.startTun()

	go func() {
		buf := make([]byte, UDP_PACKET_SIZE)
		for {
			n, _, err := this.remoteConn.ReadFromUDP(buf)
			if err != nil {
				if !this.stopped {
					log.Error("Error reading from remote", err)
				}
				this.Stop()
				return
			}
			this.iface.Write(buf[:n])
		}
	}()

	for {
		select {
		case <-this.stopChan:
			this.stopped = true
			log.Info("closing remote connection: ", remoteAddr)
			this.remoteConn.Close()
			return nil
		}
	}

	return nil
}

func (this *Client) send(buf []byte) error {
	_, err := this.remoteConn.Write(buf)

	return err
}

func (this *Client) Stop(){
	select {
	case this.stopChan <- true:
	default:
	}
}
