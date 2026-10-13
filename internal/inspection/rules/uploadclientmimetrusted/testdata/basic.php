<?php
if ($_FILES['file']['type'] === 'image/png') { <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png')</warning>; }
