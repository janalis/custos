<?php
class Sep { public static function chars() { return '-'; } }
function title($s) {
    return ucwords(<weak_warning descr="The inner 'ucwords(...)' call has no effect here.">ucwords($s, Sep::chars())</weak_warning>, SEP::Chars());
}
