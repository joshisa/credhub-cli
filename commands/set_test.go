package commands_test

import (
	"net/http"

	"fmt"

	"runtime"

	"os"

	"strings"

	"github.com/cloudfoundry-incubator/credhub-cli/commands"
	"github.com/cloudfoundry-incubator/credhub-cli/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Set", func() {
	BeforeEach(func() {
		login()
	})

	ItRequiresAuthentication("set", "-n", "test-credential", "-t", "password", "-w", "value")
	ItAutomaticallyLogsIn("PUT", "set", "-n", "test-credential", "-t", "password", "-w", "test-value")

	Describe("not specifying type", func() {
		It("returns an error", func() {
			session := runCommand("set", "-n", "my-password", "-w", "potatoes")

			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("A type must be specified when setting a credential. Valid types include 'value', 'json', 'password', 'user', 'certificate', 'ssh' and 'rsa'."))
		})
	})

	Describe("setting value secrets", func() {
		It("puts a secret using explicit value type", func() {
			SetupPutValueServer("my-value", "value", "potatoes")

			session := runCommand("set", "-n", "my-value", "-v", "potatoes", "-t", "value")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyValuePotatoesYaml))
		})

		It("escapes special characters in the value", func() {
			SetupPutValueServer("my-character-test", "value", `{\"password\":\"some-still-bad-password\"}`)

			session := runCommand("set", "-t", "value", "-n", "my-character-test", "-v", `{"password":"some-still-bad-password"}`)

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySpecialCharacterValue))
		})

		It("puts a secret using explicit value type and returns in json format", func() {
			SetupPutValueServer("my-value", "value", "potatoes")

			session := runCommand("set", "-n", "my-value", "-v", "potatoes", "-t", "value", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyValuePotatoesJson))
		})

		It("accepts case-insensitive type", func() {
			SetupPutValueServer("my-value", "value", "potatoes")

			session := runCommand("set", "-n", "my-value", "-v", "potatoes", "-t", "VALUE", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyValuePotatoesJson))
		})
	})

	Describe("setting json secrets", func() {
		It("puts a secret using explicit json type", func() {
			jsonValue := `{"foo":"bar","nested":{"a":1},"an":["array"]}`
			setupPutJsonServer("json-secret", jsonValue)

			session := runCommand("set", "-n", "json-secret", "-v", jsonValue, "-t", "json")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyJsonFormatYaml))
		})

		It("escapes special characters in the json", func() {
			setupPutJsonServer("my-character-test", `{"foo":"b\"ar"}`)

			session := runCommand("set", "-t", "json", "-n", "my-character-test", "-v", `{"foo":"b\"ar"}`)

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySpecialCharacterJson))
		})

		It("puts a secret using explicit json type and returns in json format", func() {
			jsonValue := `{"foo":"bar","nested":{"a":1},"an":["array"]}`
			setupPutJsonServer("json-secret", jsonValue)

			session := runCommand("set", "-n", "json-secret", "-v", jsonValue, "-t", "json", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyJsonFormatJson))
		})

		It("accepts case-insensitive type", func() {
			jsonValue := `{"foo":"bar","nested":{"a":1},"an":["array"]}`
			setupPutJsonServer("json-secret", jsonValue)

			session := runCommand("set", "-n", "json-secret", "-v", jsonValue, "-t", "JSON")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyJsonFormatYaml))
		})
	})

	Describe("setting SSH secrets", func() {
		It("puts a secret using explicit ssh type", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-ssh-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "ssh")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySSHFooYaml))
		})

		It("puts a secret using values read from files", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", "some-public-key", "some-private-key", true)

			tempDir := test.CreateTempDir("sshFilesForTesting")
			publicFileName := test.CreateCredentialFile(tempDir, "rsa.pub", "some-public-key")
			privateFilename := test.CreateCredentialFile(tempDir, "rsa.key", "some-private-key")

			session := runCommand("set", "-n", "foo-ssh-key",
				"-t", "ssh",
				"-u", publicFileName,
				"-p", privateFilename)

			os.RemoveAll(tempDir)
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySSHFooYaml))
		})

		It("puts a secret specifying no-overwrite", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", "some-public-key", "some-private-key", false)

			session := runCommand("set", "-n", "foo-ssh-key", "-t", "ssh", "-u", "some-public-key", "-p", "some-private-key", "--no-overwrite")

			Eventually(session).Should(Exit(0))
		})

		It("puts a secret using explicit ssh type and returns in json format", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-ssh-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "ssh", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMySSHFooJson))
		})

		It("accepts case-insensitive type", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-ssh-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "SSH")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySSHFooYaml))
		})

		It("handles newline characters", func() {
			SetupPutRsaSshServer("foo-ssh-key", "ssh", `some\npublic\nkey`, `some\nprivate\nkey`, true)
			session := runCommand("set", "-n", "foo-ssh-key", "-u", `some\npublic\nkey`, "-p", `some\nprivate\nkey`, "-t", "ssh", "--output-json")

			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).Should(MatchJSON(responseMySSHWithNewlinesJson))
		})
	})

	Describe("setting RSA secrets", func() {
		It("puts a secret using explicit rsa type", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-rsa-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "rsa")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyRSAFooYaml))
		})

		It("puts a secret using values read from files", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", "some-public-key", "some-private-key", true)

			tempDir := test.CreateTempDir("rsaFilesForTesting")
			publicFileName := test.CreateCredentialFile(tempDir, "rsa.pub", "some-public-key")
			privateFilename := test.CreateCredentialFile(tempDir, "rsa.key", "some-private-key")

			session := runCommand("set", "-n", "foo-rsa-key",
				"-t", "rsa",
				"-u", publicFileName,
				"-p", privateFilename)

			os.RemoveAll(tempDir)
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyRSAFooYaml))
		})

		It("puts a secret specifying no-overwrite", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", "some-public-key", "some-private-key", false)

			session := runCommand("set", "-n", "foo-rsa-key", "-t", "rsa", "-u", "some-public-key", "-p", "some-private-key", "--no-overwrite")

			Eventually(session).Should(Exit(0))
		})

		It("puts a secret using explicit rsa type and returns in json format", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-rsa-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "rsa", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyRSAFooJson))
		})

		It("accepts case-insensitive type", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", "some-public-key", "some-private-key", true)

			session := runCommand("set", "-n", "foo-rsa-key", "-u", "some-public-key", "-p", "some-private-key", "-t", "RSA")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyRSAFooYaml))
		})

		It("handles newline characters", func() {
			SetupPutRsaSshServer("foo-rsa-key", "rsa", `some\npublic\nkey`, `some\nprivate\nkey`, true)
			session := runCommand("set", "-n", "foo-rsa-key", "-u", `some\npublic\nkey`, "-p", `some\nprivate\nkey`, "-t", "rsa", "--output-json")

			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).Should(MatchJSON(responseMyRSAWithNewlinesJson))
		})
	})

	Describe("setting password secrets", func() {
		It("puts a secret specifying no-overwrite", func() {
			SetupOverwritePutValueServer("my-password", "password", "potatoes", false)

			session := runCommand("set", "-n", "my-password", "-t", "password", "-w", "potatoes", "--no-overwrite")

			Eventually(session).Should(Exit(0))
		})

		It("puts a secret using explicit password type  and returns in yaml format", func() {
			SetupPutValueServer("my-password", "password", "potatoes")

			session := runCommand("set", "-n", "my-password", "-w", "potatoes", "-t", "password")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyPasswordPotatoesYaml))
		})

		It("prompts for value if value is not provided", func() {
			SetupPutValueServer("my-password", "password", "potatoes")

			session := runCommandWithStdin(strings.NewReader("potatoes\n"), "set", "-n", "my-password", "-t", "password")

			Eventually(session.Out).Should(Say("password:"))
			Eventually(session.Wait("10s").Out).Should(Say(responseMyPasswordPotatoesYaml))
			Eventually(session).Should(Exit(0))
		})

		It("can set password that contains spaces interactively", func() {
			SetupPutValueServer("my-password", "password", "potatoes potatoes")

			session := runCommandWithStdin(strings.NewReader("potatoes potatoes\n"), "set", "-n", "my-password", "-t", "password")

			response := fmt.Sprintf(STRING_CREDENTIAL_RESPONSE_YAML, "my-password", "password", "potatoes potatoes")

			Eventually(session.Out).Should(Say("password:"))
			Eventually(session.Wait("10s").Out).Should(Say(response))
			Eventually(session).Should(Exit(0))
		})

		It("escapes special characters in the password", func() {
			SetupPutValueServer("my-character-test", "password", `{\"password\":\"some-still-bad-password\"}`)

			session := runCommand("set", "-t", "password", "-n", "my-character-test", "-w", `{"password":"some-still-bad-password"}`)

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMySpecialCharacterPassword))
		})

		It("puts a secret using explicit password type and returns in json format", func() {
			SetupPutValueServer("my-password", "password", "potatoes")

			session := runCommand("set", "-n", "my-password", "-w", "potatoes", "-t", "password", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyPasswordPotatoesJson))
		})

		It("accepts case-insensitive type", func() {
			SetupPutValueServer("my-password", "password", "potatoes")

			session := runCommand("set", "-n", "my-password", "-w", "potatoes", "-t", "PASSWORD")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyPasswordPotatoesYaml))
		})
	})

	Describe("setting certificate secrets", func() {
		It("puts a secret using explicit certificate type and string values", func() {
			SetupPutCertificateServer("my-secret", "my-ca", "my-cert", "my-priv")

			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--root", "my-ca",
				"--certificate", "my-cert", "--private", "my-priv")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyCertificateYaml))
		})

		It("puts a secret using explicit certificate type, string values, and certificate authority name", func() {
			SetupPutCertificateWithCaNameServer("my-secret", "my-ca", "my-cert", "my-priv")

			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--ca-name", "my-ca",
				"--certificate", "my-cert", "--private", "my-priv")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyCertificateWithNamedCAYaml))
		})

		It("puts a secret using explicit certificate type and string values with no-overwrite", func() {
			SetupOverwritePutCertificateServer("my-secret", "my-ca", "my-cert", "my-priv", false)

			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--root", "my-ca",
				"--certificate", "my-cert", "--private", "my-priv", "--no-overwrite")

			Eventually(session).Should(Exit(0))
		})

		It("puts a secret using explicit certificate type and values read from files", func() {
			SetupPutCertificateServer("my-secret", "my-ca", "my-cert", "my-priv")
			tempDir := test.CreateTempDir("certFilesForTesting")
			caFilename := test.CreateCredentialFile(tempDir, "ca.txt", "my-ca")
			certificateFilename := test.CreateCredentialFile(tempDir, "certificate.txt", "my-cert")
			privateFilename := test.CreateCredentialFile(tempDir, "private.txt", "my-priv")

			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--root", caFilename,
				"--certificate", certificateFilename, "--private", privateFilename)

			os.RemoveAll(tempDir)
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyCertificateYaml))
		})

		It("fails to put a secret when reading from unreadable file", func() {
			testSetFileFailure("unreadable.txt", "", "")
			testSetFileFailure("", "unreadable.txt", "")
			testSetFileFailure("", "", "unreadable.txt")
		})

		It("puts a secret using explicit certificate type and string values in json format", func() {
			SetupPutCertificateServer("my-secret", "my-ca", "my-cert", "my-priv")

			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--root", "my-ca",
				"--certificate", "my-cert", "--private", "my-priv", "--output-json")

			Eventually(session).Should(Exit(0))
			Eventually(string(session.Out.Contents())).Should(MatchJSON(responseMyCertificateJson))
		})

		It("accepts case insensitive type", func() {
			SetupPutCertificateServer("my-secret", "my-ca", "my-cert", "my-priv")

			session := runCommand("set", "-n", "my-secret",
				"-t", "CERTIFICATE", "--root", "my-ca",
				"--certificate", "my-cert", "--private", "my-priv")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseMyCertificateYaml))
		})

		It("handles newline characters", func() {
			SetupPutCertificateServer("my-secret", `my\nca`, `my\ncert`, `my\npriv`)
			session := runCommand("set", "-n", "my-secret",
				"-t", "certificate", "--root", `my\nca`,
				"--certificate", `my\ncert`, "--private", `my\npriv`, "--output-json")
			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).Should(MatchJSON(responseMyCertificateWithNewlinesJson))
		})
	})

	Describe("setting User secrets", func() {
		It("puts a secret using explicit user type", func() {
			SetupPutUserServer("my-username-credential", `{"username": "my-username", "password": "test-password"}`, "my-username", "test-password", "passw0rd-H4$h", true)

			session := runCommand("set", "-n", "my-username-credential", "-z", "my-username", "-w", "test-password", "-t", "user")

			Eventually(session).Should(Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring(responseMyUsernameYaml))
		})

		It("puts a secret specifying no-overwrite", func() {
			SetupPutUserServer("my-username-credential", `{"username": "my-username", "password": "test-password"}`, "my-username", "test-password", "passw0rd-H4$h", false)

			session := runCommand("set", "-n", "my-username-credential", "-t", "user", "-z", "my-username", "-w", "test-password", "--no-overwrite")

			Eventually(session).Should(Exit(0))
		})

		It("should set password interactively for user", func() {
			SetupPutUserServer("my-username-credential", `{"username": "my-username", "password": "test-password"}`, "my-username", "test-password", "passw0rd-H4$h", true)

			session := runCommandWithStdin(strings.NewReader("test-password\n"), "set", "-n", "my-username-credential", "-t", "user", "--username", "my-username")

			response := fmt.Sprintf(USER_CREDENTIAL_RESPONSE_YAML, "my-username-credential", "test-password", "passw0rd-H4$h", "my-username")

			Eventually(session.Out).Should(Say("password:"))
			Eventually(session.Wait("10s").Out.Contents()).Should(ContainSubstring(response))
			Eventually(session).Should(Exit(0))
		})

		It("should set null username when it isn't provided", func() {
			SetupPutUserWithoutUsernameServer("my-username-credential", `{"password": "test-password"}`, "test-password", "passw0rd-H4$h", true)

			session := runCommandWithStdin(strings.NewReader("test-password\n"), "set", "-n", "my-username-credential", "-t", "user")

			response := fmt.Sprintf(USER_WITHOUT_USERNAME_CREDENTIAL_RESPONSE_YAML, "my-username-credential", "test-password", "passw0rd-H4$h")

			Eventually(session.Out).Should(Say("password:"))
			Eventually(session.Wait("10s").Out.Contents()).Should(ContainSubstring(response))
			Eventually(session).Should(Exit(0))
		})

		It("puts a secret using explicit user type in json format", func() {
			SetupPutUserServer("my-username-credential", `{"username": "my-username", "password": "test-password"}`, "my-username", "test-password", "passw0rd-H4$h", true)

			session := runCommand("set", "-n", "my-username-credential", "-z", "my-username", "-w", "test-password", "-t", "user",
				"--output-json")

			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).Should(MatchJSON(responseMyUsernameJson))
		})

		It("accepts case-insensitive type", func() {
			SetupPutUserServer("my-username-credential", `{"username": "my-username", "password": "test-password"}`, "my-username", "test-password", "passw0rd-H4$h", true)

			session := runCommand("set", "-n", "my-username-credential", "-z", "my-username", "-w", "test-password", "-t", "USER")

			Eventually(session).Should(Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring(responseMyUsernameYaml))
		})
	})

	Describe("Help", func() {
		It("short flags", func() {
			Expect(commands.SetCommand{}).To(SatisfyAll(
				commands.HaveFlag("name", "n"),
				commands.HaveFlag("type", "t"),
				commands.HaveFlag("value", "v"),
				commands.HaveFlag("no-overwrite", "O"),
				commands.HaveFlag("root", "r"),
				commands.HaveFlag("certificate", "c"),
				commands.HaveFlag("private", "p"),
			))
		})

		ItBehavesLikeHelp("set", "s", func(session *Session) {
			Expect(session.Err).To(Say("set"))
			Expect(session.Err).To(Say("name"))
			Expect(session.Err).To(Say("credential"))
		})

		It("displays missing 'n' option as required parameter", func() {
			session := runCommand("set", "-v", "potatoes")

			Eventually(session).Should(Exit(1))
			if runtime.GOOS == "windows" {
				Expect(session.Err).To(Say("the required flag `/n, /name' was not specified"))
			} else {
				Expect(session.Err).To(Say("the required flag `-n, --name' was not specified"))
			}
		})

		It("displays the server provided error when an error is received", func() {
			server.AppendHandlers(
				RespondWith(http.StatusBadRequest, `{"error": "test error"}`),
			)

			session := runCommand("set", "-n", "my-value", "-t", "value", "-v", "tomatoes")

			Eventually(session).Should(Exit(1))

			Expect(session.Err).To(Say("test error"))
		})
	})
})

func SetupPutRsaSshServer(name, keyType, publicKey, privateKey string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(RSA_SSH_CREDENTIAL_REQUEST_JSON, keyType, name, publicKey, privateKey, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(RSA_SSH_CREDENTIAL_RESPONSE_JSON, keyType, name, publicKey, privateKey)),
		),
	)
}

func SetupPutValueServer(name, credentialType, value string) {
	SetupOverwritePutValueServer(name, credentialType, value, true)
}

func SetupOverwritePutValueServer(name, credentialType, value string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(STRING_CREDENTIAL_OVERWRITE_REQUEST_JSON, credentialType, name, value, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(STRING_CREDENTIAL_RESPONSE_JSON, credentialType, name, value)),
		),
	)
}

func setupPutJsonServer(name, value string) {
	setupOverwritePutJsonServer(name, value, true)
}

func setupOverwritePutJsonServer(name, value string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(JSON_CREDENTIAL_OVERWRITE_REQUEST_JSON, name, value, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(JSON_CREDENTIAL_RESPONSE_JSON, name, value)),
		),
	)
}

func SetupPutCertificateServer(name, ca, cert, priv string) {
	SetupOverwritePutCertificateServer(name, ca, cert, priv, true)
}

func SetupPutCertificateWithCaNameServer(name, caName, cert, priv string) {
	SetupOverwritePutCertificateWithCaNameServer(name, caName, cert, priv, true)
}

func SetupOverwritePutCertificateServer(name, ca, cert, priv string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(CERTIFICATE_CREDENTIAL_REQUEST_JSON, name, ca, cert, priv, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_JSON, name, ca, cert, priv)),
		),
	)
}

func SetupOverwritePutCertificateWithCaNameServer(name, caName, cert, priv string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(CERTIFICATE_CREDENTIAL_WITH_NAMED_CA_REQUEST_JSON, name, caName, cert, priv, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_JSON, name, "known-ca-value", cert, priv)),
		),
	)
}

func SetupPutUserServer(name, value, username, password, passwordHash string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(USER_SET_CREDENTIAL_REQUEST_JSON, name, value, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(USER_CREDENTIAL_RESPONSE_JSON, name, username, password, passwordHash)),
		),
	)
}

func SetupPutUserWithoutUsernameServer(name, value, password, passwordHash string, overwrite bool) {
	var jsonRequest string
	jsonRequest = fmt.Sprintf(USER_SET_CREDENTIAL_REQUEST_JSON, name, value, overwrite)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(jsonRequest),
			RespondWith(http.StatusOK, fmt.Sprintf(USER_WITHOUT_USERNAME_CREDENTIAL_RESPONSE_JSON, name, password, passwordHash)),
		),
	)
}

func SetupPutBadRequestServer(body string) {
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest("PUT", "/api/v1/data"),
			VerifyJSON(body),
			RespondWith(http.StatusBadRequest, `{"error":"test error"}`),
		),
	)
}

func testSetFileFailure(caFilename, certificateFilename, privateFilename string) {
	tempDir := test.CreateTempDir("certFilesForTesting")
	if caFilename == "unreadable.txt" {
		caFilename = test.CreateCredentialFile(tempDir, caFilename, "my-ca")
		err := os.Chmod(caFilename, 0222)
		Expect(err).To(BeNil())
	}
	if certificateFilename == "unreadable.txt" {
		certificateFilename = test.CreateCredentialFile(tempDir, certificateFilename, "my-cert")
		err := os.Chmod(certificateFilename, 0222)
		Expect(err).To(BeNil())
	}
	if privateFilename == "unreadable.txt" {
		privateFilename = test.CreateCredentialFile(tempDir, privateFilename, "my-priv")
		err := os.Chmod(privateFilename, 0222)
		Expect(err).To(BeNil())
	}

	session := runCommand("set", "-n", "my-secret",
		"-t", "certificate", "--root", caFilename,
		"--certificate", certificateFilename, "--private", privateFilename)

	os.RemoveAll(tempDir)
	Eventually(session).Should(Exit(1))
	Eventually(session.Err).Should(Say("A referenced file could not be opened. Please validate the provided filenames and permissions, then retry your request."))
}
