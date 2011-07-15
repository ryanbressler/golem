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
	"goconf.googlecode.com/hg"
)

type Configuration struct {
	ConfigFile *conf.ConfigFile
}

func NewConfiguration(filepath string) Configuration {
	if filepath != "" {
		c, err := conf.ReadConfigFile(filepath)
		if err != nil {
			panic(err)
		}
		return Configuration{ConfigFile: c}
	}
	panic(fmt.Sprintf("configuration file not found [%v]", filepath))
}

func (c *Configuration) GetString(section string, key string) string {
	val, err := c.ConfigFile.GetString(section, key)
	if err != nil {
		panic(err)
	}
	if val != "" {
		return val
	}

	panic(fmt.Sprintf("configuration is missing required value [%v]", key))
}
