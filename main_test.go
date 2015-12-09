package main_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	main "starwars-countdown"
	"testing"
)

func TestStarwarsCountdown(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StarwarsCountdown Suite")
}

var _ = Describe("Server", func() {
	var server http.Handler
	var req *http.Request
	var resp *httptest.ResponseRecorder

	BeforeEach(func() {
		var indexPage bytes.Buffer
		// Create a new instance
		countDown := &main.CountDown{indexPage, "/"}
		// And init the server
		server = countDown.NewServer()
		// Record HTTP responses.
		resp = httptest.NewRecorder()
	})

	Describe("GET", func() {
		Context("When requested file doesn't exist", func() {
			It("should return 404", func() {
				req, _ = http.NewRequest("GET", "/not-found-file.txt", nil)
				server.ServeHTTP(resp, req)
				Expect(resp.Code).To(Equal(404))
			})
		})
		Context("When we don't have correct file permissions", func() {
			It("should return 403", func() {
				req, _ = http.NewRequest("GET", "/root/.bashrc", nil)
				server.ServeHTTP(resp, req)
				Expect(resp.Code).To(Equal(404))
			})
		})
	})
})
