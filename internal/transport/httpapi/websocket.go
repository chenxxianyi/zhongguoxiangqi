package httpapi

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" || r.Header.Get("Sec-WebSocket-Version") != "13" {
		s.writeError(w, r, errors.New("invalid WebSocket handshake"))
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		s.writeError(w, r, errors.New("WebSocket is unavailable"))
		return
	}
	connection, buffer, err := hijacker.Hijack()
	if err != nil {
		return
	}
	acceptHash := sha1.Sum([]byte(key + websocketGUID))
	_, _ = buffer.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	_, _ = buffer.WriteString("Upgrade: websocket\r\nConnection: Upgrade\r\n")
	_, _ = buffer.WriteString("Sec-WebSocket-Accept: " + base64.StdEncoding.EncodeToString(acceptHash[:]) + "\r\n\r\n")
	if err := buffer.Flush(); err != nil {
		_ = connection.Close()
		return
	}
	after := r.URL.Query().Get("afterEventId")
	channel, history, cancel := s.events.Subscribe(r.PathValue("id"), after)
	defer cancel()
	defer connection.Close()

	closed := make(chan struct{})
	go consumeWebSocket(connection, closed)
	for _, event := range history {
		if err := writeWebSocketJSON(connection, event); err != nil {
			return
		}
	}
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case event := <-channel:
			if err := writeWebSocketJSON(connection, event); err != nil {
				return
			}
		case <-heartbeat.C:
			if err := writeWebSocketFrame(connection, 0x9, nil); err != nil {
				return
			}
		case <-closed:
			return
		case <-r.Context().Done():
			return
		}
	}
}

func writeWebSocketJSON(writer io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return writeWebSocketFrame(writer, 0x1, data)
}

func writeWebSocketFrame(writer io.Writer, opcode byte, payload []byte) error {
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	case len(payload) <= 65535:
		header = append(header, 126, 0, 0)
		binary.BigEndian.PutUint16(header[len(header)-2:], uint16(len(payload)))
	default:
		header = append(header, 127, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(header[len(header)-8:], uint64(len(payload)))
	}
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func consumeWebSocket(connection net.Conn, closed chan<- struct{}) {
	defer close(closed)
	reader := bufio.NewReader(connection)
	for {
		first, err := reader.ReadByte()
		if err != nil {
			return
		}
		second, err := reader.ReadByte()
		if err != nil {
			return
		}
		opcode := first & 0x0f
		masked := second&0x80 != 0
		length := uint64(second & 0x7f)
		switch length {
		case 126:
			var data [2]byte
			if _, err := io.ReadFull(reader, data[:]); err != nil {
				return
			}
			length = uint64(binary.BigEndian.Uint16(data[:]))
		case 127:
			var data [8]byte
			if _, err := io.ReadFull(reader, data[:]); err != nil {
				return
			}
			length = binary.BigEndian.Uint64(data[:])
		}
		if length > 64<<10 {
			return
		}
		var mask [4]byte
		if masked {
			if _, err := io.ReadFull(reader, mask[:]); err != nil {
				return
			}
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return
		}
		if opcode == 0x8 {
			return
		}
	}
}
