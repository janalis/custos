<?php
$input = trim(fgets(STDIN));
$ready = <weak_warning descr="Compare with an empty string instead: '$input !== '''.">strlen($input) > 0</weak_warning>;
if (<weak_warning descr="Compare with an empty string instead: '(string)$argv[1] === '''.">!mb_strlen($argv[1])</weak_warning>) {
    exit(1);
}
$len = strlen($input);
