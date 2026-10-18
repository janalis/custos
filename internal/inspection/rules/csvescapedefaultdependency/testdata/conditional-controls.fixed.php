<?php
function conditionalDefaultReset(bool $reset) {
    $file = new SplTempFileObject();
    $file->setCsvControl(',', '"', '');
    if ($reset) {
        $file->setCsvControl(',', '"', escape: "\\");
    }
    return $file->fgetcsv();
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
