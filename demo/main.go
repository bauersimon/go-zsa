package main

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"github.com/bauersimon/go-zsa"
)

func main() {
	client, err := zsa.ConnectDefault()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if version, keyboard, err := client.GetStatus(context.Background()); err != nil {
		panic(err)
	} else {
		fmt.Println(version, keyboard)
	}

	if err := client.ConnectAnyKeyboard(context.Background()); err != nil {
		panic(err)
	}

	fmt.Println("successful connection")

	wait := 1000 * time.Millisecond

	for {
		if err := client.SetRGBAll(context.Background(), color.RGBA{
			R: 255,
			G: 255,
			B: 255,
		}); err != nil {
			panic(err)
		}

		time.Sleep(wait)

		if err := client.SetRGBAll(context.Background(), color.RGBA{
			R: 255,
			G: 0,
			B: 0,
		}); err != nil {
			panic(err)
		}

		time.Sleep(wait)

	}
}
