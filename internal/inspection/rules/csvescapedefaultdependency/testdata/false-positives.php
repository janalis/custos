<?php
function configuredEscape(array $row) {
    $file = new SplFileObject('php://memory', 'w+');
    $file->setCsvControl(',', '"', '');
    $file->fputcsv($row, ';');
    $file->rewind();
    return [$file->fgetcsv(';'), $file->fgetcsv(enclosure: "'")];
}
