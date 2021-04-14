/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2020 jet tsang zeon-git. All Rights Reserved.
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/valyala/fasthttp"
)

type ShortMap map[string]string

var listUrl string
var blockTxt string
var shortenMap map[string]string
var listen string
var reloadURI = "/reload"

func init() {
	const (
		urlUsage     = "This program will get a list from the url you provide."
		blockUsage   = "html block to query."
		defaultBlock = ""

		defaultListen = "localhost:9530"
		listenUsage   = "TCP address to listen to"
	)
	flag.StringVar(&listUrl, "url", "https://pastebin.com/raw/9g4msZxC", urlUsage)
	flag.StringVar(&listUrl, "u", "https://pastebin.com/raw/9g4msZxC", urlUsage+" (shorthand)")

	flag.StringVar(&blockTxt, "block", defaultBlock, blockUsage)
	flag.StringVar(&blockTxt, "b", defaultBlock, blockUsage+" (shorthand)")

	flag.StringVar(&listen, "listen", defaultListen, listenUsage)
	flag.StringVar(&listen, "l", defaultListen, listenUsage+" (shorthand)")

	shortenMap = map[string]string{}
}
func main() {
	flag.Parse()
	go autoLoad()
	log.Println("Listening to:", listen)
	if err := fasthttp.ListenAndServe(listen, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.RequestURI()) == reloadURI {
		loadData()
		fmt.Fprint(ctx, "ok")
		ctx.SetStatusCode(fasthttp.StatusAccepted)
		return
	}
	if relayUrl, exist := shortenMap[string(ctx.Path())]; exist {
		var newRequest fasthttp.Request
		ctx.Request.Header.CopyTo(&newRequest.Header)
		newRequest.SetRequestURI(relayUrl)
		newRequest.Header.SetBytesV("Host", newRequest.Host())
		ctx.URI().QueryArgs().VisitAll(func(key []byte, value []byte) {
			newRequest.URI().QueryArgs().AddBytesKV(key, value)
		})
		defer fasthttp.ReleaseRequest(&newRequest)

		fasthttp.Do(&newRequest, &ctx.Response)
	} else {
		ctx.Error("err", 404)
	}
}

func autoLoad() {
	for {
		loadData()
		time.Sleep(time.Hour)
	}
}

func loadData() {
	log.Println("Loading data.")
	var code string
	if blockTxt != "" {
		if doc, err := htmlquery.LoadURL(listUrl); err != nil {
			return
		} else if found := htmlquery.FindOne(doc, blockTxt); found != nil {
			return
		} else {
			code = htmlquery.InnerText(found)
		}
	} else {
		var dst []byte
		_, byteCode, _ := fasthttp.Get(dst, string(listUrl))
		code = string(byteCode)
	}
	scanner := bufio.NewScanner(strings.NewReader(code))
	for scanner.Scan() {
		maplist := strings.Split(scanner.Text(), "#")
		shortenMap[maplist[0]] = maplist[1]
	}
	log.Println("urls mapped.")

}
