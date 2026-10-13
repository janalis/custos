<?php
class LocalStatement { public function fetchColumn() { return 1; } } function benign(LocalStatement $s) { if ($v = $s->fetchColumn()) { echo $v; } }
