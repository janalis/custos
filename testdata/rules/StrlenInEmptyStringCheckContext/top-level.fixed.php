<?php
$input = trim(fgets(STDIN));
$ready = $input !== '';
if ((string)$argv[1] === '') {
    exit(1);
}
$len = strlen($input);
