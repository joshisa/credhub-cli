package credhub_test

import (
	. "code.cloudfoundry.org/credhub-cli/credhub"
	"github.com/onsi/gomega/ghttp"

	"bytes"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permissions", func() {
	Context("GetPermissions", func() {
		It("returns the permissions", func() {
			responseString :=
				`{
	"actor":"user:A",
	"operations":["read"],
	"path":"/example-password",
	"uuid":"1234"
}`

			dummyAuth := &DummyAuth{Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
			}}

			ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))
			actualPermissions, err := ch.GetPermission("1234")
			Expect(err).NotTo(HaveOccurred())

			expectedPermission := permissions.Permission{
				Actor:      "user:A",
				Operations: []string{"read"},
				Path:       "/example-password",
				UUID:       "1234",
			}
			Expect(actualPermissions).To(Equal(&expectedPermission))

			By("calling the right endpoints")
			url := dummyAuth.Request.URL.String()
			Expect(url).To(Equal("https://example.com/api/v2/permissions/1234"))
			Expect(dummyAuth.Request.Method).To(Equal(http.MethodGet))
		})
	})

	Context("AddV2Permissions", func() {
		Context("when a credential exists", func() {
			It("can add permissions to it", func() {
				responseString :=
					`{
	"actor":"user:B",
	"operations":["read"],
	"path":"/example-password"
}`
				dummy := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
				}}
				ch, _ := New("https://example.com", Auth(dummy.Builder()))

				_, err := ch.AddV2Permission("/example-password", "user:A", []string{"read_acl", "write_acl"})
				Expect(err).NotTo(HaveOccurred())

				By("calling the right endpoints")
				url := dummy.Request.URL.String()
				Expect(url).To(Equal("https://example.com/api/v2/permissions"))
				Expect(dummy.Request.Method).To(Equal(http.MethodPost))
				params, err := ioutil.ReadAll(dummy.Request.Body)
				Expect(err).NotTo(HaveOccurred())

				expectedParams := `{
				"actor": "user:A",
				"operations": ["read_acl", "write_acl"],
				"path": "/example-password"
			}`
				Expect(params).To(MatchJSON(expectedParams))
			})
		})
	})

	Context("AddV1Permissions", func() {
		Context("when a credential exists", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v1/permissions"),
						ghttp.RespondWith(http.StatusOK, `{
	"credential_name": "/example-password",
	"permissions": [
		{
	  		"actor": "uaa-user:e3366b5c-1c5a-4df8-8b0f-9001ee5a0cf0",
	  		"operations": ["read"]
		}
	]
}`),
					),
				)
			})

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			It("can add permissions to it", func() {

				ch, _ := New(server.URL(), ServerVersion("1.9.0"))
				_, err := ch.AddV1Permissions("/example-password", []permissions.Permission{
					{
						Actor:      "uaa-user:e3366b5c-1c5a-4df8-8b0f-9001ee5a0cf0",
						Operations: []string{"read"},
						Path:       "/example-password",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("AddPermission", func() {
		Context("when a credential exists", func() {
			var server *ghttp.Server

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			Context("When the server version is not provided", func() {
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
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				It("sends a request for the server version", func() {
					ch, _ := New(server.URL())
					_, err := ch.AddPermission("/example-certificate", "user:A", []string{"operation1","operation2"})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the server version is older than 2.x", func() {
			It("sends the v1 model request", func(){
				dummy := &DummyAuth{Response: &http.Response{
					Body: ioutil.NopCloser(bytes.NewBufferString("")),
				}}
			})
		})

		Context("when the server version is 2.x or later", func() {

		})
	})
})
