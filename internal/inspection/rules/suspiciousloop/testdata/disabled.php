<?php
function scan(array $rows) {
    foreach ($rows as $rows) {}
    <error descr="Only the last expression of the 'for' condition is evaluated as the condition; combine them with && or ||.">for</error> ($i = 0; $i < 2, $i < 3; $i++) {}
}
