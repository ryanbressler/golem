/*
   Copyright (C) 2003-2010 Institute for Systems Biology
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


//tls configurability
var useTls bool = true

//password
var usepw bool = false
var hash = sha256.New() // or whatever
var hashedpw string


const (
	//Message type constants
	HELLO = 1


	START   = 3
	CHECKIN = 4

	COUT   = 5
	CERROR = 6

	JOBFINISHED = 7
	JOBERROR    = 8
)
