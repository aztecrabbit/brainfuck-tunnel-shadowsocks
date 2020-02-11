package libshadowsocks

import (
	"os/exec"
	"bufio"
	"strings"
	"encoding/base64"

	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libutils"
	"github.com/aztecrabbit/libredsocks"
)

var (
	Loop = true
	DefaultConfig = &Config{
		Account: "ss://YWVzLTI1Ni1jZmI6TjNaNXFMM2E@sg1-socks.jagoanssh.com:1443?plugin=obfs-local%3Bobfs%3Dtls%3Bobfs-host%3Dtwitter.com#Jagoanssh",
		ServerNameIndication: "twitter.com",
	}
)

func Stop() {
	Loop = false
}

type Config struct {
	Account string
	ServerNameIndication string
}

type Shadowsocks struct {
	Redsocks *libredsocks.Redsocks
	Config *Config
	EncryptMethod string
	Password string
	Host string
	Port string
}

func (s *Shadowsocks) Start() {
	data := strings.Split(strings.Split(strings.Replace(s.Config.Account, "ss://", "", 1), "?")[0], "@")
	dataMethodPasswordDecode, err := base64.RawStdEncoding.DecodeString(data[0])
	if err != nil {
		panic(err)
	}
	dataMethodPassword := strings.Split(string(dataMethodPasswordDecode), ":")
	dataHostPort := strings.Split(data[1], ":")

	s.EncryptMethod = dataMethodPassword[0]
	s.Password = dataMethodPassword[1]
	s.Host = dataHostPort[0]
	s.Port = dataHostPort[1]

	for Loop {
		s.Redsocks.RuleDirectAdd(s.Host)

		command := exec.Command(
			"ss-local", "-v", "--fast-open", "--no-delay", "-l", "3080",
			"-s", s.Host,
			"-p", s.Port,
			"-k", s.Password,
			"-m", s.EncryptMethod,
			"--plugin", "obfs-local",
			"--plugin-opts", "obfs=tls;obfs-host=" + s.Config.ServerNameIndication + ";obfs-uri=/",
		)

		stderr, err := command.StdoutPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		go func() {
			for Loop && scanner.Scan() {
				line := strings.Join(strings.Split(scanner.Text(), " ")[4:], " ")

				if line == "running from root user" {
					liblog.LogInfo("Connected", "INFO", liblog.Colors["Y1"])

				} else if line == "Request did not begin with TLS handshake." ||
						strings.HasPrefix(line, "connection") ||
						strings.HasPrefix(line, "remote") {
					continue

				} else {
					liblog.LogInfo(line, "INFO", liblog.Colors["G2"])
				}
			}

			libutils.KillProcess(command.Process)
		}()

		command.Start()
		command.Wait()
	}
}
