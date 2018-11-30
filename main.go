// pmos-bot - A bot for the postmarketOS Matrix channels
// Copyright (C) 2017 Tulir Asokan
// Copyright (C) 2018 Luca Weiss
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"maunium.net/go/mautrix"
	"regexp"
	"strings"
)

var homeserver = flag.String("homeserver", "https://matrix.org", "Matrix homeserver")
var username = flag.String("username", "", "Matrix username localpart")
var password = flag.String("password", "", "Matrix password")

func main() {
	flag.Parse()
	fmt.Println("Logging to", *homeserver, "as", *username)
	client, err := mautrix.NewClient(*homeserver, "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := client.Login(&mautrix.ReqLogin{Type: "m.login.password", User: *username, Password: *password})
	if err != nil {
		fmt.Println(err)
		return
	}
	client.SetCredentials(resp.UserID, resp.AccessToken)

	fmt.Println("Login successful")

	shortcutmap := map[string]string{
		"pma#": "https://gitlab.com/postmarketOS/pmaports/issues/",
		"pma!": "https://gitlab.com/postmarketOS/pmaports/merge_requests/",
		"pmb#": "https://gitlab.com/postmarketOS/pmbootstrap/issues/",
		"pmb!": "https://gitlab.com/postmarketOS/pmbootstrap/merge_requests/",
		"org#": "https://gitlab.com/postmarketOS/postmarketos.org/issues/",
		"org!": "https://gitlab.com/postmarketOS/postmarketos.org/merge_requests/",
	}
	shortcutmapregex := regexp.MustCompile("(pma[#!]|pmb[#!]|org[#!])(\\d+)")

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(mautrix.EventMessage, func(evt *mautrix.Event) {
		if evt.Sender != *username &&
			(evt.RoomID == "!clcCCNrLZYwdfNqkkR:disroot.org" || // #postmarketos:disroot.org
				evt.RoomID == "!MxNOnZlZaurAGfcxFy:matrix.org" || // #postmarketos-lowlevel:disroot.org
				evt.RoomID == "!VTQfOrQIBniIdCuMOq:matrix.org" || // #postmarketos-offtopic:disroot.org
				evt.RoomID == "!NBvxopLbDoLCDlqKkL:z3ntu.xyz") { // #test2:z3ntu.xyz
			matches := shortcutmapregex.FindAllStringSubmatch(evt.Content.Body, -1)
			if matches != nil {
				var buffer bytes.Buffer
				for _, match := range matches {
					fmt.Println(match[1] + match[2] + " matched!")
					fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.Body)
					buffer.WriteString(shortcutmap[match[1]] + match[2] + " ")
				}
				content := mautrix.Content{MsgType: mautrix.MsgText, Body: strings.TrimSuffix(buffer.String(), " ")}
				_, err := client.SendMessageEvent(evt.RoomID, mautrix.EventMessage, content)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	})

	err = client.Sync()
	if err != nil {
		fmt.Println(err)
	}
}
