<?php
function localDefault(array $row) {
    $file = new SplFileObject('php://memory', 'w+');
    $file->fwrite("a,b\n");
    <warning descr="Pass the CSV escape character explicitly.">$file->setCsvControl(';')</warning>;
    <warning descr="Pass the CSV escape character explicitly.">$file->fputcsv($row)</warning>;
    return <warning descr="Pass the CSV escape character explicitly.">$file->fgetcsv(';')</warning>;
}
