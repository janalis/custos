package uploadclientmimetrusted

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestIndependentNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"if($_FILES['f']['type']==='image/png' && (new finfo(FILEINFO_MIME_TYPE))->file($_FILES['f']['tmp_name'])==='image/png'){move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 0},
		{"if($_FILES['f']['type']==='image/png'){if((new finfo(FILEINFO_MIME_TYPE))->file($_FILES['f']['tmp_name'])==='image/png'){move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}}", 0},
		{"if($_FILES['f']['type']==='image/png'){if(getimagesize($_FILES['f']['tmp_name'])===false){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 0},
		{"if($_FILES['f']['type']==='image/png'){if(getimagesize($_FILES['g']['tmp_name'])===false){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 1},
		{"if($_FILES['f']['type']!=='image/png'){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');", 1},
		{"if($_FILES['f']['type']!=='image/png'){return;}if(mime_content_type($_FILES['f']['tmp_name'])!=='image/png'){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');", 0},
		{"if($_FILES['g']['type']!=='image/png'){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');", 0},
		{"if($_FILES['f']['type']==='image/png'){return;}move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');", 0},
		{"if(in_array($_FILES['f']['type'],['image/png','image/jpeg'],true)){move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 1},
		{"namespace App; function in_array($a,$b){return true;} if(in_array($_FILES['f']['type'],['image/png'])){move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 0},
		{"if($_FILES['f']['type']==='image/png' && $_FILES['f']['size']<10){move_uploaded_file($_FILES['f']['tmp_name'],'/tmp/image');}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UploadClientMimeTrusted"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
