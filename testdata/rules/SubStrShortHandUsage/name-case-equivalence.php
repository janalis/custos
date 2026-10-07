<?php
function tail($o, $from) {
    return substr($o->Text(), $from, <warning descr="The length 'strlen($o-&gt;text()) - $from' is unnecessary; remove it.">strlen($o->text()) - $from</warning>);
}
