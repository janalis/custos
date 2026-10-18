<?php
namespace First {
    function convert($first, $second) {}
    <error descr="Match the callback parameters to the supplied arrays.">array_map('First\\convert', [1])</error>;
}
namespace Second {
    function convert($first) {}
    array_map('Second\\convert', [1]);
}
