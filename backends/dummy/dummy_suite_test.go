package dummy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDummy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dummy Suite")
}
