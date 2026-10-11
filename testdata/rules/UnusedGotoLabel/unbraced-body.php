<?php
function unbraced_bodies($count)
{
    if ($count) <weak_warning descr="Label 'skip' is never targeted by a goto; remove it.">skip:</weak_warning> echo "first\n";
    while ($count-- > 0) <weak_warning descr="Label 'spin' is never targeted by a goto; remove it.">spin:</weak_warning> echo "second\n";
    if ($count) echo "third\n"; else <weak_warning descr="Label 'other' is never targeted by a goto; remove it.">other:</weak_warning> echo "fourth\n";
    foreach ([1, 2] as $item) <weak_warning descr="Label 'each' is never targeted by a goto; remove it.">each:</weak_warning> echo "fifth\n";
    switch ($count) {
        case 1:
            <weak_warning descr="Label 'inCase' is never targeted by a goto; remove it.">inCase:</weak_warning>
            echo "sixth\n";
    }
    echo "end\n";
}
