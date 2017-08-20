/*
Package datasource provides a common interface for different datasources
*/
package datasource

// Consumer is interface defining how to consume data from a datasource
type Consumer interface {
	// Start a new consumer and returns a channel for returning data
	Start() (chan string, error)
	// Stop the consumer
	Stop() error
}
