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


//killable links a SubId and JobId which uniquelly describe a job with a pid that can be used to 
//kill it
type Killable struct {
	Pid   string
	SubId string
	JobId string
}

//Killable.Kill() kills the killable via KillPid (linux kill).
func (k *Killable) Kill() {
	log("killing %v", k)
	KillPid(k.Pid)
}

//A job killer is created to monitor and kill jobs
type JobKiller struct {
	Killchan     chan string    //used to send in the SubId of jobs to kill
	Donechan     chan *Killable //used to indicate that a job is done and should no longer be killable
	Registerchan chan *Killable //used to register a job as a killable

	killables map[string]*Killable //internal stcuture to keep track of killables by subid+jobId (as strings)
}

//creates a Job Killer and starts its goroutine KillJobs
func NewJobKiller() (jk *JobKiller) {
	jk = &JobKiller{Killchan: make(chan string, 3), Donechan: make(chan *Killable, 3), Registerchan: make(chan *Killable, 3), killables: map[string]*Killable{}}
	go jk.killJobs()
	return
}

//killJobs should be run as a go routine, it monitors its job killers channel  upkeeps the 
//internal map of killables and locates jobs that ned to be killed.
func (jk *JobKiller) killJobs() {
	for {
		select {
		case SubId := <-jk.Killchan:
			vlog("looking for  killables with subid %v", SubId)
			for _, kb := range jk.killables {
				if kb.SubId == SubId {
					kb.Kill()
				}
			}
			vlog("done looking for  killables with subid %v", SubId)
		case kb := <-jk.Registerchan:
			vlog("regestering killable %v", kb)
			jk.killables[fmt.Sprintf("%v%v", kb.SubId, kb.JobId)] = kb
		case kb := <-jk.Donechan:
			vlog("removing killable %v", kb)
			jk.killables[fmt.Sprintf("%v%v", kb.SubId, kb.JobId)] = kb, false

		}
	}
}
