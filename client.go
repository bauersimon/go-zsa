package zsa

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bauersimon/go-zsa/api"
)

// Client represents a connection to the ZSA keyboard service.
type Client struct {
	client     api.KeyboardServiceClient
	connection *grpc.ClientConn
}

// Connect establishes a connection to the ZSA keyboard service at the specified path or address.
func Connect(path string) (*Client, error) {
	conn, err := grpc.NewClient(path, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     api.NewKeyboardServiceClient(conn),
		connection: conn,
	}, nil
}

// ConnectDefault establishes a connection to the ZSA keyboard service using default settings.
// On Windows, it connects to "localhost:50051", on other platforms, it uses the socket file at "$CONFIG_DIR/.keymapp/keymapp.sock" (as specified by https://github.com/zsa/kontroll?tab=readme-ov-file#prerequisites).
func ConnectDefault() (*Client, error) {
	path := ""
	if runtime.GOOS == "windows" {
		path = "localhost:50051"
	} else {
		config_dir := os.Getenv("CONFIG_DIR")
		if config_dir == "" {
			return nil, errors.New("environment key \"CONFIG_DIR\" not set")
		}
		path = filepath.Join(config_dir, ".keymapp", "keymapp.sock")
	}

	return Connect(path)
}

// GetStatus retrieves the current status of the keyboard service.
// The returned keyboard might be "nil" in case none is currently connected.
func (c *Client) GetStatus(ctx context.Context) (version string, keyboard *api.ConnectedKeyboard, err error) {
	res, err := c.client.GetStatus(ctx, &api.GetStatusRequest{})
	if err != nil {
		return "", nil, err
	}

	return res.KeymappVersion, res.ConnectedKeyboard, nil
}

// GetKeyboards retrieves a list of all detected keyboards.
func (c *Client) GetKeyboards(ctx context.Context) (keyboards []*api.Keyboard, err error) {
	res, err := c.client.GetKeyboards(ctx, &api.GetKeyboardsRequest{})
	if err != nil {
		return nil, err
	}

	return res.Keyboards, nil
}

// ConnectAnyKeyboard attempts to connect to an arbitrary available keyboard.
func (c *Client) ConnectAnyKeyboard(ctx context.Context) error {
	err := wrapSuccessToError(ctx, c.client.ConnectAnyKeyboard, &api.ConnectAnyKeyboardRequest{})
	if err != nil && !strings.Contains(err.Error(), "keyboard already connected") {
		return err
	}

	return nil
}

// ConnectKeyboardIndex connects to a specific keyboard by its index.
func (c *Client) ConnectKeyboardIndex(ctx context.Context, id int32) error {
	err := wrapSuccessToError(ctx, c.client.ConnectKeyboard, &api.ConnectKeyboardRequest{
		Id: id,
	})
	if err != nil && !strings.Contains(err.Error(), "keyboard already connected") {
		return err
	}

	return nil
}

// ConnectKeyboard connects to a specific keyboard.
func (c *Client) ConnectKeyboard(ctx context.Context, keyboard *api.Keyboard) error {
	return c.ConnectKeyboardIndex(ctx, keyboard.Id)
}

// DisconnectKeyboard disconnects from the currently connected keyboard.
func (c *Client) DisconnectKeyboard(ctx context.Context) error {
	err := wrapSuccessToError(ctx, c.client.DisconnectKeyboard, &api.DisconnectKeyboardRequest{})
	if err != nil && !strings.Contains(err.Error(), "no keyboard is connected") {
		return err
	}

	return nil
}

// SetLayer sets the active layer of the connected keyboard.
func (c *Client) SetLayer(ctx context.Context, layer int32) error {
	return wrapSuccessToError(ctx, c.client.SetLayer, &api.SetLayerRequest{
		Layer: layer,
	})
}

// UnsetLayer unsets a previously set layer.
func (c *Client) UnsetLayer(ctx context.Context, layer int32) error {
	return wrapSuccessToError(ctx, c.client.UnsetLayer, &api.SetLayerRequest{
		Layer: layer,
	})
}

// SetRGBLed sets the color of a specific LED on the keyboard.
// Each additional specified LED tirggers a separate API request. To change all LEDs at once, use "SetRGBAll".
func (c *Client) SetRGBLed(ctx context.Context, color color.Color, leds ...int32) error {
	r, g, b, _ := color.RGBA()
	var errs []error
	for _, led := range leds {
		if err := wrapSuccessToError(ctx, c.client.SetRGBLed, &api.SetRGBLedRequest{
			Led:     led,
			Red:     int32(r),
			Green:   int32(g),
			Blue:    int32(b),
			Sustain: 0, // Unclear what the sustain is for now (https://github.com/zsa/kontroll/issues/9).
		}); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// SetRGBAll sets the color of all LEDs on the keyboard.
func (c *Client) SetRGBAll(ctx context.Context, color color.Color) error {
	r, g, b, _ := color.RGBA()
	return wrapSuccessToError(ctx, c.client.SetRGBAll, &api.SetRGBAllRequest{
		Red:     int32(r),
		Green:   int32(g),
		Blue:    int32(b),
		Sustain: 0, // Unclear what the sustain is for now (https://github.com/zsa/kontroll/issues/9).
	})
}

// SetStatusLED sets the status LED on the keyboard.
func (c *Client) SetStatusLED(ctx context.Context, led int32, on bool) error {
	return wrapSuccessToError(ctx, c.client.SetStatusLed, &api.SetStatusLedRequest{
		Led:     led,
		On:      on,
		Sustain: 0, // Unclear what the sustain is for now (https://github.com/zsa/kontroll/issues/9).
	})
}

// IncreaseBrightness increases the brightness of the keyboard.
func (c *Client) IncreaseBrightness(ctx context.Context) error {
	return wrapSuccessToError(ctx, c.client.IncreaseBrightness, &api.IncreaseBrightnessRequest{})
}

// DecreaseBrightness decreases the brightness of the keyboard.
func (c *Client) DecreaseBrightness(ctx context.Context) error {
	return wrapSuccessToError(ctx, c.client.DecreaseBrightness, &api.DecreaseBrightnessRequest{})
}

// Close closes the connection to the ZSA keyboard service.
func (c *Client) Close() error {
	return c.connection.Close()
}

// successResponse is an interface for responses that have a success status.
type successResponse interface {
	GetSuccess() bool
}

// wrapSuccessToError is a helper function that wraps gRPC calls, converting unsuccessful responses to errors.
func wrapSuccessToError[R any, T successResponse](ctx context.Context, request func(context.Context, R, ...grpc.CallOption) (T, error), parameters R) error {
	res, err := request(ctx, parameters)
	if err != nil {
		return err
	} else if !res.GetSuccess() {
		return fmt.Errorf("unsuccessful %T", parameters)
	}

	return nil
}
