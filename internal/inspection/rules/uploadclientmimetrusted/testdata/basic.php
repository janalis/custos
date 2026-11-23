<?php
if ($_FILES['file']['type'] === 'image/png') { <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png')</warning>; }
// A closure over finfo::file() inspects nothing.
function closureOnly() {
    $inspect = (new finfo(FILEINFO_MIME_TYPE))->file(...);
    if ($_FILES['file']['type'] === 'image/png' && (new finfo(FILEINFO_MIME_TYPE))->file(...)) { <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png')</warning>; }
}
