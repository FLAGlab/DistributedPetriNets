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
	"log"
	"bytes"
)

/** DEFINE MESSAGES **/
var (
	opcodeChat noise.Opcode
	_          noise.Message = (*petriMessage)(nil)
)

type petriMessage struct {
	command string
	id protocol.ID 
	transtion int 
}

func (petriMessage) Read(reader payload.Reader) (noise.Message, error) {
	byts, err := reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read msg")
	}
	var petriMessage m
	dec := gob.NewDecoder(bytes.NewReader(byts))
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode msg")
	}

	return m, nil
}

func (m petriMessage) Write() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		log.Fatal(err)
	
	}
	// A probar
	//return buf
	return payload.NewWriter(nil).WriteBytes(buf).Bytes()
}

/** ENTRY POINT **/
func setup(node *noise.Node) {
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
				msg := <-peer.Receive(opcodeChat)
				log.Info().Msgf("[%s]: %s", protocol.PeerID(peer), msg.(petriMessage).command)
			}
		}()

		return nil
	})
}

func main() {
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3000, "port to listen for peers on")
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

	setup(node)
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

	reader := bufio.NewReader(os.Stdin)

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
				}

				peer.SendMessageAsync(petriMessage{command: txt, id: protocol.NodeID(node), transtion: x})
				x = x+1
		}
		//skademlia.BroadcastAsync(node, petriMessage{text: })
	}
}
