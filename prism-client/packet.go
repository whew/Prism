// Code for the prism tcp packet protocol
package main

import (
	//"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"encoding/hex"
	//"errors"
)

// PacketType :  enumeration for packet types
type PacketType int

// see above
const (
	Initial PacketType = 1
	Welcome PacketType = 2
	ServerDisconnect PacketType = 3
	ClientConnect PacketType = 5
	ClientDisconnect PacketType = 6
	GeneralMessage PacketType = 20
	Received PacketType = 255
)


// Packet : a byte buffer with some vars to aid preparing a packet to send
type Packet struct{
	data []byte
	seek int
}



// NewPacket : Packet constructor
func NewPacket(t PacketType) Packet {
	var newPacket Packet

	// If this is a received packet, don't change anything
	if t == Received{
		return newPacket
	}

	// Write the packet type
	newPacket.data = append(newPacket.data, uint8(t))

	return newPacket
}


// PrepInitial : prepare the "Inital" packet type
func (p *Packet) PrepInitial(username string) {
	// Write in the len of the username then the username
	p.data = append(p.data, uint8(len(username)))
	p.data = append(p.data, []byte(username)...)
}


// PrepWelcome : prepare the "Welcome" packet type
func (p *Packet) PrepWelcome(clients map[string]net.Conn) {
	// Length of map clients
	l := len(clients)

	// Write in the number of connected clients
	p.data = append(p.data, uint8(l))

	// Loop over connected clients in map clients and write them to the packet
	for username := range clients {
		// Write in the len of the username then the username
		p.data = append(p.data, uint8(len(username)))
		p.data = append(p.data, []byte(username)...)
	}
}

// PrepServerDisconnect : prepare the "ServerDisconnect" packet type
func (p *Packet) PrepServerDisconnect (code int, reason string) {
	// Write the disconnect reason code
	p.data = append(p.data, uint8(code))

	// Write the disconnect reason as a string
	p.data = append(p.data, uint8(len(reason)))
	p.data = append(p.data, []byte(reason)...)
}


// PrepClientConnect : prepare the ClientConnect packet type
func (p *Packet) PrepClientConnect(username string) {
	// Write in the len of the username then the username
	p.data = append(p.data, uint8(len(username)))
	p.data = append(p.data, []byte(username)...)
}


// PrepClientDisconnect : prepare the ClientDisconnect packet type
func (p *Packet) PrepClientDisconnect(username string) {
	// Write in the len of the username then the username
	p.data = append(p.data, uint8(len(username)))
	p.data = append(p.data, []byte(username)...)
}


// PrepGeneralMessage : prepare the "GeneralMessage" packet type
func (p *Packet) PrepGeneralMessage(username string, message []byte, encrypted bool) {
	// Write in the len of the username then the username
	p.data = append(p.data, uint8(len(username)))
	p.data = append(p.data, []byte(username)...)

	// Write in whether the message is encrypted
	buf := make([]byte, 21 - len(username))
	var i uint8
	if encrypted { i = 1 }
	p.data = append(p.data, buf...)
	p.data = append(p.data, i)

	// Write in the len of the message then the message
	p.data = append(p.data, uint8(len(message)))
	p.data = append(p.data, message...)
}


// PrintData : prints out packet data as an array of bytes (values 0-255)
func (p *Packet) PrintData() {
	fmt.Println( p.data)
}


// PrintDataHex : prints out packet data as a hex dump
func (p *Packet) PrintDataHex() {
	fmt.Println(hex.Dump(p.data))	// DEBUG
}


// Send : writes len(p.data) + p.data to tcp connection c
func (p *Packet) Send(c net.Conn) {
	// Prepend the length of p.data to outgoing buffer
	len := uint16(len(p.data))
	out := append(make([]byte, 2), p.data...)
	binary.BigEndian.PutUint16(out, len)

	c.Write(out)
}


// Broadcast : sends packets to all connections in map m
func (p *Packet) Broadcast(m map[string]net.Conn ) {
	for _, connection := range m {
		p.Send(connection)
	}
}


// ReadBytes : reads in a slice from p.data and updates p.seek, n must be positive .. Maybe add errs??
func (p *Packet) ReadBytes(n int) []byte {
	start := p.seek
	end := start + n
	output := p.data[start:end]

	p.seek = end

	return output
}


// ReadUint8 : output a uint8 from p.data and update the seek
func (p *Packet) ReadUint8() uint8 {
	output := p.data[p.seek]
	p.seek++

	return output
}


// ReadString : Read n bytes from p.data and output them as a string then update p.seek
func (p *Packet) ReadString(n int) string {
	if n < 1 {
		return ""
	}

	start := p.seek
	end := start + n
	output := p.data[start:end]

	p.seek = end

	return string(output)
}


// ReadBool : output a bool from p.data and update the seek
func (p *Packet) ReadBool() bool {
	var output bool

	b := p.data[p.seek]
	p.seek++

	if b == 1 {				// TODO, there is SOOO a better way
		output = true
	} else {
		output = false
	}

	return output
}

// ReadSocket : read in data from the tcp socket and return a new packet
func ReadSocket(connection net.Conn) (Packet, error) {
	p := NewPacket(Received)

	// Read in the packet size
	pSize := make([]byte, 2)
	_, err := connection.Read(pSize[0:])
	if err != nil {
		return p , err
	}

	len := binary.BigEndian.Uint16(pSize)

	// Read in len bytes (size of packet) from socket
	netData := make([]byte, len)
	_, err = connection.Read(netData[0:])
	if err != nil {
		return p , err
	}

	p.data = netData

	return p, nil

}