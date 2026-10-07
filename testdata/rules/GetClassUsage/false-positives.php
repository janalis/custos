<?php
function plain(string $s, $unknown, ?Order $o = null) {
    get_class($s);
    get_class($unknown);
    get_class($this);
    get_class(new stdClass());
    if (isset($o)) {}
    get_class($o);
}
