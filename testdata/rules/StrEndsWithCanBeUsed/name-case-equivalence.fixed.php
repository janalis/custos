<?php
function endsWithSuffix($o, $s) {
    return substr($o->Path(), -strlen($s)) === $s;
}
function endsWith($o) {
    return substr($o->Path(), -strlen($o->suffix())) === $o->Suffix();
}
