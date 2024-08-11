package main

import (
	"context"
	"fmt"
	"image/color"

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

	if err := client.SetRGBAll(context.Background(), color.RGBA{
		R: 255,
		G: 0,
		B: 0,
	}); err != nil {
		panic(err)
	}

	fmt.Println("keyboard should be white now")
}
