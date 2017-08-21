package netlink_test

import (
	"log"

	"github.com/mdlayher/netlink"
)

// This example demonstrates using a netlink.Conn to execute requests against
// netlink.
func ExampleConn_execute() {
	// Speak to generic netlink using netlink
	const familyGeneric = 16

	c, err := netlink.Dial(familyGeneric, nil)
	if err != nil {
		log.Fatalf("failed to dial netlink: %v", err)
	}
	defer c.Close()

	// Ask netlink to send us an acknowledgement, which will contain
	// a copy of the header we sent to it
	req := netlink.Message{
		Header: netlink.Header{
			// Package netlink will automatically set header fields
			// which are set to zero
			Flags: netlink.HeaderFlagsRequest | netlink.HeaderFlagsAcknowledge,
		},
	}

	// Perform a request, receive replies, and validate the replies
	msgs, err := c.Execute(req)
	if err != nil {
		log.Fatalf("failed to execute request: %v", err)
	}

	if c := len(msgs); c != 1 {
		log.Fatalf("expected 1 message, but got: %d", c)
	}

	// Decode the copied request header, starting after 4 bytes
	// indicating "success"
	var res netlink.Message
	if err := (&res).UnmarshalBinary(msgs[0].Data[4:]); err != nil {
		log.Fatalf("failed to unmarshal response: %v", err)
	}

	log.Printf("res: %+v", res)
}

// This example demonstrates using a netlink.Conn to listen for multicast group
// messages generated by the addition and deletion of network interfaces.
func ExampleConn_listenMulticast() {
	const (
		// Speak to route netlink using netlink
		familyRoute = 0

		// Listen for events triggered by addition or deletion of
		// network interfaces
		rtmGroupLink = 0x1
	)

	c, err := netlink.Dial(familyRoute, &netlink.Config{
		// Groups is a bitmask; more than one group can be specified
		// by OR'ing multiple group values together
		Groups: rtmGroupLink,
	})
	if err != nil {
		log.Fatalf("failed to dial netlink: %v", err)
	}
	defer c.Close()

	for {
		// Listen for netlink messages triggered by multicast groups
		msgs, err := c.Receive()
		if err != nil {
			log.Fatalf("failed to receive messages: %v", err)
		}

		log.Printf("msgs: %+v", msgs)
	}
}