package credhub_test

import (
	. "code.cloudfoundry.org/credhub-cli/credhub"
	"github.com/onsi/gomega/ghttp"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Permissions", func() {
//	Context("GetPermissions", func() {
//		It("returns the permissions", func() {
//			responseString :=
//				`{
//	"actor":"user:A",
//	"operations":["read"],
//	"path":"/example-password",
//	"uuid":"1234"
//}`
//
//			dummyAuth := &DummyAuth{Response: &http.Response{
//				StatusCode: http.StatusOK,
//				Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
//			}}
//
//			ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))
//			actualPermissions, err := ch.GetPermission("1234")
//			Expect(err).NotTo(HaveOccurred())
//
//			expectedPermission := permissions.Permission{
//				Actor:      "user:A",
//				Operations: []string{"read"},
//				Path:       "/example-password",
//				UUID:       "1234",
//			}
//			Expect(actualPermissions).To(Equal(&expectedPermission))
//
//			By("calling the right endpoints")
//			url := dummyAuth.Request.URL.String()
//			Expect(url).To(Equal("https://example.com/api/v2/permissions/1234"))
//			Expect(dummyAuth.Request.Method).To(Equal(http.MethodGet))
//		})
//	})

	Context("GetPermissions", func() {
		Context("when server version is less than 2.0.0", func() {
			Context("when permission exists", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/info"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/version"),
							ghttp.RespondWith(http.StatusOK, `{"version": "1.9.0"}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/permissions?credential_name=/example-password"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				It("returns permission", func() {
					ch, _ := New(server.URL())
					_, err := ch.GetPermission("name")
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})
		})

		Context("when the server version is greater than or equal to 2.0.0", func() {
			Context("when permissions exists", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/info"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/version"),
							ghttp.RespondWith(http.StatusOK, `{"version": "2.0.0"}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v2/permissions/1234"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				It("returns permission", func() {
					ch, _ := New(server.URL())
					_, err := ch.GetPermission("1234")
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})
		})
	})

	Context("AddPermissions", func() {
		Context("when server version is less than 2.0.0", func() {
			Context("when a credential exists", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/info"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/version"),
							ghttp.RespondWith(http.StatusOK, `{"version": "1.9.0"}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/permissions"),
							ghttp.VerifyJSON(`{
	"credential_name": "/example-password",
	"permissions": [
	{
	"actor": "some-actor",
	"operations": ["read", "write"],
	"path":"", "uuid":""
	}]
	}`),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				It("can add permissions to it", func() {
					ch, _ := New(server.URL())
					_, err := ch.AddPermission("/example-password", "some-actor", []string{"read", "write"})
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})

			//Context("when a credential doesn't exist", func() {
			//	It("cannot add permissions to it", func() {
			//		dummy := &DummyAuth{Response: &http.Response{
			//			StatusCode: http.StatusNotFound,
			//			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error":"The request could not be completed because the credential does not exist or you do not have sufficient authorization."}`)),
			//		}}
			//		ch, _ := New("https://example.com", Auth(dummy.Builder()))
			//
			//		_, err := ch.AddPermission("/example-password", "some-actor", []string{"read", "write"})
			//
			//		Expect(err).To(MatchError(ContainSubstring("The request could not be completed because the credential does not exist or you do not have sufficient authorization.")))
			//	})
			//})
		})

		Context("when the server version is greater than or equal to 2.0.0", func() {
			Context("when a credential exists", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/info"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/version"),
							ghttp.RespondWith(http.StatusOK, `{"version": "2.0.0"}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v2/permissions"),
							ghttp.VerifyJSON(`{
	"actor": "some-actor",
	"operations":["read", "write"],
	"path":"/example-password"
	}`),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				It("can add permissions to it", func() {
					ch, _ := New(server.URL())
					_, err := ch.AddPermission("/example-password", "some-actor", []string{"read", "write"})
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})

			//Context("when a credential doesn't exist", func() {
			//	It("cannot add permissions to it", func() {
			//		dummy := &DummyAuth{Response: &http.Response{
			//			StatusCode: http.StatusNotFound,
			//			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error":"The request could not be completed because the credential does not exist or you do not have sufficient authorization."}`)),
			//		}}
			//		ch, _ := New("https://example.com", Auth(dummy.Builder()))
			//
			//		_, err := ch.AddPermission("/example-password", "some-actor", []string{"read", "write"})
			//
			//		Expect(err).To(MatchError(ContainSubstring("The request could not be completed because the credential does not exist or you do not have sufficient authorization.")))
			//	})
			//})
		})
	})
})

