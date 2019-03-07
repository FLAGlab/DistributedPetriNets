package main

import (
	"bufio"
	"flag"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/payload"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
	"encoding/gob"
	"bytes"
	"github.com/FLAGlab/DCoPN/petrinet"
)

/** DEFINE MESSAGES **/
var (
	opcodeChat noise.Opcode
	_          noise.Message = (*petriMessage)(nil)
)

type petriNode struct {
	node noise.Node
	petriNet petrinet.PetriNet
	isLeader bool
}

type petriMessage struct {
	Command string
	Address string
	Transtion int
}

func (petriMessage) Read(reader payload.Reader) (noise.Message, error) {
	byts, err := reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read msg")
	}
	var m petriMessage
	dec := gob.NewDecoder(bytes.NewReader(byts))
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode msg")
	}

	return m, nil
}

func (m petriMessage) Write() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		log.Info().Msgf("Got a fucking error: %v", err)
	}
	// A probar
	//return buf.Bytes()
	return payload.NewWriter(nil).WriteBytes(buf.Bytes()).Bytes()
}
func buildPetriNet() *petrinet.PetriNet {
	/*p := petrinet.Init(1)
  	p.AddPlace(1, 1, "")
  	p.AddPlace(2, 1, "")
  	p.AddPlace(3, 2, "")
	p.AddPlace(4, 1, "")
	p.AddTransition(1,1)
	p.AddTransition(2,0)
	p.AddInArc(1,1,1)
	p.AddInArc(2,2,1)
	p.AddInArc(3,2,1)
	p.AddOutArc(1,2,1)
	p.AddOutArc(1,3,1)
	p.AddOutArc(2,4,1)
  //p.AddInhibitorArc(4,2,1)
	fmt.Printf("%v\n", p)*/
	return petrinet.Build()
}

func ask(node *noise.Node) {
	skademlia.BroadcastAsync(node, petriMessage{Command: "transitions", Address: node.ExternalAddress()})
}

func wait(opcode *noise.Opcode , pn *petrinet.PetriNet, node *noise.Node) []int {
	peers := skademlia.Table(node).GetPeers()
	log.Info().Msgf("peers: %v",peers)
	return make([]int, 5)
}

/** ENTRY POINT **/
func setup(node *noise.Node, pn *petrinet.PetriNet, leader bool) {
	opcodeChat = noise.RegisterMessage(noise.NextAvailableOpcode(), (*petriMessage)(nil))

	node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)

			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %v has disconnected.", peer.RemoteIP().String()+":"+strconv.Itoa(int(peer.RemotePort())))

			return nil
		})

		go func() {
			for {

				if leader {
					ask(node)
					transitions := wait(&opcodeChat,pn,node)
					//selectAndFire(transitions,node)
					log.Info().Msgf("a petri net: %v", pn)
					msg := <-peer.Receive(opcodeChat)
				} else {
					msg := <-peer.Receive(opcodeChat)
					log.Info().Msgf("[%s]: %s, %s", protocol.PeerID(peer), msg.(petriMessage).Command,msg.(petriMessage).Address)
						//recieve(msg,pn,node)
				}
				/*log.Info().Msgf("[%s]: %s, %s", protocol.PeerID(peer), msg.(petriMessage).Command,msg.(petriMessage).Address)
				if msg.(petriMessage).Transtion < 5 {
					peer, err := node.Dial(msg.(petriMessage).Address)
					if err != nil {
						panic(err)
					}
					peer.SendMessageAsync(petriMessage{Command: "jajajja", Address: node.ExternalAddress(), Transtion: msg.(petriMessage).Transtion+1 })
				}*/

			}
		}()

		return nil
	})
}

func main() {
	//gob.Register(skademlia.ID{})
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3000, "port to listen for peers on")
	leaderFlag := flag.Bool("l",false,"is leader node")
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

	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(node)
	setup(node,buildPetriNet(),*leaderFlag)
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

/*	reader := bufio.NewReader(os.Stdin)

	for {
		txt, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}
		x := 0
		txt = strings.TrimSpace(txt)
		for _, peerID := range skademlia.FindClosestPeers(skademlia.Table(node), protocol.NodeID(node).Hash(), skademlia.BucketSize()) {
				peer := protocol.Peer(node, peerID)

				if peer == nil {
					continue
*///				}

	//			peer.SendMessageAsync(petriMessage{Command: txt, Address: node.ExternalAddress(), Transtion: x})
	//				//peer.SendMessageAsync(petriMessage{Command: txt, Transtion: x})
	//			x = x+1
	//	}
		//skademlia.BroadcastAsync(node, petriMessage{text: })
	//}
}
