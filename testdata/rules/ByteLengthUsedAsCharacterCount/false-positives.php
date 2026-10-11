<?php
// Byte offsets and frame lengths handed to byte-oriented builtins.
function stripCurrency($price) {
    return substr($price, strlen("€"));
}
function frame() {
    return pack('N', strlen("héllo")) . "héllo";
}
function tail($s) {
    return substr($s, -strlen('é'));
}
function skip($handle) {
    return fseek($handle, (strlen('é') + 4) * 2);
}
function pad($s) {
    return \str_pad($s, (strlen('ñ')) - 1);
}
