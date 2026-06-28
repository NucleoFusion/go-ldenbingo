package neto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type EventCallback func(interface{})

type NetoClient struct {
	address       string
	port          int
	conn          net.Conn
	connMutex     sync.Mutex
	clientGuid    string
	identityToken string
	isConnected   bool

	// Callback handlers
	handlers      map[string][]EventCallback
	handlersMutex sync.RWMutex

	// Lifecycle channels
	stopChan   chan struct{}
	writeMutex sync.Mutex

	// Connection callbacks
	OnConnected    func()
	OnDisconnected func(err error)
	OnKicked       func(reason string)
	OnStatus       func(status string)
	OnError        func(err error)

	lastKeepAlive time.Time
}

func NewNetoClient(address string, port int, identityToken string) *NetoClient {
	return &NetoClient{
		address:       address,
		port:          port,
		identityToken: identityToken,
		handlers:      make(map[string][]EventCallback),
	}
}

func (c *NetoClient) ClientGuid() string {
	return c.clientGuid
}

func (c *NetoClient) IsConnected() bool {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	return c.isConnected
}

// Connect starts the TCP connection and registers the client on the server.
func (c *NetoClient) Connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.isConnected {
		return errors.New("already connected")
	}

	target := net.JoinHostPort(c.address, fmt.Sprintf("%d", c.port))
	c.logStatus(fmt.Sprintf("Connecting to %s...", target))

	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	c.conn = conn
	c.isConnected = true
	c.stopChan = make(chan struct{})
	c.lastKeepAlive = time.Now()

	// Start reading from connection
	go c.readLoop()

	// Send registration packet
	c.logStatus("Sending registration handshake...")
	regPayload := ClientRegisterPayload{
		Message:       "hello",
		Version:       "1",
		IdentityToken: c.identityToken,
	}

	regPacket := &Packet{
		Type:    ClientRegister,
		Objects: []interface{}{regPayload},
	}

	err = c.sendPacketDirect(regPacket)
	if err != nil {
		c.conn.Close()
		c.isConnected = false
		return fmt.Errorf("failed to send registration packet: %w", err)
	}

	// Start watchdog to verify keepalives
	go c.watchdog()

	return nil
}

// Disconnect gracefully disconnects from the server.
func (c *NetoClient) Disconnect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if !c.isConnected {
		return nil
	}

	c.logStatus("Disconnecting...")
	disconnectPacket := &Packet{
		Type:    ClientDisconnect,
		Objects: []interface{}{},
	}
	_ = c.sendPacketDirect(disconnectPacket)

	c.closeConnection(nil)
	return nil
}

func (c *NetoClient) closeConnection(err error) {
	if !c.isConnected {
		return
	}
	c.isConnected = false
	c.conn.Close()
	close(c.stopChan)

	if c.OnDisconnected != nil {
		c.OnDisconnected(err)
	}
}

// Send wraps any object in the ObjectData envelope and transmits it to the server.
func (c *NetoClient) Send(obj interface{}) error {
	packet := &Packet{
		Type:    ObjectData,
		Objects: []interface{}{obj},
	}
	return c.SendPacket(packet)
}

func (c *NetoClient) SendPacket(p *Packet) error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if !c.isConnected {
		return errors.New("not connected")
	}

	return c.sendPacketDirect(p)
}

func (c *NetoClient) sendPacketDirect(p *Packet) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	data, err := msgpack.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	length := int32(len(data))
	err = binary.Write(c.conn, binary.LittleEndian, length)
	if err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write packet data: %w", err)
	}

	return nil
}

// On registers a callback function for a specific struct type.
func (c *NetoClient) On(ptr interface{}, cb EventCallback) {
	t := reflect.TypeOf(ptr)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName, ok := nameRegistry[t]
	if !ok {
		typeName = t.Name()
	}

	c.handlersMutex.Lock()
	defer c.handlersMutex.Unlock()
	c.handlers[typeName] = append(c.handlers[typeName], cb)
}

func (c *NetoClient) dispatch(obj interface{}) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName, ok := nameRegistry[t]
	if !ok {
		typeName = t.Name()
	}

	c.handlersMutex.RLock()
	callbacks, exists := c.handlers[typeName]
	c.handlersMutex.RUnlock()

	if exists {
		for _, cb := range callbacks {
			go cb(obj)
		}
	}
}

func (c *NetoClient) readLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			var length int32
			err := binary.Read(c.conn, binary.LittleEndian, &length)
			if err != nil {
				c.connMutex.Lock()
				c.closeConnection(err)
				c.connMutex.Unlock()
				return
			}

			payload := make([]byte, length)
			_, err = io.ReadFull(c.conn, payload)
			if err != nil {
				c.connMutex.Lock()
				c.closeConnection(err)
				c.connMutex.Unlock()
				return
			}

			fmt.Printf("[RAW PAYLOAD] len=%d hex=%x\n", len(payload), payload)

			decompressed, err := Decompress(payload)
			if err != nil {
				c.logError(fmt.Errorf("failed to decompress payload: %w", err))
				continue
			}

			var packet Packet
			err = msgpack.Unmarshal(decompressed, &packet)
			if err != nil {
				c.logError(fmt.Errorf("failed to unmarshal packet: %w", err))
				continue
			}

			switch packet.Type {
			case ServerRegisterAccepted:
				if len(packet.Objects) > 0 {
					payload, ok := packet.Objects[0].(ServerRegisterAcceptedPayload)
					if ok && payload.Message == "neto server" {
						c.clientGuid = payload.ClientGuid
						c.logStatus("Handshake completed successfully.")
						if c.OnConnected != nil {
							c.OnConnected()
						}
					}
				}
			case ServerRegisterDenied:
				var msg = "unknown reason"
				if len(packet.Objects) > 0 {
					payload, ok := packet.Objects[0].(ServerRegisterDeniedPayload)
					if ok {
						msg = payload.Message
					}
				}
				c.logError(fmt.Errorf("registration denied: %s", msg))
				c.connMutex.Lock()
				c.closeConnection(errors.New("registration denied: " + msg))
				c.connMutex.Unlock()
				return
			case ServerClientDropped:
				var msg = "unknown reason"
				if len(packet.Objects) > 0 {
					payload, ok := packet.Objects[0].(ServerKickedPayload)
					if ok {
						msg = payload.Reason
					}
				}
				if c.OnKicked != nil {
					c.OnKicked(msg)
				}
				c.connMutex.Lock()
				c.closeConnection(errors.New("kicked: " + msg))
				c.connMutex.Unlock()
				return
			case ServerShutdown:
				c.logStatus("Server shutting down.")
				c.connMutex.Lock()
				c.closeConnection(errors.New("server shutdown"))
				c.connMutex.Unlock()
				return
			case KeepAlive:
				c.lastKeepAlive = time.Now()
			case ObjectData:
				for _, obj := range packet.Objects {
					c.dispatch(obj)
				}
			}
		}
	}
}

func (c *NetoClient) watchdog() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.connMutex.Lock()
			if !c.isConnected {
				c.connMutex.Unlock()
				return
			}
			if time.Since(c.lastKeepAlive) > 15*time.Second {
				c.logError(errors.New("keepalive timeout: server silent for 15 seconds"))
				c.closeConnection(errors.New("keepalive timeout"))
				c.connMutex.Unlock()
				return
			}
			c.connMutex.Unlock()
		}
	}
}

func (c *NetoClient) logStatus(status string) {
	if c.OnStatus != nil {
		c.OnStatus(status)
	}
}

func (c *NetoClient) logError(err error) {
	if c.OnError != nil {
		c.OnError(err)
	}
}
