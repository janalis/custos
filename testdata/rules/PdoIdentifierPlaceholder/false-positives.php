<?php
class LocalConnection { public function prepare($sql) {} } function benign(LocalConnection $c) { $c->prepare('SELECT * FROM :table'); }
