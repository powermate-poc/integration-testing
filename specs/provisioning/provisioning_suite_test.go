package devices_test

import (
	"encoding/json"
	"fmt"
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
	RunSpecs(t, "Provisioning")
}

type Thing struct {
	Name string `json:"name"`
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

func checkIfDeviceWasCreated() []Thing {
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

	Expect(devices).Should(ContainElement(Thing{Name: deviceName}))

	return devices
}

var _ = Describe("Provisioning Endpoint", func() {
	It("should create device", func() {
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

		checkIfDeviceWasCreated()
	})

	It("should not create device if exists", func() {
		// create device
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

		checkIfDeviceWasCreated()

		// create same device again
		resp, err = client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("should delete created device if exists", func() {
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

		checkIfDeviceWasCreated()

		req, err = http.NewRequest(http.MethodDelete, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		resp, err = client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
	})

	It("should not delete device if not exists", func() {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)

		deviceName := fmt.Sprintf("does-not-exist-%d", r1.Intn(100))

		path, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err := http.NewRequest(http.MethodDelete, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})
})
