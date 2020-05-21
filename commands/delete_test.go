package commands_test

import (
	"code.cloudfoundry.org/credhub-cli/commands"
	"net/http"

	"runtime"

	"code.cloudfoundry.org/credhub-cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Delete", func() {
	BeforeEach(func() {
		login()
	})

	ItRequiresAuthentication("delete", "-n", "test-credential")
	ItRequiresAnAPIToBeSet("delete", "-n", "test-credential")

	testAutoLogIns := []TestAutoLogin{
		{
			method:              "DELETE",
			responseFixtureFile: "delete_test.json",
			responseStatus:      http.StatusOK,
			endpoint:            "/api/v1/data",
		},
	}
	ItAutomaticallyLogsIn(testAutoLogIns, "delete", "-n", "test-credential")

	Describe("Help", func() {
		ItBehavesLikeHelp("delete", "d", func(session *Session) {
			Expect(session.Err).To(Say("Usage"))
			if runtime.GOOS == "windows" {
				Expect(session.Err).To(Say("credhub-cli.exe \\[OPTIONS\\] delete \\[delete-OPTIONS\\]"))
			} else {
				Expect(session.Err).To(Say("credhub-cli \\[OPTIONS\\] delete \\[delete-OPTIONS\\]"))
			}
		})

		It("short flags", func() {
			Expect(commands.DeleteCommand{}).To(SatisfyAll(
				commands.HaveFlag("name", "n"),
				commands.HaveFlag("path", "p"),
			))
		})
	})

	It("deletes a secret by name", func() {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest("DELETE", "/api/v1/data", "name=my-secret"),
				RespondWith(http.StatusOK, ""),
			),
		)

		session := runCommand("delete", "-n", "my-secret")

		Eventually(session).Should(Exit(0))
		Eventually(session.Out).Should(Say("Credential successfully deleted"))
	})

	It("deletes secrets by path", func() {
		responseJSON := `{
					"credentials": [
							{
								"name": "deploy123/dan.password",
								"version_created_at": "2016-09-06T23:26:58Z"
							},
							{
								"name": "deploy123/dan.key",
								"version_created_at": "2016-09-06T23:26:58Z"
							}
					]
				}`
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest("GET", "/api/v1/data", "path=deploy123"),
				RespondWith(http.StatusOK, responseJSON),
			),
			CombineHandlers(
				VerifyRequest("DELETE", "/api/v1/data", "name=deploy123/dan.password"),
				RespondWith(http.StatusOK, ""),
			),
			CombineHandlers(
				VerifyRequest("DELETE", "/api/v1/data", "name=deploy123/dan.key"),
				RespondWith(http.StatusOK, ""),
			),
		)

		session := runCommand("delete", "-p", "deploy123")

		Eventually(session).Should(Exit(0))
		// first 2 requests that the server receives result from test BeforeEach's login
		Expect(server.ReceivedRequests()[2].URL.RawQuery).To(Equal("path=deploy123"))
		Expect(server.ReceivedRequests()[3].URL.RawQuery).To(Equal("name=deploy123%2Fdan.password"))
		Expect(server.ReceivedRequests()[4].URL.RawQuery).To(Equal("name=deploy123%2Fdan.key"))
		Eventually(session.Out).Should(Say("Successfully deleted credential: deploy123/dan.password"))
		Eventually(session.Out).Should(Say("Successfully deleted credential: deploy123/dan.key"))
	})

	Describe("Errors", func() {
		It("requires a name or a path", func() {
			session := runCommand("delete")
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("A name or path must be provided. Please update and retry your request."))
		})

		It("prints an error when the network request fails", func() {
			cfg := config.ReadConfig()
			cfg.ApiURL = "mashed://potatoes"
			config.WriteConfig(cfg)

			session := runCommand("delete", "-n", "my-secret")

			Eventually(session).Should(Exit(1))
			Eventually(string(session.Err.Contents())).Should(ContainSubstring("unsupported protocol scheme"))
		})
	})
})
