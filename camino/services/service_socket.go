// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package services

// ServiceSocket ...
type ServiceSocket struct {
	ipAddr string
	port   int
}

// NewServiceSocket ...
func NewServiceSocket(ipAddr string, port int) *ServiceSocket {
	return &ServiceSocket{
		ipAddr: ipAddr,
		port:   port,
	}
}

// GetIpAddr ...
func (socket *ServiceSocket) GetIpAddr() string {
	return socket.ipAddr
}

// GetPort ...
func (socket *ServiceSocket) GetPort() int {
	return socket.port
}
