<?php
class Data { public static array $rows = []; }
function dump() {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count(Data::$rows); $i++) {
        echo data::$rows[$i];
    }
}
