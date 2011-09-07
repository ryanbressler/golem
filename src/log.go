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
	"log"
	"os"
)

func vlog(format string, a ...interface{}) {
	if verbose {
		logger.Output(2, fmt.Sprintf(format, a...))
	}
}

// constructs a new verbose logger that wraps a standard out logger
func NewVerboseLogger() *VerboseLogger {
	return &VerboseLogger{log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)}
}
// extends log.Logger to add functions for verbose logging control and warning
type VerboseLogger struct {
	*log.Logger
}

func (this *VerboseLogger) Debug(format string, a ...interface{}) {
	if verbose {
		this.Output(2, fmt.Sprintf(format, a...))
	}
}
func (this *VerboseLogger) Warn(err os.Error) {
	this.Output(2, fmt.Sprintf("[WARN] %v", err))
}
