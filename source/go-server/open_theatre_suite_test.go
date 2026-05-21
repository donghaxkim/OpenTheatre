package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpenTheatre(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenTheatre Go Server Suite")
}
