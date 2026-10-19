<?php
// A jump can land between the length test and the decode: the test proves nothing.
function jumpPastCheck($in) {
    $raw = fread($in, 4);
    goto decode;
    if (strlen($raw) !== 4) { return null; }
    decode:
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}
