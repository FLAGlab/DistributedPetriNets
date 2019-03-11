package communication

import (
	"flag"
	"fmt"
	"strconv"
	"time"

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

		if pn.isLeader {
			time.Sleep(2 * time.Second)
		}
		// acá solo se comunica con el peer que se acaba de inicializar
		go func() {
			for {
				fmt.Println("WILL PROCESS")
				if pn.isLeader {
					fmt.Println("LISTENING TO MESSAGES AS LEADER")
			    pn.pMsg <- (<-peer.Receive(opcodeChat)).(petriMessage)
				} else {
					fmt.Println("LISTENING TO MESSAGES AS FOLLOWER")
					pMsg := (<-peer.Receive(opcodeChat)).(petriMessage)
					fmt.Printf("RECEIVED: %v\n", pMsg)
					switch pMsg.Command {
					case TransitionCommand:
						fmt.Println("Type transition")
						transitionOptions := pn.petriNet.GetTransitionOptions()
						msgToSend := petriMessage{
							Command: TransitionCommand,
							Address: pn.node.ExternalAddress(),
							Transitions: transitionOptions}
						fmt.Printf("will send %v\n", msgToSend)
						peer.SendMessage(msgToSend)
						fmt.Printf("sent transitions: %v\n", transitionOptions)
					case FireCommand:
						transitionID := pMsg.Transitions[0].ID
						fmt.Printf("WILL FIRE transition with id: %v\n", transitionID)
						err := pn.petriNet.FireTransitionByID(transitionID)
						if err != nil {
							fmt.Println(err)
						}
					case PrintCommand:
						fmt.Println("WILL PRINT CURRENT PETRI NET")
						fmt.Printf("%v\n", pn.petriNet)
					default:
						fmt.Println("Unknown command")
					}
				}
				time.Sleep(2 * time.Second)
			}
		}()
		return nil
	})
}

// Run function that starts everything
func Run() {
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
	if *leaderFlag {
		pnNode.initLeader()
		pnNode.runLeader()
		defer pnNode.closeLeader()
	}
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
