package nntp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Conn struct {
	net.Conn
	*bufio.Scanner
}

func Dial(network, address string) (*Conn, error) {
	netConn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	conn := Conn{netConn, bufio.NewScanner(netConn)}

	ok := conn.Scan()
	if !ok {
		return nil, conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "200" {
		return nil, errors.New(resp)
	}
	return &conn, nil
}

func (conn *Conn) ModeReader() error {
	fmt.Fprintln(conn, "MODE READER")
	ok := conn.Scan()
	if !ok {
		return conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "200" {
		return errors.New(resp)
	}
	return nil
}

func (conn *Conn) List() ([]string, error) {
	fmt.Fprintln(conn, "LIST")
	ok := conn.Scan()
	if !ok {
		return nil, conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "215" {
		return nil, errors.New(resp)
	}

	list := make([]string, 0)
	for conn.Scan() {
		text := conn.Text()
		if text == "." {
			break
		}
		list = append(list, text)
	}
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

type Group struct {
	first, last int
	headers     []string
}

func (conn *Conn) Group(group string) (Group, error) {
	fmt.Fprintln(conn, fmt.Sprintf("GROUP %s", group))
	ok := conn.Scan()
	if !ok {
		return Group{}, conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "211" {
		return Group{}, errors.New(resp)
	}

	fields := strings.Fields(resp)
	if len(fields) < 5 {
		return Group{}, fmt.Errorf("GROUP: Can't parse %v", resp)
	}
	first, err := strconv.Atoi(fields[2])
	if err != nil {
		return Group{}, err
	}
	last, err := strconv.Atoi(fields[3])
	if err != nil {
		return Group{}, err
	}

	headers, err := conn.ListOverviewFmt()
	if err != nil {
		return Group{}, err
	}

	return Group{first: first, last: last, headers: headers}, nil
}

func (conn *Conn) ListOverviewFmt() ([]string, error) {
	fmt.Fprintln(conn, "LIST OVERVIEW.FMT")
	ok := conn.Scan()
	if !ok {
		return nil, conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "215" {
		return nil, errors.New(resp)
	}

	headers := make([]string, 0)
	for conn.Scan() {
		text := conn.Text()
		if text == "." {
			break
		}
		headers = append(headers, text)
	}
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return headers, nil
}

type Article struct {
	Headers map[string]string
}

func (conn *Conn) Xover(group Group) ([]Article, error) {
	first := group.first
	last := group.last
	fmt.Fprintln(conn, fmt.Sprintf("XOVER %d-%d", first, last))
	ok := conn.Scan()
	if !ok {
		return nil, conn.Err()
	}
	resp := conn.Text()
	if resp[:3] != "224" {
		return nil, errors.New(resp)
	}

	articles := make([]Article, 0)
	var count int
	for conn.Scan() {
		text := conn.Text()
		if text == "." {
			break
		}
		count++
		fmt.Printf("\r%d / %d", count, last-first)
		split := strings.Split(text, "\t")
		a := Article{make(map[string]string)}
		for i, h := range group.headers {
			a.Headers[h] = split[i+1]
		}
		articles = append(articles, a)
	}
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}
