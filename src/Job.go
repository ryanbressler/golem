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
	"json"
)

//Internal Job Representation used primairly as the body of job related messages
type Job struct {
	SubId  string
	LineId int
	JobId  int
	Args   []string
}

//NewJob creates a job struct from a json string (usually a message body)
func NewJob(jsonjob string) *Job {
	var job Job

	err := json.Unmarshal([]byte(jsonjob), &job)
	if err != nil {
		log("error parseing job json: %v", err)
	}
	return &job

}
