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

// Message type constants
const (
	HELLO   = iota //sent from worker to master on connect, body is number of processes available
	CHECKIN        //sent from worker every minute to keep connection alive

	START //sent from master to start job, body is json job
	KILL  //sent from master to stop jobs, SubId indicates what jobs to stop.

	COUT   //cout from worker, body is line of cout
	CERROR //cout from worker, body is line of cerror

	JOBFINISHED //sent from worker on job finish, body is json job SubId set
	JOBERROR    //sent from worker on job error, body is json job, SubId set

	RESTART //Sent by master to nodes telling them to resart and reconnec themselves.
	DIE     //tell nodes to shutdown.
)

// messages sent between master and workers over websockets
type clientMsg struct {
	Type  int
	SubId string
	Body  string
}
