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
	"syscall"
)


// links a SubId and JobId with a pid that can be used to kill it
type Killable struct {
	Pid   int
	SubId string
	JobId int
}

// kills via the stored process id
func (k *Killable) Kill() {
	log("Killable.Kill(%v):", k.Pid)
	errno := syscall.Kill(k.Pid, syscall.SIGCHLD)
	log("Killable.Kill(%v):%v", k.Pid, errno)
}

//A job killer is created to monitor and kill jobs
type JobKiller struct {
	Killchan     chan string    //used to send in the SubId of jobs to kill
	Donechan     chan *Killable //used to indicate that a job is done and should no longer be killable
	Registerchan chan *Killable //used to register a job as a killable

	killables map[string]*Killable //internal structure to keep track of killables by subid+jobId (as strings)
}

//creates a Job Killer and starts its routine KillJobs
func NewJobKiller() (jk *JobKiller) {
	jk = &JobKiller{Killchan: make(chan string, 3), Donechan: make(chan *Killable, 3), Registerchan: make(chan *Killable, 3), killables: map[string]*Killable{}}
	go jk.KillJobs()
	return
}

// should be run as a go routine, monitors job killers channel.  maintains the internal map of killables and locates jobs that need to be killed.
func (jk *JobKiller) KillJobs() {
	for {
		select {
		case SubId := <-jk.Killchan:
			vlog("JobKiller.KillJobs() killing: %v", SubId)
			for _, kb := range jk.killables {
				if kb.SubId == SubId {
					kb.Kill()
				}
			}
			vlog("JobKiller.KillJobs() done killing: %v", SubId)
		case kb := <-jk.Registerchan:
			vlog("JobKiller.KillJobs() registering: %v", kb)
			jk.killables[fmt.Sprintf("%v%v", kb.SubId, kb.JobId)] = kb
		case kb := <-jk.Donechan:
			vlog("JobKiller.KillJobs() removing: %v", kb)
			jk.killables[fmt.Sprintf("%v%v", kb.SubId, kb.JobId)] = kb, false
		}
	}
}
