package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"
)

type MsgHandler func(*Packet)

type Client struct {
	notifyLock sync.Mutex
	redirector map[int]MsgHandler
	subscriber []MsgHandler

	sequenceNumLock sync.Mutex
	sequenceNum     int
	connection      net.Conn
}

func Dial(address string, password string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	client := &Client{
		connection: conn,

		sequenceNum:     2,
		sequenceNumLock: sync.Mutex{},

		notifyLock: sync.Mutex{},
		redirector: make(map[int]MsgHandler),
		subscriber: make([]MsgHandler, 0),
	}

	go func() {
		for {
			err := client.routeMsg()
			if err != nil {
				log.Printf("Error reading rcon channel %v", err)
			}
		}
	}()

	accepted := false
	var wg sync.WaitGroup
	wg.Add(2)

	authID := client.nextID()
	authReg, err := client.registerRedirect(authID, func(p *Packet) {
		if p.Type == AuthResp {
			accepted = true
		}
		wg.Done()
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer authReg()
	authFailReg, err := client.registerRedirect(-1, func(p *Packet) {
		// auth really failed
		wg.Done()
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer authFailReg()

	if err := client.writeCommand(Auth, authID, password); err != nil {
		conn.Close()
		return nil, err
	}
	wg.Wait()

	if !accepted {
		conn.Close()
		return nil, errors.New("Failed to authenticate with password")
	}

	return client, nil
}

func (c *Client) Execute(command string) (string, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	commandID := c.nextID()
	followID := c.nextID()
	body := ""
	aggr, err := c.registerRedirect(commandID, func(p *Packet) {
		if p.Type == Resp {
			body += p.Body
		}
	})
	if err != nil {
		return "", err
	}
	defer aggr()
	term, err := c.registerRedirect(followID, func(p *Packet) {
		if p.Type == Resp {
			wg.Done()
		}
	})
	if err != nil {
		return "", err
	}
	defer term()
	if err := c.writeCommand(Exec, commandID, command); err != nil {
		return "", err
	}
	if err := c.writeCommand(Resp, followID, ""); err != nil {
		return "", err
	}
	return body, nil
}

func (c *Client) Subscribe(handler MsgHandler) {
	c.notifyLock.Lock()
	defer c.notifyLock.Unlock()
	c.subscriber = append(c.subscriber, handler)
}

func (c *Client) registerRedirect(id int, handler MsgHandler) (func(), error) {
	c.notifyLock.Lock()
	defer c.notifyLock.Unlock()
	if _, ok := c.redirector[id]; ok {
		return nil, errors.New("Redirector already written")
	}
	c.redirector[id] = handler
	return func() {
		c.notifyLock.Lock()
		defer c.notifyLock.Unlock()
		delete(c.redirector, id)
	}, nil
}

func (c *Client) nextID() int {
	c.sequenceNumLock.Lock()
	id := c.sequenceNum
	if id > 10000 {
		id = 2
	}
	c.sequenceNum = id + 1
	c.sequenceNumLock.Unlock()
	return id
}

func (c *Client) routeMsg() error {
	packet, err := c.readCommand()
	if err != nil {
		return err
	}
	log.Printf("Recieved %v", packet)
	c.notifyLock.Lock()
	handler, ok := c.redirector[packet.ID]
	if ok {
		go handler(packet)
	} else {
		for _, sub := range c.subscriber {
			go sub(packet)
		}
	}
	c.notifyLock.Unlock()
	return nil
}

type PacketType int

const (
	Auth     PacketType = 3
	Exec     PacketType = 2
	Resp     PacketType = 0
	AuthResp PacketType = 2
)

type Packet struct {
	Type PacketType
	ID   int
	Body string
}

func (c *Client) writeCommand(cmdType PacketType, id int, cmd string) error {
	buffer := bytes.NewBuffer(make([]byte, 0, 14+len(cmd)))

	binary.Write(buffer, binary.LittleEndian, int(10+len(cmd)))
	binary.Write(buffer, binary.LittleEndian, int(id))
	binary.Write(buffer, binary.LittleEndian, int(cmdType))
	buffer.WriteString(cmd) // no null terminator
	binary.Write(buffer, binary.LittleEndian, byte(0))
	binary.Write(buffer, binary.LittleEndian, byte(0))

	_, err := c.connection.Write(buffer.Bytes())
	return err
}

func fill(conn net.Conn, buff []byte) error {
	bufLen := len(buff)
	start := 0
	for start < bufLen {
		len, err := conn.Read(buff[start:])
		start += len
		if start < bufLen && err != nil {
			return err
		}
	}
	return nil
}

func readNum(arr []byte) int {
	buf := bytes.NewBuffer(arr)
	var val int
	binary.Read(buf, binary.LittleEndian, &val)
	return val
}

func readString(arr []byte) string {
	buf := bytes.NewBuffer(arr)
	data, _ := buf.ReadString(0x00)
	return data
}

func (c *Client) readCommand() (*Packet, error) {
	sizeArr := make([]byte, 4)
	if err := fill(c.connection, sizeArr); err != nil {
		return nil, err
	}

	size := readNum(sizeArr)

	dataArr := make([]byte, size)
	if err := fill(c.connection, dataArr); err != nil {
		return nil, err
	}

	packet := &Packet{
		ID:   readNum(dataArr[0:4]),
		Type: PacketType(readNum(dataArr[4:8])),
		Body: readString(dataArr[8:]),
	}

	return packet, nil
}
