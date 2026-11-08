package main

import (
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func main() {
	container := composition.NewContainer()
	cli.Execute(container.ToServices())
}
