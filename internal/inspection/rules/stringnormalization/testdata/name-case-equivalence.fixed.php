<?php
class Sep { public static function chars() { return '-'; } }
function title($s) {
    return ucwords($s, SEP::Chars());
}
