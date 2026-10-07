<?php
class Data { public static array $rows = []; }
function dump() {
    foreach (data::$rows as $iValue) {
        echo $iValue;
    }
}
