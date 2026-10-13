<?php
// Script-level guard clause.
if ($_FILES['f']['type'] !== 'application/pdf') { exit('rejected'); }
<warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
