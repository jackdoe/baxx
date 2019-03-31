package monitoring

import (
	"log"
	"testing"
)

func TestExample(t *testing.T) {
	s := GetMDADM("md0")
	log.Printf("%#v", s)
}
