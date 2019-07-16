package credhub_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/credhub-cli/credhub"
)

var _ = Describe("Certificates", func() {
	It("requests to get all certificates", func() {
		dummy := &DummyAuth{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}}

		ch, _ := New("https://example.com", Auth(dummy.Builder()))
		_, _ = ch.GetAllCertificatesMetadata()
		url := dummy.Request.URL.String()
		Expect(url).To(Equal("https://example.com/api/v1/certificates/"))
		Expect(dummy.Request.Method).To(Equal(http.MethodGet))

	})

	Context("getting certificate metadata", func() {
		Describe("when there is data returned", func() {
			It("marshals it properly", func() {
				responseString :=
					`{"certificates": [
				{
				  "id": "some-id",
				  "name": "/some-cert",
				  "signed_by": "/some-cert",
				  "signs": ["/another-cert"],
				  "versions": [
					{
					  "expiry_date": "2020-05-29T12:33:50Z",
					  "id": "some-version-id",
					  "transitional": false,
                      "self_signed": true,
                      "certificate_authority": true,
                      "generated": true
					},
					{
					  "expiry_date": "2020-05-29T12:33:50Z",
					  "id": "some-other-version-id",
					  "transitional": true,
                      "self_signed": true,
                      "certificate_authority": true
					}
				  ]
				},
				{
				  "id": "some-other-id",
				  "name": "/some-other-cert",
				  "signed_by": "/some-cert",
				  "signs": [],
				  "versions": [
					{
					  "expiry_date": "2020-05-29T12:33:50Z",
					  "id": "some-other-other-version-id",
					  "transitional": false,
                      "self_signed": false,
                      "certificate_authority": false,
                      "generated": false
					}
				  ]
				}
			]}`
				dummyAuth := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
				}}

				ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))

				metadata, err := ch.GetAllCertificatesMetadata()

				t := new(bool)
				f := new(bool)
				*t = true
				*f = false

				Expect(err).To(BeNil())
				Expect(metadata[0].Id).To(Equal("some-id"))
				Expect(metadata[0].Name).To(Equal("/some-cert"))
				Expect(metadata[0].SignedBy).To(Equal("/some-cert"))
				Expect(metadata[0].Signs[0]).To(Equal("/another-cert"))
				Expect(metadata[0].Versions[0].Id).To(Equal("some-version-id"))
				Expect(metadata[0].Versions[0].ExpiryDate).To(Equal("2020-05-29T12:33:50Z"))
				Expect(metadata[0].Versions[0].Transitional).To(BeFalse())
				Expect(metadata[0].Versions[0].CertificateAuthority).To(BeTrue())
				Expect(metadata[0].Versions[0].SelfSigned).To(BeTrue())
				Expect(metadata[0].Versions[0].Generated).To(Equal(t))
				Expect(metadata[0].Versions[1].Id).To(Equal("some-other-version-id"))
				Expect(metadata[0].Versions[1].ExpiryDate).To(Equal("2020-05-29T12:33:50Z"))
				Expect(metadata[0].Versions[1].Transitional).To(BeTrue())
				Expect(metadata[0].Versions[1].CertificateAuthority).To(BeTrue())
				Expect(metadata[0].Versions[1].SelfSigned).To(BeTrue())
				Expect(metadata[0].Versions[1].Generated).To(BeNil())
				Expect(metadata[1].Id).To(Equal("some-other-id"))
				Expect(metadata[1].Name).To(Equal("/some-other-cert"))
				Expect(metadata[1].SignedBy).To(Equal("/some-cert"))
				Expect(metadata[1].Signs).To(Equal([]string{}))
				Expect(metadata[1].Versions[0].Id).To(Equal("some-other-other-version-id"))
				Expect(metadata[1].Versions[0].ExpiryDate).To(Equal("2020-05-29T12:33:50Z"))
				Expect(metadata[1].Versions[0].Transitional).To(BeFalse())
				Expect(metadata[1].Versions[0].CertificateAuthority).To(BeFalse())
				Expect(metadata[1].Versions[0].SelfSigned).To(BeFalse())
				Expect(metadata[1].Versions[0].Generated).To(Equal(f))
			})
		})

		Describe("when no certificates are returned", func() {
			It("returns empty array", func() {
				responseString :=
					`{"certificates": []}`

				dummyAuth := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
				}}

				ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))
				metadata, err := ch.GetAllCertificatesMetadata()

				Expect(err).To(BeNil())
				Expect(len(metadata)).To(Equal(0))
			})
		})

		Describe("when certificates key is missing", func() {
			It("returns empty array", func() {
				responseString :=
					`{}`

				dummyAuth := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
				}}

				ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))
				metadata, err := ch.GetAllCertificatesMetadata()

				Expect(err).To(BeNil())
				Expect(len(metadata)).To(Equal(0))
			})
		})

		Describe("when response is empty", func() {
			It("returns error", func() {
				responseString := ``

				dummyAuth := &DummyAuth{Response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(responseString)),
				}}

				ch, _ := New("https://example.com", Auth(dummyAuth.Builder()))
				metadata, err := ch.GetAllCertificatesMetadata()

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("The response body could not be decoded"))
				Expect(metadata).To(BeNil())
			})
		})
	})
})
