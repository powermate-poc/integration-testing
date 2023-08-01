package devices_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"powermate-integration-testing/configuration"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devices")
}

type Thing struct {
	Name string `json:"name"`
}

var (
	config *configuration.Configuration
)

var _ = BeforeSuite(func() {
	config = configuration.Load()
})

var _ = Describe("Devices Endpoint", func() {
	It("should list devices", func() {
		path, err := url.JoinPath(config.Host, "/api/devices")
		Expect(err).NotTo(HaveOccurred())

		req, err := http.NewRequest(http.MethodGet, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var devices []Thing
		err = json.NewDecoder(resp.Body).Decode(&devices)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(devices)).Should(BeNumerically(">", 0))
	})
})
