/*
   Copyright (C) 2003-2011 Institute for Systems Biology
                           Seattle, Washington, USA.

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library; if not, write to the Free Software
   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA

*/
package main

import (
	"crypto/sha256"
)


var verbose = false
var iobuffersize = 1000

//
var isMaster bool
var isScribe bool
var configuration Configuration

//tls configurability
var useTls bool = true

//password
var usepw bool = false
var hash = sha256.New() // use the same hasher
var certpath string


const (
	//Message type constants ... should maybe be in clientMsg.go 
	HELLO   = iota //sent from client to master on connect, body is bumber of jobs at once
	CHECKIN        //sent from client every minute to keep conection alive

	START //sent from master to start job, body is json job
	KILL  //sent from master to stop jobs, SubId indicates what jobs to stop.

	COUT   //cout from clent, body is line of cout
	CERROR //cout from clent, body is line of cerror

	JOBFINISHED //sent from client on job finish, body is json job SubId set
	JOBERROR    //sent from client on job error, body is json job, SubId set

	RESTART //Sent by master to nodes telling them to resart and reconnec themselves.
	DIE     //tell nodes to shutdown.
)
