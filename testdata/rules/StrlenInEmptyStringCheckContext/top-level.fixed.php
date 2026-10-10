<?php
$input = trim(fgets(STDIN));
$ready = $input !== '';
if (!mb_strlen($argv[1])) {
    exit(1);
}
$len = strlen($input);
