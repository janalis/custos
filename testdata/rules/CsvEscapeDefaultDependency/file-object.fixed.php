<?php
function localDefault(array $row) {
    $file = new SplFileObject('php://memory', 'w+');
    $file->fwrite("a,b\n");
    $file->setCsvControl(';', escape: "\\");
    $file->fputcsv($row, escape: "\\");
    return $file->fgetcsv(';', escape: "\\");
}
