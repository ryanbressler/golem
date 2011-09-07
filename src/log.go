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
	"time"
	"fmt"
)

// wrapper for fmt.Printf that prepends the time
// log("error is %v, n is %v", err, n)
// log("hello")
func log(format string, a ...interface{}) {
	t := time.LocalTime()
	format = t.String() + ": " + format + "\n"
	fmt.Printf(format, a...)
}

// wrapper for log that only prints if verbose=true in config file
func vlog(format string, a ...interface{}) {
	if verbose {
		log(format, a...)
	}
}

// wrapper for log that appends WARN message
func warn(format string, a ...interface{}) {
	log("WARN: "+format, a...)
}
