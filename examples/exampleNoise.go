package main

import (
    "fmt"
	
    "github.com/perlin-network/noise"
    "github.com/perlin-network/noise/cipher/aead"
    "github.com/perlin-network/noise/handshake/ecdh"
    "github.com/perlin-network/noise/identity/ed25519"
    "github.com/perlin-network/noise/protocol"
    "github.com/perlin-network/noise/rpc"
    "github.com/perlin-network/noise/skademlia"
)

type chatMessage struct {
	text string
}

func (chatMessage) Read(reader payload.Reader) (noise.Message, error) {
	text, err := reader.ReadString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read chat msg")
	}

	return chatMessage{text: text}, nil
}

func (m chatMessage) Write() []byte {
	return payload.NewWriter(nil).WriteString(m.text).Bytes()
}

func main() {
    // Register message type to Noise.
    opcodeChatMessage := noise.RegisterMessage(noise.NextAvailableOpcode(), (*chatMessage)(nil))
    
    params := noise.DefaultParams()
    params.Keys = ed25519.Random()
    params.Port = uint16(3000)
    
    node, err := noise.NewNode(params)
    if err != nil {
        panic(err)
    }
    
    protocol.New().
    	Register(ecdh.New()).
    	Register(aead.New()).
    	Register(skademlia.New()).
    	Enforce(node)
    
    fmt.Printf("Listening for peers on port %d.\n", node.ExternalPort())
    
    go node.Listen()
    
    // Dial peer via TCP located at address 127.0.0.1:3001.
    peer, err := node.Dial("127.0.0.1:3001")
    if err != nil {
        panic(err)
    }
    
    // Wait until the peer has finished all cryptographic handshake procedures.
    skademlia.WaitUntilAuthenticated(peer)
    
    // Send a single chat message over the peer knowing that it's encrypted over the wire.
    err = peer.SendMessage(chatMessage{text: "Hello peer!"})
    if err != nil {
        panic(err)
    }
    
    // Receive and print out a single chat message back from our peer.
    fmt.Println(<-peer.Receive(opcodeChatMessage))
}

