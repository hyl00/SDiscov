package discov

import (
	"context"
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	d, err := NewDiscover(context.Background(), []string{"123.57.228.39:2379"}, "", "")
	if err != nil {
		fmt.Println(err)
	}
	go d.Load("test")
	_, ok := <-d.Change
	fmt.Println(ok)
	fmt.Println(d.GetAddrs("test")[0])
	defer d.Stop()
}
