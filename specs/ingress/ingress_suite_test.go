package devices_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"powermate-integration-testing/configuration"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingress")
}

var (
	config     *configuration.Configuration
	deviceName string
)

var _ = BeforeSuite(func() {
	config = configuration.Load()
})

var _ = BeforeEach(func() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	deviceName = fmt.Sprintf("it-test-device-provisioning-%d", r1.Intn(100))
})

var _ = AfterEach(func() {
	path, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s", deviceName))
	Expect(err).NotTo(HaveOccurred())

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

	client := &http.Client{}
	resp, err2 := client.Do(req)
	Expect(err2).NotTo(HaveOccurred())
	defer resp.Body.Close()

	Expect(resp.StatusCode).To(Or(Equal(http.StatusNoContent), Equal(http.StatusNotFound)))
})

var _ = Describe("Ingress Endpoint", func() {
	It("should send data to it", func() {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)

		//// create device
		path, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err := http.NewRequest(http.MethodPut, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		//// send data through ingress
		type Measurement struct {
			Name  string  `json:"name"`
			Value float64 `json:"value"`
		}

		type Body struct {
			Measurements []Measurement `json:"measurements"`
		}

		i := r1.Float64()
		body := &Body{Measurements: []Measurement{{Name: "x", Value: i}}}

		out, _ := json.Marshal(body)

		path2, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s/ingress", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err = http.NewRequest(http.MethodPost, path2, bytes.NewBuffer(out))
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		resp, err = client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

		//// check data arrived correctly
		path3, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s/sensors/x/current", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err = http.NewRequest(http.MethodGet, path3, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		response := make([]byte, 1000)

		for ok := true; ok; ok = (string(response) == "[]") {
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			_, err := resp.Body.Read(response)
			if err != nil {
				log.Panic("could not read response")
			}
		}

		Expect(string(response)).To(Not(Equal("[]")))
	})
})
