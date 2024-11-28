package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	url           = "https://www.wissel.nl/collections/cadeaubon-kopen-met-korting-coolblue"
	checkInterval = 1 * time.Minute
	debugSleep    = 10 * time.Second
)

var debug = flag.Bool("debug", false, "Enable debug mode")

func main() {
	flag.Parse()

	for {
		fmt.Println("Checking stock...")

		ctx, cancel := createChrome(*debug)
		outOfStock := isOutOfStock(ctx)

		if *debug {
			log.Printf("Is out of stock: %v", outOfStock)
			time.Sleep(debugSleep)
		}

		if !outOfStock {
			sendTelegramAlert()
		}

		if !*debug {
			cancel()
		}

		fmt.Println("Sleeping...")
		time.Sleep(checkInterval)
	}
}

func createChrome(debug bool) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("headless", !debug), // Set headless based on debug flag
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	return ctx, cancel
}

func isOutOfStock(ctx context.Context) bool {
	var pageContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("body", &pageContent),
	)
	if err != nil {
		log.Printf("Error fetching page: %v", err)
		return true
	}

	return strings.Contains(strings.ToLower(pageContent), "tijdelijk uitverkocht")
}

func sendTelegramAlert() {
	cmd := exec.Command("telegram-send", url)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error sending Telegram alert: %v", err)
	} else {
		log.Println("Telegram alert sent successfully")
	}
}
