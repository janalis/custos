<?php
function probe($n, $m, $b, $flag) {
    $r = $n == $m ? true : 0;
    $r = in_array($n, $b) ? false : true;
    $r = $n ?: false;
    $r = $flag ? true : false;
    $r = !$flag ? false : true;
    $r = $n > $m ? null : false;
    $r = $n > $m ? 'yes' : 'no';
    return $r;
}
