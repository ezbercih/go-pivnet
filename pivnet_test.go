package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-cf-experimental/go-pivnet/logger/loggerfakes"
)

type pivnetErr struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var _ = Describe("PivnetClient", func() {
	var (
		server    *ghttp.Server
		client    pivnet.Client
		token     string
		userAgent string

		releases pivnet.ReleasesResponse

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		releases = pivnet.ReleasesResponse{Releases: []pivnet.Release{
			{
				ID:      1,
				Version: "1234",
			},
			{
				ID:      99,
				Version: "some-other-version",
			},
		}}

		server = ghttp.NewServer()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig = pivnet.ClientConfig{
			Host:      server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	It("has authenticated headers for each request", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for _, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		}

		_, err := client.Releases.List(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets custom user agent", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.VerifyHeaderKV("User-Agent", userAgent),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for _, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyHeaderKV("User-Agent", userAgent),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		}

		_, err := client.Releases.List(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when parsing the url fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Host = "%%%"
			client = pivnet.NewClient(newClientConfig, fakeLogger)

			_, err := client.Releases.List("some product")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("%%%"))
		})
	})

	Context("when making the request fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Host = "https://not-a-real-url.com"
			client = pivnet.NewClient(newClientConfig, fakeLogger)

			_, err := client.Releases.List("some-product")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when Pivnet returns a 401", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an ErrUnauthorized error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusUnauthorized, body),
				),
			)

			_, err := client.Releases.List("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrUnauthorized{
					ResponseCode: http.StatusUnauthorized,
					Message:      "foo message",
				},
			))
		})
	})

	Context("when Pivnet returns a 404", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an ErrNotFound error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusNotFound, body),
				),
			)

			_, err := client.Releases.List("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrNotFound{
					ResponseCode: http.StatusNotFound,
					Message:      "foo message",
				},
			))
		})
	})

	Context("when an unexpected status code comes back from Pivnet", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusTeapot, body),
				),
			)

			_, err := client.Releases.List("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				"Pivnet returned status code: 418 for the request - expected 200"))
		})

		Context("when unmarshalling the response from Pivnet returns an error", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.Releases.List("my-product-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("JSON"))
			})
		})
	})

	Context("when the json unmarshalling fails with error", func() {
		It("forwards the error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusOK, "%%%"),
				),
			)

			_, err := client.Releases.List("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})

	Context("when nil interface is provided for deserialization", func() {
		It("skips deserialization", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/some/endpoint", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, "{}"),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/some/endpoint",
				http.StatusOK,
				nil,
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})