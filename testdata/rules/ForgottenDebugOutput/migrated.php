<?php
function f($x) {
    var_dump($x);
    <error descr="Debug output call; remove it if it was left over from debugging.">my_dump($x)</error>;
}
