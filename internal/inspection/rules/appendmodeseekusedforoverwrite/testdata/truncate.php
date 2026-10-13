<?php
function keepHead($path)
{
    $fp = fopen($path, 'a+');
    if ($fp === false) {
        return;
    }
    ftruncate($fp, 16);
    if (fseek($fp, 0) !== 0) {
        return;
    }
    <warning descr="Use an overwrite-capable stream mode.">fwrite($fp, "head")</warning>;
}

function rewriteHead($path)
{
    $fp = fopen($path, 'a+');
    if ($fp === false) {
        return;
    }
    ftruncate($fp, 0);
    fwrite($fp, "body\n");
    if (!rewind($fp)) {
        return;
    }
    <warning descr="Use an overwrite-capable stream mode.">fputs($fp, "head")</warning>;
}
