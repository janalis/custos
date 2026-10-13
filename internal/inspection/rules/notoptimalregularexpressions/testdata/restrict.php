<?php
function restrictDemo($w, $m) {
    preg_match('/kelvin/iur', $w, $m);
    preg_match('/(k)\d/n', $w, $m);
    preg_match(<error descr="The /r flag needs /u to take effect."><error descr="The /r flag needs /i to take effect.">'/kelvin/r'</error></error>, $w, $m);
    preg_match(<error descr="The /r flag needs /i to take effect.">'/kelvin/ur'</error>, $w, $m);
}
