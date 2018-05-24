package network



import (
"fmt"
)
type CLI struct{}

func (cli *CLI) StartNode(nodeID string, i int) {
	fmt.Printf("Starting node %s\n", nodeID)
	StartServer(nodeID, i)
}


