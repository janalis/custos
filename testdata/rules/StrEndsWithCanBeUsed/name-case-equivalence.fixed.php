<?php
function endsWithSuffix($o, $s) {
    return str_ends_with($o->Path(), $s);
}
function endsWith($o) {
    return str_ends_with($o->Path(), $o->Suffix());
}
