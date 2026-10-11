<?php
<error descr="Request a count within the array size.">array_rand(['oak', 'elm'], 4)</error>;

function pickNone(array $rows)
{
    $none = [];
    $seen = 0;
    foreach ($rows as $row) {
        $seen++;
    }
    return <error descr="Request a count within the array size.">array_rand($none)</error>;
}

$empty = [];
<error descr="Request a count within the array size.">array_rand($empty)</error>;
