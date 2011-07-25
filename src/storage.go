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
	"os"
)

type JobStore interface {
	Create(item JobDetails, tasks []Task) (err os.Error)

	All() (items []JobDetails, err os.Error)

	Active() (items []JobDetails, err os.Error)

	Unscheduled() (items []JobDetails, err os.Error)

	Get(jobId string) (item JobDetails, err os.Error)

	Tasks(jobId string) (tasks []Task, err os.Error)

	Update(item JobDetails) (err os.Error)
}
