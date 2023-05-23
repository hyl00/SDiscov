package register

import (
	"context"
	"fmt"
	"testing"
)

func TestRegister(t *testing.T) {
	p, err := NewPublisher(context.Background(), []string{"123.57.228.39:2379"}, "", "", "adan", "真帅")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p.register())
}

func TestKeepAlive(t *testing.T) {
	p, err := NewPublisher(context.Background(), []string{"123.57.228.39:2379"}, "", "", "adan", "真帅")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p.Start())
}
