<?php
class Item { public readonly int $id; public function __construct() {$this->id=1; <error descr="Initialize this readonly property only once.">$this->id</error>=2;} }
