<?php
class Item { public readonly int $id; public function __construct() {$this->id=1; <error descr="Initialize this readonly property only once.">$this->id</error>=2;} }
class Counter { public readonly int $n; public function __construct(int $n) {$this->n=$n; <error descr="Initialize this readonly property only once.">$this->n</error>+=2;} }
class Label { public readonly string $text; public function __construct() {$this->text='a'; <error descr="Initialize this readonly property only once.">$this->text</error>.='b';} }
