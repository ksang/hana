/*
Package pusher provides functionality for parsing and pushing metrics to prometheus
*/
package pusher

// Pusher is the common interface defining how to consume data from a datasource
type Pusher interface {
	// Start a new pusher and by providing a datasource channel
	Start(chan string) error
	// Stop the pusher
	Stop() error
}
