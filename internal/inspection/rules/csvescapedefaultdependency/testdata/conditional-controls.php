<?php
function conditionalDefaultReset(bool $reset) {
    $file = new SplTempFileObject();
    $file->setCsvControl(',', '"', '');
    if ($reset) {
        <warning descr="Pass the CSV escape character explicitly.">$file->setCsvControl(',', '"')</warning>;
    }
    return <warning descr="Pass the CSV escape character explicitly.">$file->fgetcsv()</warning>;
}
function explicitOnBothPaths(bool $reset) {
    $file = new SplTempFileObject();
    $file->setCsvControl(',', '"', '');
    if ($reset) { $file->setCsvControl(',', '"', '~'); }
    return $file->fgetcsv();
}
function forwardedArguments(array $arguments) {
    return str_getcsv(...$arguments);
}
