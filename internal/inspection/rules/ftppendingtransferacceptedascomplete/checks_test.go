package ftppendingtransferacceptedascomplete

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FtpPendingTransferAcceptedAsComplete"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }\n", 1},
		{"\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY) === FTP_FINISHED) { return true; }\n", 0},
		{"if(ftp_nb_continue($f)){echo 'complete';}", 1},
		{"if(!ftp_nb_continue($f)){return false;}", 0},
		{"if(ftp_nb_continue($f)){echo $text;}", 0},
		{"if(ftp_nb_continue($f)===FTP_MOREDATA){return true;}", 0},
		{"if(false){\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }\n}", 0},
		{"unknown();while(ftp_nb_continue($f)){break;}", 0},
		{"if(ftp_nb_continue($f)){$g=function(){return true;};}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
