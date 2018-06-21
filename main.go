
package main

import (
	"os"
	"os/exec"

	"github.com/songgao/water"

	"./client"
	"./server"
	"./logging"
	"./config"
	"./cmd"
)

func sh(script string) error {
	log := logging.For("main")
	log.Info("Running script ", script)

	cmd := exec.Command(script) 
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); nil != err {
		log.Error("running command err ", err);
		return err
	}

	return nil
}

func setupTun(script string) (*water.Interface, error) {
	log := logging.For("main")
	config := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(config)
	if err != nil {
		return nil, err
	}

	log.Info("Interface Name: ", ifce.Name())

	sh(script)

	return ifce, nil
}

func startClient(ifce *water.Interface, cfg config.Client) {
	log := logging.For("main")
	log.Info("Starting Client...")

	cli, _ := client.New(ifce, cfg)
	cli.Start()

}

func startServer(ifce *water.Interface, cfg config.Server) {
	log := logging.For("main")
	log.Info("Starting Server...")

	server, _ := server.New(ifce, cfg)
	server.Start()
}

func main() {
	cmd.Execute(func(cfg *config.Config){
		logging.Configure(cfg.Logging.Output, cfg.Logging.Level)

		ifce, _ := setupTun(cfg.Script)

		if cfg.Mode == "server" {
			startServer(ifce, cfg.Server)
		} else if cfg.Mode == "client" {
			startClient(ifce, cfg.Client)
		}

		//<-(chan string)(nil)
	})
}
