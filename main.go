// sphidewin - Hide/Unmap windows in X11
// Copyright (C) 2023 Safal Piya

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func main() {
	var toUnmapPreviouslySpawned bool
	parseArgs(&toUnmapPreviouslySpawned)

	if len(flag.Args()) != 1 {
		usage(os.Stderr)
		os.Exit(1)
	}

	wmNameToHide := flag.Args()[0]

	X, err := xgb.NewConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer X.Close()

	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	xproto.ChangeWindowAttributes(X, root, xproto.CwEventMask, []uint32{xproto.EventMaskSubstructureNotify})

	var unMappedWindows []xproto.Window
	defer mapUnmappedWindows(X, unMappedWindows)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			mapUnmappedWindows(X, unMappedWindows)
			X.Close()
			os.Exit(0)
		}
	}()

	if toUnmapPreviouslySpawned {
		mappedWindows := getMappedWindows(X, root)
		for _, windowId := range mappedWindows {
			windowClasses := windowClassFromId(X, windowId)
			for _, windowClass := range windowClasses {
				if windowClass == wmNameToHide {
					cookie := xproto.UnmapWindowChecked(X, windowId)
					if cookie.Check() != nil {
						log.Printf("[WARNING] Couldn't unmap %s %d (Previously spawned)", windowClasses, windowId)
						continue
					}
					unMappedWindows = append(unMappedWindows, windowId)
					log.Printf("%s %d (Previously spawned) unmapped", windowClasses, windowId)
				}
			}
		}
	}

	for {
		ev, xerr := X.WaitForEvent()
		if ev == nil && xerr == nil {
			log.Println("Both event and error are nil. Exiting...")
			return
		}

		if xerr != nil {
			log.Printf("Error: %s", xerr)
		}

		switch ev.(type) {
		case xproto.MapRequestEvent:
			fmt.Println("MapRequest")
		case xproto.MapNotifyEvent:
			windowId := ev.(xproto.MapNotifyEvent).Window
			windowClasses := windowClassFromId(X, windowId)
			for _, windowClass := range windowClasses {
				if windowClass == wmNameToHide {
					cookie := xproto.UnmapWindowChecked(X, windowId)
					if cookie.Check() != nil {
						log.Printf("[WARNING] Couldn't unmap %s %d", windowClasses, windowId)
						continue
					}
					unMappedWindows = append(unMappedWindows, windowId)
					log.Printf("%s %d unmapped", windowClasses, windowId)
				}
			}
		default:
			continue
		}
	}
}

func parseArgs(toUnmapPreviouslySpawned *bool) {
	flag.BoolVar(toUnmapPreviouslySpawned, "p", false, "Unmap previously spawned windows")
	toPrintHelp := flag.Bool("h", false, "Show help message")
	toPrintHelp = flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *toPrintHelp {
		usage(os.Stdout)
		os.Exit(0)
	}
}

func usage(file io.Writer) {
	fmt.Fprintf(file, "%s [options] WM_CLASS\n\n", os.Args[0])
	fmt.Fprintf(file, "where [options] are:\n")
	flag.PrintDefaults()
}

func windowClassFromId(X *xgb.Conn, windowId xproto.Window) []string {
	aname := "WM_CLASS"
	nameAtom, err := xproto.InternAtom(X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal(err)
	}
	reply, err := xproto.GetProperty(X, false, windowId, nameAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		log.Fatal(err)
	}

	classSplits := strings.Split(string(reply.Value), "\000")
	return classSplits[:len(classSplits)-1]
}

func mapUnmappedWindows(X *xgb.Conn, unmappedWindows []xproto.Window) {
	for _, windowId := range unmappedWindows {
		cookie := xproto.MapWindowChecked(X, windowId)
		if cookie.Check() != nil {
			log.Println("[WARNING] Couldn't map", windowId)
			continue
		}
		log.Println("Mapped", windowId)
	}
}

func getMappedWindows(X *xgb.Conn, root xproto.Window) []xproto.Window {
	aname := "_NET_CLIENT_LIST"
	nameAtom, err := xproto.InternAtom(X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal(err)
	}
	reply, err := xproto.GetProperty(X, false, root, nameAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		log.Fatal(err)
	}

	var mappedWindows []xproto.Window
	for i := 0; i < int(reply.ValueLen); i++ {
		mappedWindows = append(mappedWindows, xproto.Window(xgb.Get32(reply.Value[i*4:])))
	}
	return mappedWindows
}
