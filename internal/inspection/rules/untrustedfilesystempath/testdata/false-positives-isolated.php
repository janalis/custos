<?php
$name = $_GET['file']; if (!in_array($name, ['guide.txt', 'terms.txt'], true)) { exit; } readfile('/srv/downloads/' . $name);
