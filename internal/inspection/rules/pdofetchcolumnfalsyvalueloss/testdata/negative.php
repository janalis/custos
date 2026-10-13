<?php
function good(PDOStatement $s) { if (($value = $s->fetchColumn()) !== false) { echo $value; } }
