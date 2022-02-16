package handlers_test

import (
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/onsi/gomega/ghttp"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

var _ = Describe("Server", func() {
	var server *ghttp.Server
	var body io.Reader

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Context("when post request is sent to /api/user/register path", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				handlers.Signup,
			)
			body = strings.NewReader("{\"login\": \"lsnudds\",\"password\": \"password>\"}")
		})
		It("Returns 200 OK", func() {
			resp, err := http.Post(server.URL()+"/api/user/register", "application/json", body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
	})
})
