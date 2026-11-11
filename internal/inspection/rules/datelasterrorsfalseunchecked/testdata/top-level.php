<?php
// A parse failed in the file's top-level code.
$date = date_create($argv[1] ?? '');
if ($date === false) {
    $errors = DateTime::getLastErrors();
    echo $errors['error_count'];
}
