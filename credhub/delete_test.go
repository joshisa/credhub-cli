package credhub_test

import (
	. "code.cloudfoundry.org/credhub-cli/credhub"
	"errors"

	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Delete", func() {

	Describe("DeleteByName", func() {
		It("requests a delete by name", func() {
			dummy := &DummyAuth{Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			}}

			ch, _ := New("https://example.com", Auth(dummy.Builder()))
			ch.DeleteByName("/example-password")

			url := dummy.Request.URL.String()
			Expect(url).To(Equal("https://example.com/api/v1/data?name=%2Fexample-password"))
			Expect(dummy.Request.Method).To(Equal(http.MethodDelete))

		})

		Context("when the credential exists", func() {
			It("deletes the credential", func() {
				dummy := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				}}

				ch, _ := New("https://example.com", Auth(dummy.Builder()))
				err := ch.DeleteByName("/example-password")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the credential does not exist", func() {
			It("returns an error", func() {
				dummy := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error":"The request could not be completed because the credential does not exist or you do not have sufficient authorization."}`)),
				}}

				ch, _ := New("https://example.com", Auth(dummy.Builder()))
				err := ch.DeleteByName("/example-password")
				Expect(err).To(MatchError("The request could not be completed because the credential does not exist or you do not have sufficient authorization."))
			})
		})
	})

	Describe("DeleteByPath", func() {
		It("finds the credentials under the path", func() {
			expectedResponse := `{
  "credentials": [
    {
      "version_created_at": "2017-05-09T21:09:26Z",
      "name": "/some/example/path/example-cred-0"
    },
    {
      "version_created_at": "2017-05-09T21:09:07Z",
      "name": "/some/example/path/example-cred-1"
    },
	{
      "version_created_at": "2017-05-09T21:09:07Z",
      "name": "/some/example/path/path2/example-cred-2"
    }
  ]
}`
			server := NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "path=/some"),
					RespondWith(http.StatusOK, expectedResponse),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/example-cred-0"),
					RespondWith(http.StatusOK, ""),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/example-cred-1"),
					RespondWith(http.StatusOK, ""),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/path2/example-cred-2"),
					RespondWith(http.StatusOK, ""),
				),
			)

			ch, _ := New(server.URL())
			failedCreds, err := ch.DeleteByPath("/some")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(server.ReceivedRequests())).To(Equal(4))
			Expect(failedCreds).To(BeEmpty())
		})

		It("returns an error when find request fails", func() {
			dummy := &DummyAuth{Error: errors.New("Network error occurred")}

			ch, _ := New("https://example.com", Auth(dummy.Builder()))

			failedCreds, err := ch.DeleteByPath("/some/example/path")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Network error occurred"))
			Expect(failedCreds).To(BeEmpty())
		})

		It("returns an error when a delete request fails", func() {
			expectedResponse := `{
  "credentials": [
    {
      "version_created_at": "2017-05-09T21:09:26Z",
      "name": "/some/example/path/example-cred-0"
    },
    {
      "version_created_at": "2017-05-09T21:09:07Z",
      "name": "/some/example/path/example-cred-1"
    },
	{
      "version_created_at": "2017-05-09T21:09:07Z",
      "name": "/some/example/path/path2/example-cred-2"
    }
  ]
}`
			server := NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "path=/some"),
					RespondWith(http.StatusOK, expectedResponse),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/example-cred-0"),
					RespondWith(http.StatusOK, ""),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/example-cred-1"),
					RespondWith(http.StatusInternalServerError, ""),
				),
				CombineHandlers(
					VerifyRequest("DELETE", "/api/v1/data", "name=/some/example/path/path2/example-cred-2"),
					RespondWith(http.StatusOK, ""),
				),
			)

			ch, _ := New(server.URL())
			failedCreds, err := ch.DeleteByPath("/some")
			Expect(len(server.ReceivedRequests())).To(Equal(4))
			Expect(err).ToNot(HaveOccurred())
			Expect(failedCreds).To(HaveLen(1))
			Expect(failedCreds[0].Path).To(Equal("/some/example/path/example-cred-1"))
			Expect(failedCreds[0].Err.Error()).To(Equal("The response body could not be decoded: unexpected end of JSON input"))
		})
	})
})
