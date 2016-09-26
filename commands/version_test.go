package commands_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Version", func() {
	Context("when the request succeeds", func() {
		BeforeEach(func() {
			responseJson := `{"app":{"name":"Pivotal Credential Manager","version":"0.2.0"}}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/info"),
					RespondWith(http.StatusOK, responseJson),
				),
			)
		})

		It("displays the version with --version", func() {
			session := runCommand("--version")

			Eventually(session).Should(Exit(0))
			sout := string(session.Out.Contents())
			testVersion(sout)
			Expect(sout).To(ContainSubstring("CM Version: 0.2.0"))
		})
	})

	Context("when the request fails", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/info"),
					RespondWith(http.StatusNotFound, ""),
				),
			)
		})

		It("displays the version with --version", func() {
			session := runCommand("--version")

			Eventually(session).Should(Exit(0))
			sout := string(session.Out.Contents())
			testVersion(sout)
			Expect(sout).To(ContainSubstring("CM Version: Not Found"))
		})
	})

})

func testVersion(sout string) {
	Expect(sout).To(ContainSubstring("CLI Version: 0.2.0"))
	Expect(sout).ToNot(ContainSubstring("build DEV"))
}
