<?php
function interpolation(array $a, array $o) {
    foreach ($a as $i => $iValue) {
        echo "{$iValue}abc $iValue $iValue {$iValue}_x {$iValue}[0] {$iValue}->p $iValue-x é{$iValue}é";
        $a[$i]++;
    }
}
