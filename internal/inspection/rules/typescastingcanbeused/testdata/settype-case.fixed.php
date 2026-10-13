<?php
function normalise($qty, $tags, $flag, $ratio) {
    $qty = (int)$qty;
    $tags = (array)$tags;
    $flag = (bool)$flag;
    settype($ratio, 'NULL');
    settype($ratio, 'Object');
}
