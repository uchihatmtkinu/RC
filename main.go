package main
import (
	"RC/Reputation"
)

func main() {
	bc := Reputation.NewRepBlockchain()
	defer bc.Db.Close()

	cli := Reputation.CLI{bc}
	cli.Run()
}