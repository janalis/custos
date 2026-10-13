<?php
function review($f, $h) {
    <warning descr="Output the stream without printing its byte count.">echo(readfile($f));</warning>
    <warning descr="Output the stream without printing its byte count.">print(readfile($f));</warning>
    <warning descr="Output the stream without printing its byte count.">echo ((fpassthru($h)));</warning>
    echo (readfile($f)), "count";
    echo (readfile($f)) . "\n";
    print(readfile($f)) or exit(1);
    $r = print(readfile($f));
}
