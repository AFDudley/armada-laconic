package nitro

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/gogoproto/proto"
	nitrop2p "github.com/statechannels/go-nitro/node/engine/messageservice"
	"github.com/statechannels/go-nitro/protocols"
	nitrotypes "github.com/statechannels/go-nitro/types"
	"golang.org/x/time/rate"
)

const (
	P2PMessageChannel = byte(0x80)
	// SigRequestChannel = byte(0x81)

	maxMsgSize     = 1024 * 1024
	msgBufSize     = 1000
	limiterTimeout = 5 * time.Second
)

var (
	_ p2p.Reactor             = (*p2pReactor)(nil)
	_ nitrop2p.MessageService = (*p2pMsgService)(nil)
)

// p2pReactor handles Nitro P2P messages via CometBFT
type p2pReactor struct {
	p2p.BaseReactor
	incoming    chan nitrop2p.Message
	outgoing    map[p2p.ID]chan nitrop2p.Message
	sendQueue   chan nitrop2p.Message
	rateLimiter *rate.Limiter
	config      *Config
	logger      log.Logger
	mtx         sync.RWMutex
	wg          sync.WaitGroup
}

type p2pMsgService struct {
	id      string
	logger  log.Logger
	reactor *p2pReactor // one reactor per msg service
}

type streamDesc struct {
	id      byte
	msgType proto.Message
}

func (sd streamDesc) StreamID() byte { return sd.id }

// TODO: convert Message to protobuf
func (sd streamDesc) MessageType() proto.Message { return sd.msgType }

// Reactor

func newReactor(config *Config, logger log.Logger) *p2pReactor {
	var rateLimiter *rate.Limiter
	if config.P2PRateLimitEnable {
		rateLimiter = rate.NewLimiter(rate.Limit(config.P2PRateLimitRate), config.P2PRateLimitBurst)
	}

	ret := &p2pReactor{
		incoming:    make(chan nitrop2p.Message, msgBufSize),
		outgoing:    make(map[p2p.ID]chan nitrop2p.Message),
		sendQueue:   make(chan nitrop2p.Message, msgBufSize),
		rateLimiter: rateLimiter,
		config:      config,
		logger:      logger,
	}
	ret.BaseReactor = *p2p.NewBaseReactor("nitro-reactor", ret)
	return ret
}

// StreamDescriptors returns the stream descriptor for Nitro messages.
func (r *p2pReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  P2PMessageChannel,
			Priority:            5,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1000,
			RecvMessageCapacity: maxMsgSize,
			MessageType:         &P2PMessage{},
		},
		// TODO Nitro peer info channel
	}
}

// AddPeer begins sending messages to a peer.
func (r *p2pReactor) AddPeer(peer p2p.Peer) {
	if !r.IsRunning() {
		return
	}

	peerCh := make(chan nitrop2p.Message, msgBufSize)
	r.mtx.Lock()
	r.outgoing[peer.ID()] = peerCh
	r.mtx.Unlock()

	// Start goroutine to handle sending to this specific peer
	go r.handlePeer(peer, peerCh)
}

func (r *p2pReactor) handlePeer(peer p2p.Peer, peerCh chan nitrop2p.Message) {
	logger := r.logger.With("peer", peer)

	defer func() {
		r.mtx.Lock()
		defer r.mtx.Unlock()
		if _, ok := r.outgoing[peer.ID()]; ok {
			close(peerCh)
			delete(r.outgoing, peer.ID())
		}
	}()

	for {
		if !peer.IsRunning() || !r.IsRunning() {
			return
		}
	inner_loop:
		for {
			select {
			case msg := <-peerCh:
				encoded, err := msg.Serialize()
				if err != nil {
					logger.Error("Failed to serialize message", "msg", msg, "err", err)
					continue inner_loop
				}
				sent := peer.Send(p2p.Envelope{
					Message:   &P2PMessage{Content: encoded},
					ChannelID: P2PMessageChannel,
				})
				if !sent {
					logger.Error("Failed to send message", "msg", msg)
					return
				}
			default:
				break inner_loop
			}
		}
	}
}

// Receive handles an envelope received from any connected peer on any registered channel.
func (r *p2pReactor) Receive(e p2p.Envelope) {
	if !r.IsRunning() {
		r.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID)
		return
	}
	switch e.ChannelID {
	case P2PMessageChannel:
		switch msg := e.Message.(type) {
		case *P2PMessage:
			nitroMsg, err := protocols.DeserializeMessage(msg.Content)
			if err != nil {
				r.Logger.Error("Failed to deserialize message", "err", err)
				r.Switch.StopPeerForError(e.Src, err)
				return
			}
			r.incoming <- nitroMsg
		default:
			r.Logger.Error(fmt.Sprintf("Unknown message type: %T", e.Message))
		}
	default:
		r.Logger.Error(fmt.Sprintf("Unknown channel ID: %X", e.ChannelID))
	}
}

func (r *p2pReactor) OnStart() error {
	// Start the single sending goroutine
	r.wg.Add(1)
	go r.sendLoop()
	return nil
}

func (r *p2pReactor) OnStop() { r.close() }

func (r *p2pReactor) close() {
	close(r.incoming)
	close(r.sendQueue)
	r.wg.Wait() // Wait for sending goroutine to finish

	r.mtx.Lock()
	defer r.mtx.Unlock()
	for id, ch := range r.outgoing {
		close(ch)
		delete(r.outgoing, id)
	}
}

// sendLoop runs in a single goroutine and handles all outgoing messages with rate limiting
func (r *p2pReactor) sendLoop() {
	defer r.wg.Done()

	for msg := range r.sendQueue {
		// Apply rate limiting if enabled
		if r.rateLimiter != nil {
			ctx, cancel := context.WithTimeout(context.Background(), limiterTimeout)
			if err := r.rateLimiter.Wait(ctx); err != nil {
				r.logger.Error("Rate limit exceeded, dropping message", "error", err)
				cancel()
				continue
			}
			cancel()
		}

		// Send to all connected peers
		r.mtx.RLock()
		for peerID, ch := range r.outgoing {
			select {
			case ch <- msg:
				// Message sent successfully
			default:
				// Channel is full, log warning but don't block
				r.logger.Warn("Peer channel full, dropping message", "peer", peerID)
			}
		}
		r.mtx.RUnlock()
	}
}

func (r *p2pReactor) send(msg nitrop2p.Message) {
	select {
	case r.sendQueue <- msg:
		// Message queued successfully
	default:
		// Send queue is full, log error
		r.logger.Error("Send queue full, dropping message")
	}
}

// GetRateLimitStats returns current rate limiting statistics
func (r *p2pReactor) GetRateLimitStats() (enabled bool, limit float64, burst int, tokens float64) {
	if r.rateLimiter == nil {
		return false, 0, 0, 0
	}
	return true, float64(r.rateLimiter.Limit()), r.rateLimiter.Burst(), r.rateLimiter.Tokens()
}

// UpdateRateLimit updates the rate limiting parameters at runtime
func (r *p2pReactor) UpdateRateLimit(newRate float64, newBurst int) {
	if r.rateLimiter != nil {
		r.rateLimiter.SetLimit(rate.Limit(newRate))
		r.rateLimiter.SetBurst(newBurst)
	}
}

// MessageService

func newMessageService(id string, r *p2pReactor /* , logger log.Logger */) *p2pMsgService {
	return &p2pMsgService{
		id:      id,
		reactor: r,
		// logger:  logger,
	}
}

func (ms *p2pMsgService) P2PMessages() <-chan nitrop2p.Message {
	return ms.reactor.incoming
}

func (ms *p2pMsgService) Send(p nitrotypes.Participant, m nitrop2p.Message) error {
	if !ms.reactor.IsRunning() {
		return errors.New("reactor not running")
	}

	ms.reactor.send(m)
	return nil
}

func (ms *p2pMsgService) Close() error {
	// the lifecycle of the reactor is managed by cometbft
	return nil
}

func (ms *p2pMsgService) Id() string {
	return ms.id
}
