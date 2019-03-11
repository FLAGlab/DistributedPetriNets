package communication

import (
	"flag"
	"strconv"
	"time"
	"math/rand"

	"github.com/FLAGlab/DCoPN/petribuilder"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
)


/** ENTRY POINT **/
func setup(pn *petriNode) {
	opcodeChat := noise.RegisterMessage(noise.NextAvailableOpcode(), (*petriMessage)(nil))
	pn.node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		// init se llama cuando se conecta un nodo o se le hace dial
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)
			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %v has disconnected.",
				peer.RemoteIP().String()+":"+strconv.Itoa(int(peer.RemotePort())))
			return nil
		})

		if pn.nodeType == Leader {
			time.Sleep(500 * time.Millisecond)
		}
		// acá solo se comunica con el peer que se acaba de inicializar
		go func() {
			for {
				pn.pMsg <- (<-peer.Receive(opcodeChat)).(petriMessage)
			}
		}()
		return nil
	})
}

// Run function that starts everything
func Run() {
	rand.Seed(time.Now().UnixNano())
	//gob.Register(skademlia.ID{})
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3000, "port to listen for peers on")
	leaderFlag := flag.Bool("l", false, "is leader node")
	flag.Parse()

	params := noise.DefaultParams()
	//params.NAT = nat.NewPMP()
	params.Keys = skademlia.RandomKeys()
	params.Host = *hostFlag
	params.Port = uint16(*portFlag)

	node, err := noise.NewNode(params)
	if err != nil {
		panic(err)
	}
	defer node.Kill()

	pnNode := &petriNode{node: node, petriNet: petribuilder.BuildPetriNet()}
	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(node)
	pnNode.init(*leaderFlag)
	pnNode.run()
	defer pnNode.close()
	setup(pnNode)
	go node.Listen()

	log.Info().Msgf("Listening for peers on port %d.", node.ExternalPort())

	if len(flag.Args()) > 0 {
		for _, address := range flag.Args() {
			peer, err := node.Dial(address)
			if err != nil {
				panic(err)
			}

			skademlia.WaitUntilAuthenticated(peer)
		}

		peers := skademlia.FindNode(node, protocol.NodeID(node).(skademlia.ID), skademlia.BucketSize(), 8)
		log.Info().Msgf("Bootstrapped with peers: %+v", peers)
	}

	for {}
}
