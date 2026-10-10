<?php
class LocalStatement { public function execute($params) {} } function benign(LocalStatement $s) { $s->execute(['wrong' => 1]); }
