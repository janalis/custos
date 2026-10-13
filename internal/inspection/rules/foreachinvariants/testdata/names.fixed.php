<?php
function show(array $items) {
    foreach ($items as $iValue) {
        echo $iValue;
    }
    foreach ($items as $v) {
        echo $v;
    }
}
