<?php
// Constant names are case-sensitive even when spelled like a keyword.
class Limits { const Array = null; const ARRAY = 6; const Default = null; const DEFAULT = 'D'; }
$a = Limits::ARRAY !== null ? Limits::Array : 1;
$b = null === Limits::DEFAULT ? 'x' : Limits::Default;
function pick($in) {
    if (isset($in[Limits::ARRAY])) {
        return $in[Limits::Array];
    }
    return null;
}
