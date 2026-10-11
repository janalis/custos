<?php
function receivedObject(SplFileObject $file) {
    return <warning descr="Pass the CSV escape character explicitly.">$file->fgetcsv()</warning>;
}

function sharedObject(array $row) {
    $file = new SplFileObject('php://memory', 'w+');
    configure($file);
    return <warning descr="Pass the CSV escape character explicitly.">$file->fputcsv($row)</warning>;
}

function maybeConfigured(bool $strict) {
    $file = new SplTempFileObject();
    if ($strict) {
        $file->setCsvControl(',', '"', '');
    }
    return <warning descr="Pass the CSV escape character explicitly.">$file->fgetcsv()</warning>;
}
