<?php
function namedStreamFactory() { return tmpfile(); }
function namedStreamPass($stream) { return $stream; }
function namedResourceCaller() {
    $first = namedStreamFactory();
    $second = namedStreamFactory();
    if ($first === false || $second === false) { return; }
    fclose($first);
    fread($second, 1);
    <error descr="Use an open stream for this operation.">fread($first, 1)</error>;

}

function namedPassCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    fclose($stream);
    $alias = namedStreamPass($stream);
    <error descr="Use an open stream for this operation.">fread($alias, 1)</error>;
}
