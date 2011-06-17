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
	"fmt"
)


/////////////////////////////////////////////////
//master

type Killable struct {
	Pid string
	SubId string
	JobId string	
}

func (k * Killable) Kill () {
	log("killing %v", k)
	KillPid(k.Pid)
}


type JobKiller struct {
	Killchan chan string
	Donechan chan * Killable
	Registerchan chan * Killable
	
	killables map[string] * Killable
	
}

func NewJobKiller()( jk * JobKiller) {
	jk = &JobKiller{Killchan:make(chan string,3),Donechan: make( chan * Killable, 3),Registerchan: make( chan * Killable, 3),killables: map[string] * Killable{}}
	go jk.KillJobs()
	return
}

func (jk * JobKiller) KillJobs () {
	for {
		select {
		case SubId := <-jk.Killchan:
			vlog("looking for  killables with subid %v", SubId)
			for _,kb := range jk.killables {
				if kb.SubId == SubId {
					kb.Kill()
				}
			}
			vlog("done looking for  killables with subid %v", SubId)
		case kb := <-jk.Registerchan:
			vlog("regestering killable %v", kb)
			jk.killables[fmt.Sprintf("%v%v",kb.SubId,kb.JobId)]=kb
		case kb := <- jk.Donechan:
			vlog("removing killable %v", kb)
			jk.killables[fmt.Sprintf("%v%v",kb.SubId,kb.JobId)]=kb,false
			
		}
	}
}
	
	
	
	