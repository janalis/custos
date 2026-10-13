<?php
namespace Local; function fopen($path, $mode) { return $path; } function fread($h, $n) {} $h = fopen('path', 'rb'); fread($h, 4);
