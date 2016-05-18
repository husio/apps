package qux

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"sync"
)

type Server struct {
	mu     sync.Mutex
	stacks map[string][][]byte
}

func NewServer() *Server {
	return &Server{
		stacks: make(map[string][][]byte),
	}
}

func (s *Server) Serve(rw io.ReadWriter) error {
	buf := bufio.NewReadWriter(bufio.NewReader(rw), bufio.NewWriter(rw))
	for {
		if err := s.handleSingle(buf); err != nil {
			return err
		}
		if err := buf.Flush(); err != nil {
			return err
		}
	}
}

func (s *Server) handleSingle(rw *bufio.ReadWriter) error {
	line, err := rw.ReadBytes('\n')
	if err != nil {
		return err
	}
	args := bytes.Fields(line)
	if len(args) == 0 {
		return nil
	}

	switch {
	case bytes.Equal(args[0], []byte("READ")):
		// READ topic offset limit\n
		if len(args) != 4 {
			fmt.Fprintf(rw, "ERR READ requires 3 arg, got %d\n", len(args)-1)
			return nil
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		stack, ok := s.stacks[string(args[1])]
		if !ok {
			fmt.Fprintf(rw, "ERR stack %s does not exist\n", args[1])
			return nil
		}
		offset, err := strconv.ParseInt(string(args[2]), 10, 64)
		if err != nil || offset < 0 {
			fmt.Fprint(rw, "ERR invalid offset\n")
			return nil
		}
		limit, err := strconv.ParseInt(string(args[3]), 10, 64)
		if err != nil || limit <= 0 {
			fmt.Fprint(rw, "ERR invalid limit\n")
			return nil
		}

		if offset > int64(len(stack)) {
			return nil
		}
		if offset+limit > int64(len(stack)) {
			limit = int64(len(stack)) - offset
		}

		for i, msg := range stack[offset : offset+limit] {
			fmt.Fprintf(rw, "MSG %d %d\n", offset+int64(i), len(msg))
			rw.Write(msg)
			rw.WriteByte('\n')
		}
		rw.WriteString("END\n")

	case bytes.Equal(args[0], []byte("PUSH")):
		// PUSH topic msg-size\n
		// msg\n
		if len(args) != 3 {
			fmt.Fprintf(rw, "ERR PUSH requires 2 args, got %d\n", len(args)-1)
			return nil
		}

		msgsize, err := strconv.ParseInt(string(args[2]), 10, 64)
		if err != nil || msgsize <= 0 {
			fmt.Fprint(rw, "ERR invalid message size\n")
			return nil
		}

		b := make([]byte, msgsize+1)
		if n, err := rw.Read(b); err != nil {
			fmt.Fprintf(rw, "ERR cannot read: %s\n", err)
			return nil
		} else if n != len(b) {
			fmt.Fprint(rw, "ERR incompete message\n")
			return nil
		}
		if b[msgsize] != '\n' {
			fmt.Fprint(rw, "ERR invalid message termination\n")
			return nil
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		s.stacks[string(args[1])] = append(s.stacks[string(args[1])], b[:msgsize])

		fmt.Fprint(rw, "OK\n")

	case bytes.Equal(args[0], []byte("LEN")):
		// LEN topic \n
		if len(args) != 2 {
			fmt.Fprintf(rw, "ERR LEN requires 1 arg, got %d\n", len(args)-1)
			return nil
		}
		s.mu.Lock()
		fmt.Fprintf(rw, "%d\n", len(s.stacks[string(args[1])]))
		defer s.mu.Unlock()
	case bytes.Equal(args[0], []byte("DUMP")):
		s.mu.Lock()
		for name, stack := range s.stacks {
			fmt.Fprintf(rw, "%s %s\n", name, stack)
		}
		rw.WriteString("END\n")
		defer s.mu.Unlock()

	default:
		fmt.Fprintf(rw, "ERR unknown command: %q\n", args)
	}
	return nil
}
