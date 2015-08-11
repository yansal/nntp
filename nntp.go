package nntp

import (
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

type Conn struct {
	*textproto.Conn
	message string
}

func Dial(network, address string) (*Conn, error) {
	textprotoConn, err := textproto.Dial(network, address)
	if err != nil {
		return nil, err
	}

	conn := Conn{textprotoConn, ""}
	_, message, err := conn.ReadCodeLine(200)
	if err != nil {
		return nil, err
	}
	conn.message = message
	return &conn, nil
}

func (conn *Conn) ModeReader() error {
	id, err := conn.Cmd("MODE READER")
	if err != nil {
		return err
	}
	conn.StartResponse(id)
	defer conn.EndResponse(id)

	if _, _, err = conn.ReadCodeLine(200); err != nil {
		return err
	}
	return nil
}

func (conn *Conn) List() ([]string, error) {
	id, err := conn.Cmd("LIST")
	if err != nil {
		return nil, err
	}
	conn.StartResponse(id)
	defer conn.EndResponse(id)

	if _, _, err = conn.ReadCodeLine(215); err != nil {
		return nil, err
	}

	return conn.ReadDotLines()
}

type Group struct {
	first, last int
	headers     []string
}

func (conn *Conn) Group(s string) (Group, error) {
	group, err := conn.groupCmd(s)
	if err != nil {
		return Group{}, err
	}

	headers, err := conn.ListOverviewFmt()
	if err != nil {
		return Group{}, err
	}
	group.headers = headers
	return group, nil
}

func (conn *Conn) groupCmd(s string) (Group, error) {
	id, err := conn.Cmd("GROUP %s", s)
	if err != nil {
		return Group{}, err
	}
	conn.StartResponse(id)
	defer conn.EndResponse(id)

	_, msg, err := conn.ReadCodeLine(211)
	if err != nil {
		return Group{}, err
	}

	fields := strings.Fields(msg)
	if len(fields) < 4 {
		return Group{}, fmt.Errorf("GROUP: Can't parse %v", msg)
	}
	first, err := strconv.Atoi(fields[1])
	if err != nil {
		return Group{}, err
	}
	last, err := strconv.Atoi(fields[2])
	if err != nil {
		return Group{}, err
	}
	return Group{first: first, last: last}, nil
}

func (conn *Conn) ListOverviewFmt() ([]string, error) {
	id, err := conn.Cmd("LIST OVERVIEW.FMT")
	if err != nil {
		return nil, err
	}
	conn.StartResponse(id)
	defer conn.EndResponse(id)

	if _, _, err := conn.ReadCodeLine(215); err != nil {
		return nil, err
	}

	lines, err := conn.ReadDotLines()
	if err != nil {
		return nil, err
	}
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], ":")
	}
	return lines, nil
}

type Article struct {
	Headers map[string]string
}

func (conn *Conn) Xover(group Group) ([]Article, error) {
	first := group.first
	last := group.last

	id, err := conn.Cmd("XOVER %d-%d", first, last)
	if err != nil {
		return nil, err
	}
	conn.StartResponse(id)
	defer conn.EndResponse(id)

	if _, _, err := conn.ReadCodeLine(224); err != nil {
		return nil, err
	}

	articles := make([]Article, 0)
	var count int
	for {
		line, err := conn.ReadLine()
		if err != nil {
			return nil, err
		}

		if line == "." {
			break
		}

		count++
		fmt.Printf("\r%d / %d", count, last-first)
		split := strings.Split(line, "\t")
		a := Article{make(map[string]string)}
		for i, h := range group.headers {
			a.Headers[h] = split[i+1]
		}
		articles = append(articles, a)
	}

	return articles, nil
}
