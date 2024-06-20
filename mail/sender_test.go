package mail

import (
	"fmt"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	// if testing.Short() {
	// t.Skip()
	// }

	config, err := util.LoadVault("../.env.test")
	require.NotEmpty(t, config)
	require.NoError(t, err)

	fmt.Println(config)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="https://mnbenghuzzi.github.io/">MNB</a></p>
	`
	to := []string{"wajabatak@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
