<?php
$input = trim(fgets(STDIN));
$ready = <weak_warning descr="Compare with an empty string instead: '$input !== '''.">strlen($input) > 0</weak_warning>;
if (!mb_strlen($argv[1])) {
    exit(1);
}
$len = strlen($input);
