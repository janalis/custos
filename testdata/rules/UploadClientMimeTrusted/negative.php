<?php
if ((new finfo(FILEINFO_MIME_TYPE))->file($_FILES['file']['tmp_name']) === 'image/png') { move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png'); }
