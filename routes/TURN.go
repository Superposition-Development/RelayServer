package routes

import (
	db "RelayServer/database"
	"crypto/rand"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"net/http"

	"github.com/pion/turn/v5"
)

//RFC 6156

var (
	TURNUsername = rand.Text()
	TURNPassword = rand.Text()
)

func GetTURNCredentials(w http.ResponseWriter, r *http.Request) {
	_, err := db.AuthHeaderValidation(r)
	if err != nil {
		//do something
	}

	//w.encode the stuff, refer to getservers
}

func main() { //nolint:gocyclo,cyclop
	publicIP := "127.0.0.1" //this might need to be ipv6
	port := 3748
	TURNKnownKey := turn.GenerateAuthKey(TURNUsername, db.ServerConfig.Realm, TURNPassword)

	realm := db.ServerConfig.Realm

	if len(publicIP) == 0 {
		log.Fatalf("need public ip")
	}

	parsedIP := net.ParseIP(publicIP)
	if parsedIP == nil || parsedIP.To4() != nil {
		log.Fatalf("crap")
	}

	udpListener, err := net.ListenPacket("udp6", net.JoinHostPort("::", strconv.Itoa(port))) // nolint: noctx
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
	}

	server, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			if ra.Username == TURNUsername {
				return TURNUsername, TURNKnownKey, true
			}
			return "", nil, false
		},

		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: parsedIP,
					Address:      "::",
				},
			},
		},
	})
	if err != nil {
		log.Panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = server.Close(); err != nil {
		log.Panic(err)
	}
}
