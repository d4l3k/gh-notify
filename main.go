package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	page         = flag.String("page", "https://geekhack.org/index.php?topic=79513.0", "the page to check")
	pattern      = flag.String("pattern", ".post", "pattern to check differences on")
	every        = flag.Duration("every", 10*time.Second, "check every duration")
	startupSound = flag.String("startupsound", "startup.oga", "startup sound")
	changedSound = flag.String("changedsound", "notify.wav", "notify sound")
)

func playSound(sound string) error {
	if _, err := exec.Command("paplay", sound).Output(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		_ = playSound(*startupSound)
		log.Fatal(err)
	}
}

var first = true
var old string

func check() error {
	ctx, cancel := context.WithTimeout(context.Background(), *every)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, *page, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("got status %+v: %q", resp.StatusCode, resp.Status)
		if err := playSound(*startupSound); err != nil {
			return err
		}
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	body, err := goquery.OuterHtml(doc.Find(*pattern))
	if err != nil {
		return err
	}

	hash := sha1.Sum([]byte(body))
	log.Printf("fetched! hash %s", base64.StdEncoding.EncodeToString(hash[:]))

	if first {
		log.Printf("--- initial body ---\n\n%s\n\n------", body)
	}
	if !first && old != body {
		if err := playSound(*changedSound); err != nil {
			return err
		}
	}
	old = body
	first = false
	return nil
}

func run() error {
	flag.Parse()

	if err := playSound(*startupSound); err != nil {
		return err
	}
	tick := time.Tick(*every)

	for {
		if err := check(); err != nil {
			return err
		}
		<-tick
	}
}
