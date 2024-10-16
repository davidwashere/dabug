package main

import (
	"fmt"

	"github.com/davidwashere/dabug/log"
)

func main() {
	log.Append("A")
	thinga()
	log.Append("B")
	thingb()

	log.Append("wat")

	thingc()
	log.Flush()
}

func thinga() {
	fmt.Printf("thinga\n")
}

func thingb() {
	fmt.Printf("thingb\n")
}

/*









































































 */

func thingc() {
	fmt.Printf("thingc\n")
	log.Append("C")
}
