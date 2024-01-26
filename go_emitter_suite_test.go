package emitter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoEmitter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoEmitter Suite")
}
