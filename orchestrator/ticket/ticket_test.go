package ticket

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_JiraCloudSample(t *testing.T) {
	sampleJSON := `{
  "ticket_id": "FWS-67484",
  "summary": "DENT update: Veritone, Inc (4736 & 4942) setup a new delivery location for 4736",
  "description": "Due to an issue for client  Veritone, Inc (4736) and Oxford Road (4942), survey files were being placed in FTP folder 4942, and clearances are tied to the number. The client app server is tied to 4037. The discrepancy is causing an issue with FTP/FSX client configurations.\nClient ID 4736 is configured with an FSX folder: \\\\fse1a4000.gotostrata.com\\apppath4736B\\Strata\\Data\nDENT is not currently configured to deliver data associated with 4942 to an FSX location for 4736, as no corresponding FSX path/folder exists. This is cause data delays for the client.\nProposed Solution: \nCan this path be created for if in case future clearances are received under 4942 so that data can be delivered to 4736? in DENT, create a new Delivery Location for Client ID=4942 that points to: \\\\fdse1a4000.gotostrata.com\\apppath4736B\\Strata\\Data\n- Please see Zendesk Support tab for further comments and attachments.",
  "acceptance_criteria": "",
  "comments": [
    {
      "created": "2026-08-14T07:06:04.634+0000",
      "author": "Ramalingam, Sumi",
      "body": "Hi  , can you check the path is available? if not, please create a new Delivery Location for Client ID=4942 that points to: \\\\fdse1a4000.gotostrata.com\\apppath4736B\\Strata\\Data \nCC:"
    },
    {
      "created": "2026-08-25T16:44:39.849+0000",
      "author": "Worth, Shukura",
      "body": "I can confirm that the various types of Arbitron data have been placed in the updated (correct) fsx path: \nWe will close this ticket."
    }
  ]
}`

	tmpDir := t.TempDir()
	ticketFile := filepath.Join(tmpDir, "FWS-67484.json")
	if err := os.WriteFile(ticketFile, []byte(sampleJSON), 0644); err != nil {
		t.Fatalf("Failed to write test ticket: %v", err)
	}

	loadedTicket, err := Load(ticketFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedTicket.TicketID != "FWS-67484" {
		t.Errorf("Expected TicketID 'FWS-67484', got '%s'", loadedTicket.TicketID)
	}
	if loadedTicket.Summary != "DENT update: Veritone, Inc (4736 & 4942) setup a new delivery location for 4736" {
		t.Errorf("Expected Summary match, got '%s'", loadedTicket.Summary)
	}
	if len(loadedTicket.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(loadedTicket.Comments))
	}
	if loadedTicket.Comments[0].Author != "Ramalingam, Sumi" {
		t.Errorf("Expected first comment author 'Ramalingam, Sumi', got '%s'", loadedTicket.Comments[0].Author)
	}
}
