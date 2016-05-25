package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors/errorsfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("product file commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string

		releases     []pivnet.Release
		productFile  pivnet.ProductFile
		productFiles []pivnet.ProductFile

		responseStatusCode int
		response           interface{}
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorsfakes.FakeErrorHandler{}
		commands.ErrorHandler = fakeErrorHandler

		productSlug = "some-product-slug"

		productFile = pivnet.ProductFile{
			ID:   1234,
			Name: "some product file",
		}

		releases = []pivnet.Release{
			{
				ID:      1234,
				Version: "some-release-version",
			},
			{
				ID:      2345,
				Version: "another-release-version",
			},
		}

		productFiles = []pivnet.ProductFile{
			{
				ID:   1234,
				Name: "Some name",
			},
			{
				ID:   2345,
				Name: "Another name",
			},
		}

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ProductFilesCommand", func() {
		var (
			command commands.ProductFilesCommand
		)

		BeforeEach(func() {
			response = pivnet.ProductFilesResponse{
				ProductFiles: productFiles,
			}

			command = commands.ProductFilesCommand{
				ProductSlug: productSlug,
			}
		})

		Describe("when only product-slug is provided", func() {
			JustBeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/products/%s/product_files",
								apiPrefix,
								productSlug,
							),
						),
						ghttp.RespondWithJSONEncoded(responseStatusCode, response),
					),
				)
			})

			It("lists all product files for the provided product slug", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returned []pivnet.ProductFile

				err = json.Unmarshal(outBuffer.Bytes(), &returned)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal(productFiles))
			})

			Context("when there is an error", func() {
				BeforeEach(func() {
					responseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the release version is provided", func() {
			It("lists all product files for the provided product slug and release version", func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: releases,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)

				productFilesResponse := pivnet.ProductFilesResponse{
					ProductFiles: productFiles,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/products/%s/releases/%d/product_files",
								apiPrefix,
								productSlug,
								releases[0].ID,
							),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, productFilesResponse),
					),
				)

				productFilesCommand := commands.ProductFilesCommand{
					ProductSlug:    productSlug,
					ReleaseVersion: releases[0].Version,
				}

				err := productFilesCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returned []pivnet.ProductFile

				err = json.Unmarshal(outBuffer.Bytes(), &returned)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal(productFiles))
			})

		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFilesCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFilesCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("ProductFileCommand", func() {
		var (
			command commands.ProductFileCommand
		)

		BeforeEach(func() {
			response = pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			command = commands.ProductFileCommand{
				ProductSlug:   productSlug,
				ProductFileID: productFile.ID,
			}
		})

		Describe("when only product-slug is provided", func() {
			JustBeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/products/%s/product_files/%d",
								apiPrefix,
								productSlug,
								productFile.ID,
							),
						),
						ghttp.RespondWithJSONEncoded(responseStatusCode, response),
					),
				)
			})

			It("shows the product file for the provided product slug and product file id", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returned pivnet.ProductFile

				err = json.Unmarshal(outBuffer.Bytes(), &returned)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal(productFile))
			})

			Context("when there is an error", func() {
				BeforeEach(func() {
					responseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the release version is provided", func() {
			It("lists all product file for the provided product slug and release version", func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: releases,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)

				productFileResponse := pivnet.ProductFileResponse{
					ProductFile: productFile,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/products/%s/releases/%d/product_files/%d",
								apiPrefix,
								productSlug,
								releases[0].ID,
								productFile.ID,
							),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, productFileResponse),
					),
				)

				productFileCommand := commands.ProductFileCommand{
					ProductSlug:    productSlug,
					ReleaseVersion: releases[0].Version,
					ProductFileID:  productFile.ID,
				}

				err := productFileCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returned pivnet.ProductFile

				err = json.Unmarshal(outBuffer.Bytes(), &returned)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal(productFile))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})
	})

	Describe("AddProductFileCommand", func() {
		var (
			command commands.AddProductFileCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.AddProductFileCommand{
				ProductSlug:    productSlug,
				ProductFileID:  productFile.ID,
				ReleaseVersion: releases[0].Version,
			}
		})

		JustBeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("adds the product file for the provided product slug and product file id to the specified release", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("RemoveProductFileCommand", func() {
		var (
			command commands.RemoveProductFileCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.RemoveProductFileCommand{
				ProductSlug:    productSlug,
				ProductFileID:  productFile.ID,
				ReleaseVersion: releases[0].Version,
			}
		})

		JustBeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("removes the product file for the provided product slug and product file id from the specified release", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("DeleteProductFileCommand", func() {
		var (
			command commands.DeleteProductFileCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent
			response = pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			command = commands.DeleteProductFileCommand{
				ProductSlug:   productSlug,
				ProductFileID: productFile.ID,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
							productFile.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("deletes the product file for the provided product slug and product file id", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteProductFileCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})
	})
})
