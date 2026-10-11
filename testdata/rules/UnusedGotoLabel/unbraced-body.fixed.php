<?php
function unbraced_bodies($count)
{
    if ($count) ; echo "first\n";
    while ($count-- > 0) ; echo "second\n";
    if ($count) echo "third\n"; else ; echo "fourth\n";
    foreach ([1, 2] as $item) ; echo "fifth\n";
    switch ($count) {
        case 1:
            echo "sixth\n";
    }
    echo "end\n";
}
